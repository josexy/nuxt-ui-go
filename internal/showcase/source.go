package showcase

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
)

var imports = map[string]string{
	"fmt": "fmt", "time": "time", "strings": "strings", "slices": "slices", "cmp": "cmp", "context": "context", "log": "log",
	"base64": "encoding/base64", "errors": "errors",
	"mygo": "github.com/egoist/mygo", "ui": "github.com/egoist/mygo/ui",
	"nui": "github.com/josexy/nuxt-ui-go", "lucide": "github.com/josexy/nuxt-ui-go/icons/lucide",
	"showcase": "github.com/josexy/nuxt-ui-go/internal/showcase",
}

// FormatSource resolves imports from the example's own Go syntax.
func FormatSource(pkg, body string) (string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "example.go", "package "+pkg+"\n"+body, 0)
	if err != nil {
		return "", err
	}
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if selector, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := selector.X.(*ast.Ident); ok && imports[ident.Name] != "" {
				used[ident.Name] = true
			}
		}
		return true
	})
	var names []string
	for name := range used {
		names = append(names, name)
	}
	slices.SortFunc(names, func(a, b string) int {
		aExternal, bExternal := strings.Contains(imports[a], "."), strings.Contains(imports[b], ".")
		if aExternal != bExternal {
			if aExternal {
				return 1
			}
			return -1
		}
		return strings.Compare(imports[a], imports[b])
	})
	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n\nimport (\n", pkg)
	previousExternal := false
	for _, name := range names {
		external := strings.Contains(imports[name], ".")
		if external && !previousExternal {
			out.WriteByte('\n')
		}
		previousExternal = external
		if name == "nui" {
			fmt.Fprintf(&out, "nui %q\n", imports[name])
		} else {
			fmt.Fprintf(&out, "%q\n", imports[name])
		}
	}
	out.WriteString(")\n\n" + body)
	formatted, err := format.Source(out.Bytes())
	return string(formatted), err
}

func (d Demo) Source(input Config) (string, string, error) {
	cfg, err := d.Normalize(input)
	if err != nil {
		return "", "", err
	}
	values := &valueSource{}
	body, err := d.configuredBody(cfg, false, values)
	if err != nil {
		return "", "", err
	}
	appConfig := "config"
	if patch, ok := cfg.Options["AppConfig"]; ok {
		for _, b := range d.Bindings() {
			if b.ID == "AppConfig" {
				appConfig, err = overlayExpression(b, "config", patch, values)
				if err != nil {
					return "", "", err
				}
			}
		}
	}
	var state strings.Builder
	state.WriteString(d.State)
	configuredSetup, err := d.configuredSetup(cfg, false, values)
	if err != nil {
		return "", "", err
	}
	var setup strings.Builder
	for _, path := range sortedPointerKeys(values.pointers) {
		p := values.pointers[path]
		state.WriteString("\n" + p.name + " " + p.typ)
		setup.WriteString("\ns." + p.name + " = " + p.init)
	}
	setup.WriteString("\n" + configuredSetup)
	windowAssignment := ""
	if values.window {
		state.WriteString("\nWindow *mygo.Window")
		windowAssignment = "s.Window = "
	}
	if values.focus {
		state.WriteString("\nReturnFocus ui.Handle")
		body = "nui.Button(c.Key(\"preview-return-focus\"), t, \"Focus target\", nui.ButtonOptions{Variant: nui.Outline}).AlignSelf(ui.Start).Bind(&s.ReturnFocus)\n" + body
	}
	formattedBody, err := format.Source([]byte(body))
	if err != nil {
		return "", "", err
	}
	body = string(formattedBody)
	if cleanup := d.disposeBody(); cleanup != "" {
		setup.WriteString("\ndefer " + cleanup)
	}
	mode := map[string]string{"light": "LightMode", "dark": "DarkMode", "system": "SystemMode"}[cfg.Mode]
	primaryPalette := fmt.Sprintf("palettes.Primary, _ = nui.PaletteByName(%q)", cfg.Primary)
	primaryTheme := ""
	if cfg.Primary == "black" {
		primaryPalette = ""
		primaryTheme = ".WithBlackPrimary()"
	}
	code := fmt.Sprintf(`type example struct { %s }

func (s *example) view(c *ui.Context, t nui.Theme) {
%s
}

func main() {
 s := &example{}
 %s
 palettes := nui.DefaultPalettes()
 %s
 palettes.Neutral, _ = nui.PaletteByName(%s)
 config := nui.AppConfig{Mode: nui.%s, Light: nui.NewTheme(false, palettes)%s, Dark: nui.NewTheme(true, palettes)%s}
 mygo.App.SetLanguages(nui.SupportedLocales()...)
 mygo.App.WhenReady(func() {
  mygo.App.SetLocale(%s)
  %smygo.NewWindow(mygo.WindowOptions{Title: %s, Width: 760, Height: 560, Content: ui.View(func(c *ui.Context) {
   t := nui.App(c, %s)
   ui.Column(c).Fill().Padding(24).Gap(16).Children(func() { s.view(c, t) })
  })})
 })
 if err := mygo.App.Run(); err != nil { log.Fatal(err) }
}
`, state.String(), body, setup.String(), primaryPalette, strconv.Quote(cfg.Neutral), mode, primaryTheme, primaryTheme, strconv.Quote(cfg.Locale), windowAssignment, strconv.Quote(d.Name+" · NuxtUI Go"), appConfig)
	full, err := FormatSource("main", code)
	if err != nil {
		return "", "", err
	}
	fragment := "// Inside a view; c is *ui.Context, t is nui.Theme.\n"
	if state.String() != "" {
		fragment += "// Keep s (example state) outside the view. See the complete example.\n"
	}
	return fragment + body, full, nil
}

func (d Demo) disposeBody() string {
	switch d.ID {
	case "form":
		return "s.Form.Cancel()"
	case "command-palette":
		return "s.Commands.Close()"
	case "tree":
		return "s.Tree.Close()"
	case "file-upload":
		return "s.Files.Dispose()"
	case "select-menu":
		return "s.Results.Cancel()"
	default:
		return ""
	}
}

// PreviewSource compiles the exact displayed templates into native/WASM views.
func PreviewSource() (string, error) {
	var body, cases strings.Builder
	body.WriteString("type Example interface { View(*ui.Context, nui.Theme, showcase.Config); Configure(showcase.Config) error; Dispose(); SetReport(func(string, []string)); SetWindow(*mygo.Window) }\n")
	for _, demo := range Catalog() {
		name := strings.ReplaceAll(demo.ID, "-", "_") + "Example"
		view, err := demo.configuredBody(Config{}, true, nil)
		if err != nil {
			return "", err
		}
		if demo.ID == "tree" {
			view = "s.Live.Tree = &s.Tree\n" + view
		}
		if demo.ID == "command-palette" {
			view = "s.Live.Commands = &s.Commands\n" + view
		}
		fmt.Fprintf(&body, "type %s struct { Live showcase.Runtime; %s }\nfunc(s *%s) View(c *ui.Context,t nui.Theme,cfg showcase.Config){\n%s\n}\n", name, demo.State, name, view)
		fmt.Fprintf(&body, "func(s *%s) SetReport(fn func(string, []string)){ s.Live.SetReport(fn) }\n", name)
		fmt.Fprintf(&body, "func(s *%s) SetWindow(window *mygo.Window){ s.Live.Window = window }\n", name)
		configure, err := demo.configureMethod(name)
		if err != nil {
			return "", err
		}
		body.WriteString(configure)
		fmt.Fprintf(&body, "func(s *%s) Dispose(){%s}\n", name, demo.disposeBody())
		fmt.Fprintf(&cases, "case %q: s:= &%s{}; %s; return s,nil\n", demo.ID, name, demo.Setup)
	}
	fmt.Fprintf(&body, "func New(id string)(Example,error){ switch id { %s }; return nil,fmt.Errorf(\"unknown example %%s\",id) }", cases.String())
	source, err := FormatSource("preview", body.String())
	return "// Code generated by go run ./internal/cmd/genwebsite; DO NOT EDIT.\n\n" + source, err
}

func sortedPointerKeys(pointers map[string]sourcePointer) []string {
	keys := make([]string, 0, len(pointers))
	for key := range pointers {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
