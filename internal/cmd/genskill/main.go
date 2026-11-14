// Genskill regenerates the component references of the nuxt-ui-go agent skill
// (.agents/skills/nuxt-ui-go/references/components/*.md) from the root
// package's Go API, internal/showcase/catalog.json and docs/*.md. The output
// is self-contained: it names no repository paths, so the skill folder can be
// copied into any project.
//
// Run it from the repository root after changing a component's API, docs or
// catalog entry:
//
//	CGO_ENABLED=0 go run ./internal/cmd/genskill
//	CGO_ENABLED=0 go run ./internal/cmd/genskill -check
//
// -check writes nothing and exits 1 when a reference is stale. The command
// prints exported API that no reference covers; assign it in the override
// tables below.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/josexy/nuxt-ui-go/internal/showcase"
)

const modulePath = "github.com/josexy/nuxt-ui-go"

// Override tables. Keys are catalog IDs.
var (
	// extraSymbols assigns exported functions that the catalog does not list.
	extraSymbols = map[string][]string{
		"app":        {"Window", "Themed", "SupportedLocales"},
		"theme":      {"DefaultPalettes", "PaletteByName", "PrimaryPaletteNames", "NeutralPaletteNames"},
		"kbd":        {"FormatKey"},
		"avatar":     {"Initials"},
		"input-menu": {"SearchSelect"},
	}
	// extraTypes assigns exported types that the catalog does not list.
	// Types owned by "theme" are linked rather than repeated elsewhere.
	extraTypes = map[string][]string{
		"command-palette": {"CommandIndex"},
		"listbox":         {"ListboxEvent"},
		"file-upload":     {"UploadStatus"},
		"theme": {"Color", "Size", "Variant", "ButtonVariant", "Orientation", "Slot", "Styles", "Component",
			"ShadowTokens", "Palette", "Palettes", "Mode", "MotionTokens"},
	}
	// matchNames replaces the names used to find documentation.
	matchNames = map[string][]string{
		"app":    {"App", "AppConfig", "DefaultApp"},
		"theme":  {"Theme", "NewTheme", "WithComponent", "Themed"},
		"header": {"HeaderLayout"},
		"footer": {"FooterLayout"},
		"error":  {"ErrorView"},
	}
	// extraSections includes "file.md#Heading" sections whose heading does not
	// name the component.
	extraSections = map[string][]string{
		"theme":  {"foundations.md#App and theme", "foundations.md#Style overrides"},
		"link":   {"foundations.md#Buttons and links"},
		"avatar": {"foundations.md#Images, avatars, and icons"},
		"icon":   {"foundations.md#Images, avatars, and icons"},
	}
	// wholeSectionsOnly disables paragraph and table-row excerpts for
	// components whose name appears in passing throughout docs/.
	wholeSectionsOnly = map[string]bool{"theme": true}
	// skipSections excludes "file.md#Heading" sections that name a component
	// only in passing.
	skipSections = map[string][]string{
		"switch": {"i18n.md#Switch languages"},
	}
)

// Stop words are catalog symbols that are too generic to find documentation.
var stopNames = map[string]bool{"Set": true, "Light": true, "Dark": true}

var categories = map[string]struct{ title, doc string }{
	"foundations": {"Themes and basic components", "foundations.md"},
	"inputs":      {"Inputs and selection", "inputs.md"},
	"overlays":    {"Overlays and navigation", "overlays-navigation.md"},
	"data":        {"Forms and data", "forms-data.md"},
	"advanced":    {"Advanced interaction and layout", "advanced.md"},
}

func main() {
	rootFlag := flag.String("root", ".", "repository root")
	skillFlag := flag.String("skill", filepath.Join(".agents", "skills", "nuxt-ui-go"), "skill directory, relative to the root")
	check := flag.Bool("check", false, "report stale references without writing")
	flag.Parse()
	root, skillDir, err := locate(*rootFlag, *skillFlag)
	if err != nil {
		fail(err)
	}
	g, err := newGenerator(root, skillDir)
	if err != nil {
		fail(err)
	}
	outDir := filepath.Join(skillDir, "references", "components")
	for _, d := range g.demos {
		g.docExcerpts(d) // records which reference contains each section
	}
	files := map[string][]byte{}
	for _, d := range g.demos {
		text, err := g.reference(d)
		if err != nil {
			fail(fmt.Errorf("%s: %w", d.ID, err))
		}
		files[d.ID+".md"] = []byte(text)
	}
	g.reportUncovered()
	if err := g.checkSkillIndex(); err != nil {
		if *check {
			fail(err)
		}
		fmt.Fprintln(os.Stderr, "warning:", err)
	}
	stale := syncFiles(outDir, files, *check)
	if *check && len(stale) > 0 {
		fmt.Fprintf(os.Stderr, "stale references (run go run ./internal/cmd/genskill):\n  %s\n", strings.Join(stale, "\n  "))
		os.Exit(1)
	}
	if !*check {
		fmt.Printf("wrote %d references to %s (%d changed)\n", len(files), rel(root, outDir), len(stale))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genskill:", err)
	os.Exit(1)
}

// locate checks that root is the nuxt-ui-go repository and resolves the skill.
func locate(root, skill string) (string, string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !bytes.HasPrefix(mod, []byte("module "+modulePath+"\n")) {
		return "", "", fmt.Errorf("%s is not the %s repository root", root, modulePath)
	}
	skillDir := filepath.Join(root, skill)
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		return "", "", err
	}
	return root, skillDir, nil
}

func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

// ---------------------------------------------------------------- generator

type generator struct {
	root    string
	skill   string
	fset    *token.FileSet
	pkg     *doc.Package
	funcs   map[string]*doc.Func // package-level functions
	types   map[string]*doc.Type
	typeFns map[string]string // constructor name -> type name
	demos   []showcase.Demo
	owner   map[string][]string // type name -> catalog IDs that list it
	allName map[string]string   // every match name -> catalog ID
	docs    []*mdFile
	consts  map[string]string // Component value -> constant name
	covered map[string]bool
	version string // "nuxt-ui-go v0.0.2 (MyGo v0.8.3)"
	// include maps "file.md#anchor" to the references that contain that
	// documentation section in full, for resolving cross-references.
	include map[string][]string
}

func newGenerator(root, skillDir string) (*generator, error) {
	g := &generator{root: root, skill: skillDir, fset: token.NewFileSet(), funcs: map[string]*doc.Func{}, types: map[string]*doc.Type{},
		typeFns: map[string]string{}, owner: map[string][]string{}, allName: map[string]string{}, consts: map[string]string{},
		covered: map[string]bool{}, demos: showcase.Catalog(), include: map[string][]string{}}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(g.fset, filepath.Join(root, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	g.pkg, err = doc.NewFromFiles(g.fset, files, modulePath)
	if err != nil {
		return nil, err
	}
	for _, f := range g.pkg.Funcs {
		g.funcs[f.Name] = f
	}
	for _, t := range g.pkg.Types {
		g.types[t.Name] = t
		for _, f := range t.Funcs {
			g.typeFns[f.Name] = t.Name
		}
		if t.Name == "Component" {
			for _, c := range t.Consts {
				for _, spec := range c.Decl.Specs {
					vs := spec.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if i < len(vs.Values) {
							if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
								v, _ := strconv.Unquote(lit.Value)
								g.consts[v] = n.Name
							}
						}
					}
				}
			}
		}
	}
	for _, d := range g.demos {
		for _, t := range g.demoTypes(d) {
			g.owner[t] = append(g.owner[t], d.ID)
		}
		for _, n := range g.names(d) {
			g.allName[n] = d.ID
		}
	}
	if g.version, err = libraryVersion(root); err != nil {
		return nil, err
	}
	docDir := filepath.Join(root, "docs")
	docEntries, err := os.ReadDir(docDir)
	if err != nil {
		return nil, err
	}
	for _, e := range docEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		// Historical investigation notes link to application checkouts and old
		// implementations; they are not part of the standalone usage reference.
		if e.Name() == "dip-line-rendering.md" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(docDir, e.Name()))
		if err != nil {
			return nil, err
		}
		g.docs = append(g.docs, parseMarkdown(e.Name(), string(data)))
	}
	return g, nil
}

func (g *generator) demoSymbols(d showcase.Demo) []string {
	return append(slices.Clone(d.Symbols), extraSymbols[d.ID]...)
}

func (g *generator) demoTypes(d showcase.Demo) []string {
	return append(slices.Clone(d.Types), extraTypes[d.ID]...)
}

func (g *generator) names(d showcase.Demo) []string {
	if n, ok := matchNames[d.ID]; ok {
		return n
	}
	var out []string
	seen := map[string]bool{}
	for _, n := range append([]string{d.Name}, d.Symbols...) {
		if !stopNames[n] && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// ---------------------------------------------------------------- reference

func (g *generator) reference(d showcase.Demo) (string, error) {
	var b strings.Builder
	cat := categories[d.Category]
	symbols := g.demoSymbols(d)
	title := d.Name
	if len(d.Symbols) > 0 && d.Symbols[0] != d.Name {
		title += " (`nui." + d.Symbols[0] + "`)"
	}
	fmt.Fprintf(&b, "<!-- Generated from the nuxt-ui-go source. DO NOT EDIT. -->\n\n")
	fmt.Fprintf(&b, "# %s\n\n%s\n\n", title, strings.TrimSpace(d.Description.EN))
	fmt.Fprintf(&b, "- Category: %s\n", cat.title)
	fmt.Fprintf(&b, "- Version: %s\n", g.version)
	if len(symbols) > 0 {
		fmt.Fprintf(&b, "- Functions: %s\n", codeList(symbols))
	}
	if types := g.demoTypes(d); len(types) > 0 {
		fmt.Fprintf(&b, "- Types: %s\n", codeList(types))
	}
	if c, ok := g.consts[d.ID]; ok {
		fmt.Fprintf(&b, "- Theme key: `nui.%s` (for `Theme.WithComponent`)\n", c)
	}
	b.WriteString("- Sections: Usage · Things to know · API · Documentation excerpts · Complete example\n\n")

	b.WriteString("## Usage\n\n")
	b.WriteString(demoteHeadings(filterProse(stripPreviewSteps(d.Guide.EN), previewOnly), 1))
	b.WriteString("\n\n")
	if note := filterProse(strings.TrimSpace(d.Note.EN), previewOnly); note != "" {
		b.WriteString("## Things to know\n\n" + note + "\n\n")
	}

	api, linked := g.apiSection(d)
	b.WriteString("## API\n\n" + api)
	if len(linked) > 0 {
		b.WriteString("Types documented in other references: ")
		var parts []string
		for _, t := range sortedKeys(linked) {
			parts = append(parts, fmt.Sprintf("`%s` → [%s.md](%s.md)", t, linked[t], linked[t]))
		}
		b.WriteString(strings.Join(parts, ", ") + ".\n\n")
	}

	if excerpts := g.docExcerpts(d); excerpts != "" {
		b.WriteString("## Documentation excerpts\n\nFrom the library documentation. Links point to other component references in this folder.\n\n" + excerpts)
	}

	example, err := g.example(d)
	if err != nil {
		return "", err
	}
	b.WriteString("## Complete example\n\nA runnable program. App state lives in `example`, outside the view, and the view only builds UI.\n\n```go\n" + example + "```\n\n")

	return strings.TrimRight(b.String(), "\n") + "\n", nil
}

func codeList(items []string) string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = "`" + s + "`"
	}
	return strings.Join(out, ", ")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// stripPreviewSteps drops the website-only "Try the preview" walkthrough.
func stripPreviewSteps(guide string) string {
	var out []string
	skip := false
	for line := range strings.SplitSeq(strings.TrimSpace(guide), "\n") {
		if strings.HasPrefix(line, "### ") {
			skip = strings.Contains(strings.ToLower(line), "try the preview")
		}
		if !skip {
			out = append(out, line)
		}
	}
	text := strings.TrimSpace(strings.Join(out, "\n"))
	return strings.TrimSpace(strings.TrimPrefix(text, "### Basic usage"))
}

// demoteHeadings adds levels to every Markdown heading outside code fences.
func demoteHeadings(text string, by int) string {
	lines := strings.Split(text, "\n")
	fence := false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			fence = !fence
		}
		if !fence && strings.HasPrefix(line, "#") {
			lines[i] = strings.Repeat("#", by) + line
		}
	}
	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------- API

func (g *generator) apiSection(d showcase.Demo) (string, map[string]string) {
	var b strings.Builder
	var funcs []string
	var signatureRefs []string
	for _, s := range g.demoSymbols(d) {
		if f, ok := g.funcs[s]; ok {
			funcs = append(funcs, g.funcDecl(f))
			g.covered[s] = true
			signatureRefs = append(signatureRefs, g.referencedTypes(f.Decl.Type, "")...)
		} else if _, ok := g.typeFns[s]; !ok {
			fmt.Fprintf(os.Stderr, "warning: %s: symbol %s is not an exported package function\n", d.ID, s)
		}
	}
	if len(funcs) > 0 {
		b.WriteString("```go\n" + strings.Join(funcs, "\n") + "```\n\n")
	}
	own := g.demoTypes(d)
	linked := map[string]string{}
	seen := map[string]bool{}
	queue := slices.Clone(own)
	depth := map[string]int{}
	for _, ref := range signatureRefs {
		if slices.Contains(own, ref) || slices.Contains(queue, ref) || linked[ref] != "" {
			continue
		}
		if owners := g.owner[ref]; len(owners) > 0 && !slices.Contains(owners, d.ID) {
			linked[ref] = owners[0]
			continue
		}
		depth[ref] = 1
		queue = append(queue, ref)
	}
	var related []string
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if seen[name] {
			continue
		}
		seen[name] = true
		t, ok := g.types[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: %s: type %s is not exported\n", d.ID, name)
			continue
		}
		if !slices.Contains(own, name) {
			related = append(related, name)
		}
		g.covered[name] = true
		for _, f := range t.Funcs {
			g.covered[f.Name] = true
		}
		if depth[name] >= 2 {
			continue
		}
		refs := g.referencedTypes(g.typeSpec(t).Type, t.Name)
		for _, m := range append(slices.Clone(t.Methods), t.Funcs...) {
			refs = append(refs, g.referencedTypes(m.Decl.Type, t.Name)...)
		}
		for _, ref := range refs {
			if seen[ref] || slices.Contains(own, ref) {
				continue
			}
			if owners := g.owner[ref]; len(owners) > 0 && !slices.Contains(owners, d.ID) {
				linked[ref] = owners[0]
				continue
			}
			depth[ref] = depth[name] + 1
			queue = append(queue, ref)
		}
	}
	for _, name := range own {
		if t, ok := g.types[name]; ok {
			fmt.Fprintf(&b, "### %s\n\n```go\n%s```\n\n", name, g.typeDecl(t))
		}
	}
	if len(related) > 0 {
		// Discovery order: types named directly by the API come first.
		b.WriteString("### Related types\n\n")
		for _, name := range related {
			fmt.Fprintf(&b, "```go\n%s```\n\n", g.typeDecl(g.types[name]))
		}
	}
	return b.String(), linked
}

// referencedTypes lists the package's exported types named in node.
func (g *generator) referencedTypes(node ast.Node, self string) []string {
	var out []string
	seen := map[string]bool{}
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			return false // another package
		case *ast.Ident:
			if _, ok := g.types[n.Name]; ok && n.Name != self && !seen[n.Name] {
				seen[n.Name] = true
				out = append(out, n.Name)
			}
		}
		return true
	})
	return out
}

func (g *generator) typeSpec(t *doc.Type) *ast.TypeSpec {
	for _, s := range t.Decl.Specs {
		if ts, ok := s.(*ast.TypeSpec); ok && ts.Name.Name == t.Name {
			return ts
		}
	}
	panic("missing type spec " + t.Name)
}

func (g *generator) typeDecl(t *doc.Type) string {
	spec := *g.typeSpec(t)
	spec.Doc = nil
	var b strings.Builder
	b.WriteString(commentLines(t.Doc))
	b.WriteString("type " + g.print(&spec) + "\n")
	for _, v := range append(slices.Clone(t.Consts), t.Vars...) {
		decl := *v.Decl
		decl.Doc = nil
		b.WriteString("\n" + commentLines(v.Doc) + g.print(&decl) + "\n")
	}
	for _, f := range t.Funcs {
		b.WriteString("\n" + g.funcDecl(f))
	}
	for _, m := range t.Methods {
		b.WriteString("\n" + g.funcDecl(m))
	}
	return b.String()
}

func (g *generator) funcDecl(f *doc.Func) string {
	decl := *f.Decl
	decl.Doc = nil
	decl.Body = nil
	return commentLines(f.Doc) + g.print(&decl) + "\n"
}

// print formats a node with the comments attached to its remaining fields and
// specs, so comments of filtered unexported fields are not printed.
func (g *generator) print(node ast.Node) string {
	var comments []*ast.CommentGroup
	add := func(c *ast.CommentGroup) {
		if c != nil {
			comments = append(comments, c)
		}
	}
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Field:
			add(n.Doc)
			add(n.Comment)
		case *ast.ValueSpec:
			add(n.Doc)
			add(n.Comment)
		case *ast.TypeSpec:
			add(n.Comment)
		}
		return true
	})
	sort.Slice(comments, func(i, j int) bool { return comments[i].Pos() < comments[j].Pos() })
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&buf, g.fset, &printer.CommentedNode{Node: node, Comments: comments}); err != nil {
		panic(err)
	}
	return blankBeforeClose.ReplaceAllString(buf.String(), "\n$1")
}

func commentLines(text string) string {
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return ""
	}
	var b strings.Builder
	for line := range strings.SplitSeq(text, "\n") {
		if line == "" {
			b.WriteString("//\n")
		} else {
			b.WriteString("// " + line + "\n")
		}
	}
	return b.String()
}

func (g *generator) reportUncovered() {
	var missing []string
	for name := range g.funcs {
		if ast.IsExported(name) && !g.covered[name] {
			missing = append(missing, name+"()")
		}
	}
	for name := range g.types {
		if !g.covered[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		fmt.Fprintf(os.Stderr, "note: exported API not covered by any reference (assign in extraSymbols/extraTypes):\n  %s\n", strings.Join(missing, ", "))
	}
}

// checkSkillIndex keeps SKILL.md's component index in step with the catalog.
func (g *generator) checkSkillIndex() error {
	data, err := os.ReadFile(filepath.Join(g.skill, "SKILL.md"))
	if err != nil {
		return err
	}
	var missing []string
	for _, d := range g.demos {
		if !bytes.Contains(data, []byte("`"+d.ID+"`")) {
			missing = append(missing, d.ID)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("SKILL.md component index is missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ---------------------------------------------------------------- docs

type mdFile struct {
	name     string
	title    string // the level-1 heading
	lines    []string
	sections []*mdSection // level-2 sections
}

type mdSection struct {
	heading    string
	start, end int // heading line, exclusive end
	children   []*mdSection
}

func parseMarkdown(name, text string) *mdFile {
	f := &mdFile{name: name, lines: strings.Split(strings.TrimRight(text, "\n"), "\n")}
	fence := false
	var h2, h3 *mdSection
	for i, line := range f.lines {
		if strings.HasPrefix(line, "```") {
			fence = !fence
		}
		if fence {
			continue
		}
		if f.title == "" && strings.HasPrefix(line, "# ") {
			f.title = strings.TrimSpace(line[2:])
		}
		switch {
		case strings.HasPrefix(line, "## "):
			if h3 != nil {
				h3.end = i
				h3 = nil
			}
			if h2 != nil {
				h2.end = i
			}
			h2 = &mdSection{heading: strings.TrimSpace(line[3:]), start: i}
			f.sections = append(f.sections, h2)
		case strings.HasPrefix(line, "### ") && h2 != nil:
			if h3 != nil {
				h3.end = i
			}
			h3 = &mdSection{heading: strings.TrimSpace(line[4:]), start: i}
			h2.children = append(h2.children, h3)
		}
	}
	if h3 != nil {
		h3.end = len(f.lines)
	}
	if h2 != nil {
		h2.end = len(f.lines)
	}
	return f
}

func (g *generator) mentions(text string, names []string) bool {
	for _, n := range names {
		if mentions(text, n, true) {
			return true
		}
	}
	return false
}

// mentionsInProse is mentions without plurals, which in running text are
// usually ordinary words ("Buttons are centered").
func (g *generator) mentionsInProse(text string, names []string) bool {
	for _, n := range names {
		if mentions(text, n, false) {
			return true
		}
	}
	return false
}

// mentions reports whether name occurs as a whole word, optionally with a
// plural suffix ("Buttons", "Switches"), or with a possessive ("Card's").
func mentions(text, name string, plural bool) bool {
	for i := 0; i < len(text); {
		j := strings.Index(text[i:], name)
		if j < 0 {
			return false
		}
		j += i
		before := true
		if j > 0 {
			r, _ := utf8.DecodeLastRuneInString(text[:j])
			before = !isWord(r)
		}
		rest := text[j+len(name):]
		for _, suffix := range []string{"es", "s", "'s", "’s", ""} {
			if !plural && (suffix == "es" || suffix == "s") {
				continue
			}
			if strings.HasPrefix(rest, suffix) {
				tail := rest[len(suffix):]
				r, _ := utf8.DecodeRuneInString(tail)
				if before && (tail == "" || !isWord(r)) {
					return true
				}
			}
		}
		i = j + len(name)
	}
	return false
}

func isWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

func (g *generator) sectionKey(file, heading string) string { return file + "#" + heading }

func (g *generator) headingMatches(d showcase.Demo, file, heading string) bool {
	key := g.sectionKey(file, heading)
	if slices.Contains(skipSections[d.ID], key) {
		return false
	}
	return slices.Contains(extraSections[d.ID], key) || g.mentions(heading, g.names(d))
}

// namesOther reports whether a heading names another component but not d.
func (g *generator) namesOther(d showcase.Demo, heading string) bool {
	for n, id := range g.allName {
		if id != d.ID && !stopNames[n] && mentions(heading, n, true) {
			return true
		}
	}
	return false
}

func (g *generator) docExcerpts(d showcase.Demo) string {
	topic := categories[d.Category].doc
	order := func(f *mdFile) int {
		switch f.name {
		case topic:
			return 0
		case "options.md":
			return 1
		case "compatibility.md":
			return 3
		case "i18n.md":
			return 4
		}
		return 2
	}
	files := slices.Clone(g.docs)
	slices.SortStableFunc(files, func(a, b *mdFile) int {
		if c := order(a) - order(b); c != 0 {
			return c
		}
		return strings.Compare(a.name, b.name)
	})
	var b strings.Builder
	for _, f := range files {
		if text := g.fileExcerpts(d, f); text != "" {
			fmt.Fprintf(&b, "### %s\n\n%s\n", f.title, text)
		}
	}
	return b.String()
}

func (g *generator) fileExcerpts(d showcase.Demo, f *mdFile) string {
	var parts []string
	names := g.names(d)
	for _, h2 := range f.sections {
		introEnd := h2.end
		if len(h2.children) > 0 {
			introEnd = h2.children[0].start
		}
		if g.headingMatches(d, f.name, h2.heading) {
			var lines []string
			lines = append(lines, "#### "+h2.heading)
			lines = append(lines, f.lines[h2.start+1:introEnd]...)
			g.record(d, f.name, h2.heading)
			for _, h3 := range h2.children {
				key := g.sectionKey(f.name, h3.heading)
				if slices.Contains(skipSections[d.ID], key) {
					continue
				}
				if g.namesOther(d, h3.heading) && !g.mentions(h3.heading, names) {
					continue
				}
				lines = append(lines, "##### "+h3.heading)
				lines = append(lines, f.lines[h3.start+1:h3.end]...)
				g.record(d, f.name, h3.heading)
			}
			parts = appendBody(parts, "", g.cleanExcerpt(d, f.name, lines))
			continue
		}
		if wholeSectionsOnly[d.ID] {
			for _, h3 := range h2.children {
				if g.headingMatches(d, f.name, h3.heading) {
					lines := append([]string{"#### " + h2.heading + " › " + h3.heading}, f.lines[h3.start+1:h3.end]...)
					g.record(d, f.name, h3.heading)
					parts = appendBody(parts, "", g.cleanExcerpt(d, f.name, lines))
				}
			}
			continue
		}
		if text := g.partial(d, f, h2.start+1, introEnd); text != "" {
			parts = appendBody(parts, "#### "+h2.heading+" (excerpt)\n\n", g.cleanExcerpt(d, f.name, strings.Split(text, "\n")))
		}
		for _, h3 := range h2.children {
			if g.headingMatches(d, f.name, h3.heading) {
				lines := append([]string{"#### " + h2.heading + " › " + h3.heading}, f.lines[h3.start+1:h3.end]...)
				g.record(d, f.name, h3.heading)
				parts = appendBody(parts, "", g.cleanExcerpt(d, f.name, lines))
				continue
			}
			if text := g.partial(d, f, h3.start+1, h3.end); text != "" {
				parts = appendBody(parts, "#### "+h2.heading+" › "+h3.heading+" (excerpt)\n\n", g.cleanExcerpt(d, f.name, strings.Split(text, "\n")))
			}
		}
	}
	return strings.Join(parts, "\n")
}

// partial selects table rows whose first cell names the component, paragraphs
// that name it, and code blocks that directly follow a selected paragraph.
func (g *generator) partial(d showcase.Demo, f *mdFile, start, end int) string {
	names := g.names(d)
	var out []string
	lastText := false
	for i := start; i < end; {
		line := f.lines[i]
		switch {
		case strings.TrimSpace(line) == "":
			i++
		case strings.HasPrefix(line, "```"):
			j := i + 1
			for j < end && !strings.HasPrefix(f.lines[j], "```") {
				j++
			}
			j = min(j+1, end)
			if lastText {
				out = append(out, strings.Join(f.lines[i:j], "\n"))
			}
			lastText = false
			i = j
		case strings.HasPrefix(line, "|"):
			j := i
			for j < end && strings.HasPrefix(f.lines[j], "|") {
				j++
			}
			var rows []string
			for k := i + 2; k < j; k++ {
				cells := strings.Split(f.lines[k], "|")
				if len(cells) > 1 && g.mentions(cells[1], names) {
					rows = append(rows, f.lines[k])
				}
			}
			if len(rows) > 0 && j-i >= 2 {
				out = append(out, strings.Join(append([]string{f.lines[i], f.lines[i+1]}, rows...), "\n"))
			}
			lastText = false
			i = j
		default:
			j := i
			for j < end && strings.TrimSpace(f.lines[j]) != "" && !strings.HasPrefix(f.lines[j], "```") && !strings.HasPrefix(f.lines[j], "|") {
				j++
			}
			para := strings.Join(f.lines[i:j], "\n")
			lastText = g.mentionsInProse(para, names)
			if lastText {
				out = append(out, para)
			}
			i = j
		}
	}
	return strings.Join(out, "\n\n")
}

var (
	blankBeforeClose = regexp.MustCompile(`\n[ \t]*\n([ \t]*[})])`)
	blankAfterOpen   = regexp.MustCompile(`([{(]\n)[ \t]*\n`)
	blankRuns        = regexp.MustCompile(`\n[ \t]*\n([ \t]*\n)+`)
)

var mdLink = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)

// cleanExcerpt makes documentation text self-contained: it drops sentences
// about repository-only resources (Gallery, website, example apps), resolves
// links to sibling component references, and removes headings left without
// content.
func (g *generator) cleanExcerpt(d showcase.Demo, file string, lines []string) string {
	text := filterProse(strings.Join(lines, "\n"), repoOnly)
	text = mapProse(text, true, func(sentence string) string { return g.rewriteLinks(d, file, sentence) })
	text = dropEmptyHeadings(text)
	// Separate a heading from following text for readability.
	text = regexp.MustCompile(`(?m)^(#{4,5} [^\n]+)\n([^\n])`).ReplaceAllString(text, "$1\n\n$2")
	return strings.TrimSpace(text) + "\n"
}

// record notes that d's reference contains a documentation section in full.
func (g *generator) record(d showcase.Demo, file, heading string) {
	key := file + "#" + anchor(heading)
	if !slices.Contains(g.include[key], d.ID) {
		g.include[key] = append(g.include[key], d.ID)
	}
}

// anchor converts a heading to its GitHub-style fragment.
func anchor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}

var pointerSentence = regexp.MustCompile(`(?i)^\(?(?:for [^,]+, )?see\b`)

// rewriteLinks resolves the documentation links in one sentence. A link to a
// section of this reference becomes plain text, a link to a section in
// another reference points to that file, and a link to anything else is
// removed: a sentence that only points elsewhere is dropped, and a trailing
// "; see …" clause is cut.
func (g *generator) rewriteLinks(d showcase.Demo, file, sentence string) string {
	for _, m := range mdLink.FindAllStringSubmatch(sentence, -1) {
		whole, label, target := m[0], m[1], m[2]
		if replacement, ok := g.resolveLink(d, file, label, target); ok {
			sentence = strings.Replace(sentence, whole, replacement, 1)
			continue
		}
		if pointerSentence.MatchString(sentence) {
			return ""
		}
		clause := regexp.MustCompile(`(?i)(?:[;,]\s*|\s*\()see (?:the )?` + regexp.QuoteMeta(whole) + `\)?`)
		if loc := clause.FindStringIndex(sentence); loc != nil {
			sentence = sentence[:loc[0]] + sentence[loc[1]:]
		} else {
			sentence = strings.Replace(sentence, whole, label, 1)
		}
	}
	return sentence
}

// resolveLink returns the replacement for a documentation link, or false when
// its target is not part of any reference.
func (g *generator) resolveLink(d showcase.Demo, file, label, target string) (string, bool) {
	if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
		return "[" + label + "](" + target + ")", true
	}
	link := func(id string) (string, bool) {
		if id == d.ID {
			return label, true
		}
		return "[" + label + "](" + id + ".md)", true
	}
	if id, ok := g.allName[strings.Trim(label, "`")]; ok {
		return link(id)
	}
	doc, frag, found := strings.Cut(target, "#")
	if !found || frag == "" {
		return "", false
	}
	if doc == "" {
		doc = file
	}
	ids := g.include[filepath.Base(doc)+"#"+frag]
	if slices.Contains(ids, d.ID) {
		return label, true
	}
	for _, id := range ids {
		if g.mentions(label, g.namesOf(id)) {
			return link(id)
		}
	}
	if len(ids) > 0 {
		return link(ids[0])
	}
	return "", false
}

func (g *generator) namesOf(id string) []string {
	for _, d := range g.demos {
		if d.ID == id {
			return g.names(d)
		}
	}
	return nil
}

// appendBody adds heading+body to parts when body has content besides headings.
func appendBody(parts []string, heading, body string) []string {
	for line := range strings.SplitSeq(body, "\n") {
		if t := strings.TrimSpace(line); t != "" && !strings.HasPrefix(t, "#") {
			return append(parts, heading+body)
		}
	}
	return parts
}

// dropEmptyHeadings removes headings followed only by headings of the same or
// a higher level, or by nothing.
func dropEmptyHeadings(text string) string {
	level := func(line string) int {
		n := len(line) - len(strings.TrimLeft(line, "#"))
		if n == 0 || !strings.HasPrefix(line[n:], " ") {
			return 0
		}
		return n
	}
	lines := strings.Split(text, "\n")
	for changed := true; changed; {
		changed = false
		fence := false
		for i, line := range lines {
			if strings.HasPrefix(line, "```") {
				fence = !fence
			}
			l := level(line)
			if fence || l == 0 {
				continue
			}
			next := 0 // 0: content follows; -1: nothing follows; >0: heading level
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j == len(lines) {
				next = -1
			} else if nl := level(lines[j]); nl > 0 {
				next = nl
			}
			if next == -1 || next > 0 && next <= l {
				lines = slices.Delete(lines, i, i+1)
				changed = true
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

var (
	// repoOnly matches sentences about resources that exist only in the
	// nuxt-ui-go repository.
	repoOnly = regexp.MustCompile(`(?i)\bgallery\b|\bwebsite\b|\bsettings example\b|\bdata browser example\b|\brun the examples\b|\bexamples/|\bthe examples\b|\binternal/`)
	// previewOnly also matches sentences about the website's interactive preview.
	previewOnly = regexp.MustCompile(repoOnly.String() + `|(?i)\bpreview\b|\bthe example(?:'s|’s)?\b|\btoggle\b|\bresetting\b`)
)

// filterProse removes sentences matching drop from paragraphs and list
// items, and table rows that match it, leaving headings and code intact.
func filterProse(text string, drop *regexp.Regexp) string {
	return mapProse(text, false, func(sentence string) string {
		if drop.MatchString(sentence) {
			return ""
		}
		return sentence
	})
}

// mapProse applies fn to every sentence of paragraphs and list items; an
// empty result removes the sentence. Table rows are passed whole, or cell by
// cell when cells is true. Headings and code blocks are left intact.
func mapProse(text string, cells bool, fn func(string) string) string {
	sentences := func(body string) string {
		var kept []string
		for _, sentence := range splitSentences(body) {
			if r := strings.TrimSpace(fn(sentence)); r != "" {
				kept = append(kept, r)
			}
		}
		return strings.Join(kept, " ")
	}
	var out []string
	fence := false
	for line := range strings.SplitSeq(text, "\n") {
		if strings.HasPrefix(line, "```") {
			fence = !fence
		}
		trimmed := strings.TrimSpace(line)
		switch {
		case fence || strings.HasPrefix(line, "```") || trimmed == "" || strings.HasPrefix(trimmed, "#"):
			out = append(out, line)
		case strings.HasPrefix(trimmed, "|"):
			if !cells {
				if fn(line) != "" {
					out = append(out, line)
				}
				continue
			}
			parts := strings.Split(line, "|")
			for i := 1; i < len(parts)-1; i++ {
				parts[i] = " " + sentences(strings.TrimSpace(parts[i])) + " "
			}
			out = append(out, strings.Join(parts, "|"))
		default:
			prefix, body := "", line
			if m := listItem.FindString(line); m != "" {
				prefix, body = m, line[len(m):]
			}
			if kept := sentences(body); kept != "" {
				out = append(out, prefix+kept)
			}
		}
	}
	return strings.TrimSpace(regexp.MustCompile(`\n{3,}`).ReplaceAllString(strings.Join(out, "\n"), "\n\n"))
}

var listItem = regexp.MustCompile(`^\s*(?:[-*]|\d+\.)\s+`)

// splitSentences splits prose after ". ", "! " or "? " when the next sentence
// starts with a capital letter or Markdown markup, ignoring code spans.
func splitSentences(text string) []string {
	var out []string
	code := false
	start := 0
	for i := 0; i < len(text); i++ {
		switch c := text[i]; {
		case c == '`':
			code = !code
		case !code && (c == '.' || c == '!' || c == '?') && i+2 < len(text) && text[i+1] == ' ':
			r, _ := utf8.DecodeRuneInString(text[i+2:])
			if unicode.IsUpper(r) || strings.ContainsRune("*`[(", r) {
				out = append(out, strings.TrimSpace(text[start:i+1]))
				start = i + 2
			}
		}
	}
	if rest := strings.TrimSpace(text[start:]); rest != "" {
		out = append(out, rest)
	}
	return out
}

// libraryVersion names the latest released version in CHANGELOG.md and the
// MyGo version pinned in go.mod.
func libraryVersion(root string) (string, error) {
	changelog, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md"))
	if err != nil {
		return "", err
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	release := regexp.MustCompile(`(?m)^## (v\d+\.\d+\.\d+)\b`).FindSubmatch(changelog)
	mygo := regexp.MustCompile(`(?m)^\s*github\.com/egoist/mygo (v\S+)`).FindSubmatch(mod)
	if release == nil || mygo == nil {
		return "", fmt.Errorf("cannot read the library or MyGo version")
	}
	version := "nuxt-ui-go " + string(release[1])
	if first := regexp.MustCompile(`(?m)^## (\S+)`).FindSubmatch(changelog); string(first[1]) != string(release[1]) {
		version += " plus unreleased changes"
	}
	return version + " (MyGo " + string(mygo[1]) + ")", nil
}

// ---------------------------------------------------------------- example

func (g *generator) example(d showcase.Demo) (string, error) {
	_, full, err := d.Source(showcase.Config{Locale: "en"})
	if err != nil {
		return "", err
	}
	return cleanExample(full)
}

// cleanExample removes the website's preview toggles from a generated
// program: constant if-branches, cmp.Or with a literal first argument, and
// struct fields set to their zero value.
func cleanExample(src string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "example.go", src, 0)
	if err != nil {
		return "", err
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BlockStmt:
			n.List = cleanStmts(n.List)
		case *ast.CaseClause:
			n.Body = cleanStmts(n.Body)
		case *ast.CommClause:
			n.Body = cleanStmts(n.Body)
		case *ast.CompositeLit:
			elts := n.Elts[:0]
			for _, e := range n.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok {
					kv.Value = simplifyExpr(kv.Value)
					if _, field := kv.Key.(*ast.Ident); field && isZeroLiteral(kv.Value) {
						continue
					}
				}
				elts = append(elts, e)
			}
			n.Elts = elts
		case *ast.CallExpr:
			for i, a := range n.Args {
				n.Args[i] = simplifyExpr(a)
			}
		case *ast.AssignStmt:
			for i, e := range n.Rhs {
				n.Rhs[i] = simplifyExpr(e)
			}
		case *ast.ReturnStmt:
			for i, e := range n.Results {
				n.Results[i] = simplifyBool(e)
			}
		}
		return true
	})
	dropSyncBlocks(f)
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			dropZeroVars(fn.Body)
			dropUnreadVars(fn.Body)
		}
	}
	dropUnusedFields(f)
	var body bytes.Buffer
	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			continue
		}
		if err := format.Node(&body, fset, decl); err != nil {
			return "", err
		}
		body.WriteString("\n\n")
	}
	text := blankBeforeClose.ReplaceAllString(body.String(), "\n$1")
	text = blankAfterOpen.ReplaceAllString(text, "$1")
	text = blankRuns.ReplaceAllString(text, "\n\n")
	return showcase.FormatSource("main", text)
}

// dropSyncBlocks removes `if s.X != zero { s.X = zero; ... }` blocks that
// re-apply a preview control. With the default value and no other writes to
// s.X, they never run.
func dropSyncBlocks(f *ast.File) {
	writes := map[string]int{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, e := range n.Lhs {
				writes[types.ExprString(e)]++
			}
		case *ast.IncDecStmt:
			writes[types.ExprString(n.X)]++
		case *ast.UnaryExpr:
			if n.Op == token.AND {
				writes[types.ExprString(n.X)]++
			}
		}
		return true
	})
	isSync := func(s ast.Stmt) bool {
		ifs, ok := s.(*ast.IfStmt)
		if !ok || ifs.Init != nil || ifs.Else != nil || len(ifs.Body.List) == 0 {
			return false
		}
		cond, ok := ifs.Cond.(*ast.BinaryExpr)
		if !ok || cond.Op != token.NEQ || !isZeroLiteral(cond.Y) {
			return false
		}
		if _, ok := cond.X.(*ast.SelectorExpr); !ok {
			return false
		}
		first, ok := ifs.Body.List[0].(*ast.AssignStmt)
		target := types.ExprString(cond.X)
		return ok && first.Tok == token.ASSIGN && len(first.Lhs) == 1 && len(first.Rhs) == 1 &&
			types.ExprString(first.Lhs[0]) == target && types.ExprString(first.Rhs[0]) == types.ExprString(cond.Y) &&
			writes[target] == 1
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			block.List = slices.DeleteFunc(block.List, isSync)
		}
		return true
	})
}

// dropUnusedFields removes fields of the example's state struct that no
// remaining code reads or writes.
func dropUnusedFields(f *ast.File) {
	used := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			used[sel.Sel.Name] = true
		}
		return true
	})
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts := spec.(*ast.TypeSpec)
			st, ok := ts.Type.(*ast.StructType)
			if !ok || ts.Name.Name != "example" {
				continue
			}
			st.Fields.List = slices.DeleteFunc(st.Fields.List, func(field *ast.Field) bool {
				if len(field.Names) == 0 {
					return false // embedded field
				}
				field.Names = slices.DeleteFunc(field.Names, func(id *ast.Ident) bool { return !used[id.Name] })
				return len(field.Names) == 0
			})
		}
	}
}

// dropZeroVars removes local variables that keep their zero value and are
// only passed as struct fields, together with those fields.
func dropZeroVars(body *ast.BlockStmt) {
	for changed := true; changed; {
		changed = false
		ast.Inspect(body, func(n ast.Node) bool {
			block, ok := n.(*ast.BlockStmt)
			if !ok || changed {
				return !changed
			}
			for i, stmt := range block.List {
				for _, name := range zeroVars(stmt) {
					uses, fields := 0, 0
					ast.Inspect(body, func(n ast.Node) bool {
						switch n := n.(type) {
						case ast.Stmt:
							if n == stmt {
								return false
							}
						case *ast.KeyValueExpr:
							if id, ok := n.Value.(*ast.Ident); ok && id.Name == name {
								fields++
							}
						case *ast.Ident:
							if n.Name == name {
								uses++
							}
						}
						return true
					})
					if uses == 0 || uses != fields {
						continue
					}
					ast.Inspect(body, func(n ast.Node) bool {
						if lit, ok := n.(*ast.CompositeLit); ok {
							lit.Elts = slices.DeleteFunc(lit.Elts, func(e ast.Expr) bool {
								kv, ok := e.(*ast.KeyValueExpr)
								if !ok {
									return false
								}
								id, ok := kv.Value.(*ast.Ident)
								return ok && id.Name == name
							})
						}
						return true
					})
					if !removeDeclaredName(stmt, name) {
						block.List = slices.Delete(block.List, i, i+1)
					}
					changed = true
					return false
				}
			}
			return true
		})
	}
}

// removeDeclaredName drops name from a multi-name var declaration and
// reports whether the statement still declares other names.
func removeDeclaredName(stmt ast.Stmt, name string) bool {
	ds, ok := stmt.(*ast.DeclStmt)
	if !ok {
		return false
	}
	vs := ds.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	vs.Names = slices.DeleteFunc(vs.Names, func(id *ast.Ident) bool { return id.Name == name })
	return len(vs.Names) > 0
}

// dropUnreadVars removes local variables whose only reads were in removed
// preview branches. Assignments from calls stay as call statements because the
// calls build widgets.
func dropUnreadVars(body *ast.BlockStmt) {
	for changed := true; changed; {
		changed = false
		declared := map[string]bool{}
		writes := map[*ast.Ident]bool{}
		ast.Inspect(body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.ValueSpec:
				for _, id := range n.Names {
					declared[id.Name] = true
					writes[id] = true
				}
			case *ast.AssignStmt:
				for _, e := range n.Lhs {
					if id, ok := e.(*ast.Ident); ok {
						if n.Tok == token.DEFINE {
							declared[id.Name] = true
						}
						writes[id] = true
					}
				}
			}
			return true
		})
		reads := map[string]int{}
		ast.Inspect(body, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && !writes[id] {
				reads[id.Name]++
			}
			return true
		})
		for name := range declared {
			if name == "_" || reads[name] > 0 {
				continue
			}
			changed = removeVar(body, name) || changed
		}
	}
}

func removeVar(body *ast.BlockStmt, name string) bool {
	removed := false
	rewrite := func(list []ast.Stmt) []ast.Stmt {
		var out []ast.Stmt
		for _, stmt := range list {
			switch s := stmt.(type) {
			case *ast.DeclStmt:
				if gd, ok := s.Decl.(*ast.GenDecl); ok && gd.Tok == token.VAR && len(gd.Specs) == 1 {
					vs := gd.Specs[0].(*ast.ValueSpec)
					if len(vs.Values) == 0 && slices.ContainsFunc(vs.Names, func(id *ast.Ident) bool { return id.Name == name }) {
						vs.Names = slices.DeleteFunc(vs.Names, func(id *ast.Ident) bool { return id.Name == name })
						removed = true
						if len(vs.Names) == 0 {
							continue
						}
					}
				}
			case *ast.AssignStmt:
				if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
					if id, ok := s.Lhs[0].(*ast.Ident); ok && id.Name == name {
						removed = true
						if call, ok := s.Rhs[0].(*ast.CallExpr); ok {
							out = append(out, &ast.ExprStmt{X: call})
						}
						continue
					}
				}
			}
			out = append(out, stmt)
		}
		return out
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BlockStmt:
			n.List = rewrite(n.List)
		case *ast.CaseClause:
			n.Body = rewrite(n.Body)
		}
		return true
	})
	return removed
}

// zeroVars returns the names declared by `var x, y T` with a nilable T, or
// by `x := <zero literal>`.
func zeroVars(stmt ast.Stmt) []string {
	switch s := stmt.(type) {
	case *ast.DeclStmt:
		gd, ok := s.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR || len(gd.Specs) != 1 {
			return nil
		}
		vs := gd.Specs[0].(*ast.ValueSpec)
		if len(vs.Values) != 0 {
			return nil
		}
		nilable := false
		switch t := vs.Type.(type) {
		case *ast.StarExpr, *ast.FuncType, *ast.MapType, *ast.InterfaceType:
			nilable = true
		case *ast.ArrayType:
			nilable = t.Len == nil
		}
		if !nilable {
			return nil
		}
		var names []string
		for _, id := range vs.Names {
			names = append(names, id.Name)
		}
		return names
	case *ast.AssignStmt:
		if s.Tok == token.DEFINE && len(s.Lhs) == 1 && len(s.Rhs) == 1 && isZeroLiteral(s.Rhs[0]) {
			if id, ok := s.Lhs[0].(*ast.Ident); ok {
				return []string{id.Name}
			}
		}
	}
	return nil
}

func cleanStmts(list []ast.Stmt) []ast.Stmt {
	var out []ast.Stmt
	for _, s := range list {
		if ifs, ok := s.(*ast.IfStmt); ok && ifs.Init == nil {
			ifs.Cond = simplifyBool(ifs.Cond)
			if cond, ok := ifs.Cond.(*ast.Ident); ok && (cond.Name == "false" || cond.Name == "true") {
				if cond.Name == "true" {
					out = append(out, cleanStmts(ifs.Body.List)...)
					continue
				}
				switch e := ifs.Else.(type) {
				case *ast.BlockStmt:
					out = append(out, cleanStmts(e.List)...)
				case *ast.IfStmt:
					out = append(out, cleanStmts([]ast.Stmt{e})...)
				}
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// simplifyBool folds the literal true/false operands that preview toggles
// leave in conditions, such as `false && x` or `false || x`.
func simplifyBool(e ast.Expr) ast.Expr {
	lit := func(e ast.Expr) string {
		if id, ok := e.(*ast.Ident); ok && (id.Name == "true" || id.Name == "false") {
			return id.Name
		}
		return ""
	}
	switch b := e.(type) {
	case *ast.UnaryExpr:
		if b.Op == token.NOT {
			switch lit(simplifyBool(b.X)) {
			case "true":
				return ast.NewIdent("false")
			case "false":
				return ast.NewIdent("true")
			}
		}
	case *ast.ParenExpr:
		inner := simplifyBool(b.X)
		if lit(inner) != "" {
			return inner
		}
		b.X = inner
	case *ast.BinaryExpr:
		if b.Op != token.LAND && b.Op != token.LOR {
			return e
		}
		b.X, b.Y = simplifyBool(b.X), simplifyBool(b.Y)
		for _, pair := range [][2]ast.Expr{{b.X, b.Y}, {b.Y, b.X}} {
			switch v := lit(pair[0]); {
			case v == "false" && b.Op == token.LAND, v == "true" && b.Op == token.LOR:
				return pair[0]
			case v == "true" && b.Op == token.LAND, v == "false" && b.Op == token.LOR:
				return pair[1]
			}
		}
	}
	return e
}

func simplifyExpr(e ast.Expr) ast.Expr {
	call, ok := e.(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return e
	}
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "cmp" && sel.Sel.Name == "Or" {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value != `""` && lit.Value != "``" {
				return lit
			}
		}
	}
	return e
}

func isZeroLiteral(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name == "false" || e.Name == "nil"
	case *ast.BasicLit:
		switch e.Kind {
		case token.STRING:
			return e.Value == `""` || e.Value == "``"
		case token.INT, token.FLOAT:
			v, err := strconv.ParseFloat(e.Value, 64)
			return err == nil && v == 0
		}
	case *ast.SelectorExpr:
		if pkg, ok := e.X.(*ast.Ident); ok && pkg.Name == "nui" {
			switch e.Sel.Name {
			case "ColorDefault", "SizeDefault", "VariantDefault":
				return true
			}
		}
	case *ast.CompositeLit:
		// nui.Optional[T]{} means "unset".
		if idx, ok := e.Type.(*ast.IndexExpr); ok && len(e.Elts) == 0 {
			if sel, ok := idx.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Optional" {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------- output

func syncFiles(dir string, files map[string][]byte, check bool) []string {
	var changed []string
	existing, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	for _, p := range existing {
		if _, ok := files[filepath.Base(p)]; !ok {
			changed = append(changed, filepath.Base(p)+" (obsolete)")
			if !check {
				if err := os.Remove(p); err != nil {
					fail(err)
				}
			}
		}
	}
	if !check {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fail(err)
		}
	}
	for _, name := range sortedKeys(files) {
		p := filepath.Join(dir, name)
		old, err := os.ReadFile(p)
		if err == nil && bytes.Equal(old, files[name]) {
			continue
		}
		changed = append(changed, name)
		if !check {
			if err := os.WriteFile(p, files[name], 0o644); err != nil {
				fail(err)
			}
		}
	}
	return changed
}
