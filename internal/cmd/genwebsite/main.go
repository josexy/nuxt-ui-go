// Genwebsite extracts public Go APIs and builds the website's example sources.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

type field struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Comment     string        `json:"comment"`
	Description showcase.Text `json:"description"`
	Zero        string        `json:"zero"`
}
type api struct {
	Name      string  `json:"name"`
	Signature string  `json:"signature"`
	Fields    []field `json:"fields"`
	File      string  `json:"file"`
	Line      int     `json:"line"`
}
type page struct {
	showcase.Demo
	API      []api              `json:"api"`
	Snippet  string             `json:"snippet"`
	Source   string             `json:"source"`
	Bindings []showcase.Binding `json:"bindings"`
}
type extractor struct {
	set          *token.FileSet
	types        map[string]ast.Expr
	declarations map[string]api
	notes        map[string]showcase.Text
}

func (e *extractor) print(n any) string {
	var out bytes.Buffer
	if err := format.Node(&out, e.set, n); err != nil {
		log.Fatal(err)
	}
	return out.String()
}
func (e *extractor) fields(expr ast.Expr, seen map[string]bool) []field {
	switch v := expr.(type) {
	case *ast.Ident:
		if seen[v.Name] {
			return nil
		}
		seen[v.Name] = true
		return e.fields(e.types[v.Name], seen)
	case *ast.StructType:
		var result []field
		for _, f := range v.Fields.List {
			if len(f.Names) == 0 {
				result = append(result, e.fields(f.Type, seen)...)
				continue
			}
			for _, name := range f.Names {
				if !name.IsExported() {
					continue
				}
				comment := strings.TrimSpace(f.Doc.Text() + " " + f.Comment.Text())
				if len(f.Names) > 1 && strings.HasPrefix(comment, f.Names[0].Name+" ") && name.Name != f.Names[0].Name {
					comment = ""
				}
				typeName := e.print(f.Type)
				zero := "zero value"
				switch {
				case typeName == "bool":
					zero = "false"
				case typeName == "string":
					zero = `""`
				case strings.HasPrefix(typeName, "Optional["):
					zero = "unset"
				case strings.HasPrefix(typeName, "*") || strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "map[") || strings.HasPrefix(typeName, "func("):
					zero = "nil"
				case typeName == "Size":
					zero = "SizeDefault"
				case typeName == "Color":
					zero = "ColorDefault"
				case typeName == "Variant":
					zero = "VariantDefault"
				case strings.HasPrefix(typeName, "int") || strings.HasPrefix(typeName, "float") || typeName == "time.Duration":
					zero = "0"
				}
				result = append(result, field{Name: name.Name, Type: typeName, Comment: comment, Description: e.notes[name.Name], Zero: zero})
			}
		}
		return result
	}
	return nil
}
func extract(root string) *extractor {
	e := &extractor{set: token.NewFileSet(), types: map[string]ast.Expr{}, declarations: map[string]api{}, notes: showcase.APINotes()}
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(e.set, path, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		ast.FileExports(f)
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil || !d.Name.IsExported() {
					continue
				}
				// Function signatures only, never function bodies in the API table.
				copyDecl := *d
				copyDecl.Body = nil
				copyDecl.Doc = nil
				e.declarations[d.Name.Name] = api{Name: d.Name.Name, Signature: e.print(&copyDecl), File: filepath.Base(path), Line: e.set.Position(d.Pos()).Line}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if t, ok := spec.(*ast.TypeSpec); ok && t.Name.IsExported() {
						e.types[t.Name.Name] = t.Type
						e.declarations[t.Name.Name] = api{Name: t.Name.Name, Signature: "type " + e.print(t), File: filepath.Base(path), Line: e.set.Position(t.Pos()).Line}
					}
				}
			}
		}
	}
	return e
}
func write(path string, data []byte) {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatal(err)
	}
}

func main() {
	root := flag.String("root", ".", "repository root")
	snippets := flag.String("snippets", "", "also write standalone examples into this directory")
	flag.Parse()
	invalid := false
	for _, demo := range showcase.Catalog() {
		if _, _, err := demo.Source(showcase.Config{}); err != nil {
			log.Printf("%s: %v", demo.ID, err)
			invalid = true
		}
	}
	if invalid {
		log.Fatal("invalid example templates")
	}
	e := extract(*root)
	var pages []page
	for _, demo := range showcase.Catalog() {
		cfg, err := demo.Normalize(showcase.Config{})
		if err != nil {
			log.Fatal(demo.ID, ": ", err)
		}
		snippet, source, err := demo.Source(cfg)
		if err != nil {
			log.Fatal(demo.ID, ": ", err)
		}
		p := page{Demo: demo, Snippet: snippet, Source: source, Bindings: demo.Bindings()}
		for _, name := range append(slices.Clone(demo.Symbols), demo.Types...) {
			a, ok := e.declarations[name]
			if !ok {
				log.Fatalf("%s: missing public API %s", demo.ID, name)
			}
			a.Fields = e.fields(e.types[name], map[string]bool{})
			for i := range a.Fields {
				if description, ok := e.notes[name+"."+a.Fields[i].Name]; ok {
					a.Fields[i].Description = description
				}
			}
			p.API = append(p.API, a)
		}
		pages = append(pages, p)
		if *snippets != "" {
			write(filepath.Join(*snippets, demo.ID, "main.go"), []byte(source))
		}
		for _, language := range []string{"zh", "en"} {
			description, guide, note := demo.Description.ZH, demo.Guide.ZH, demo.Note.ZH
			noteTitle := "参数与平台说明"
			if language == "en" {
				description, guide, note = demo.Description.EN, demo.Guide.EN, demo.Note.EN
				noteTitle = "Parameters and platform notes"
			}
			// JSON strings are valid YAML scalars and escape arbitrary punctuation.
			titleJSON, _ := json.Marshal(demo.Name)
			descJSON, _ := json.Marshal(description)
			content := fmt.Sprintf("---\ntitle: %s\ndescription: %s\ncomponent: %s\n---\n\n%s\n", titleJSON, descJSON, demo.ID, guide)
			if demo.ID == "file-upload" {
				content += fmt.Sprintf("\n### %s\n\n%s\n", noteTitle, note)
			}
			write(filepath.Join(*root, "website/content", language, "components", demo.ID+".md"), []byte(content))
		}
	}
	views, err := showcase.PreviewSource()
	if err != nil {
		log.Fatal(err)
	}
	write(filepath.Join(*root, "internal/showcase/preview/generated.go"), []byte(views))
	data, err := json.MarshalIndent(map[string]any{"components": pages, "schemas": showcase.Schemas(), "primary": nui.PrimaryPaletteNames(), "neutral": nui.NeutralPaletteNames()}, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	write(filepath.Join(*root, "website/app/data/catalog.json"), append(data, '\n'))
	fmt.Printf("Generated %d examples, %d bilingual pages and public API tables\n", len(pages), len(pages)*2)
}
