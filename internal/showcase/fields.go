package showcase

//go:generate go run ../cmd/genshowcase -root ../..

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

type enumValue struct {
	Name  string
	Value reflect.Value
}

// ValueSchema describes a public Go value without exposing private widget state.
type ValueSchema struct {
	Type      string       `json:"type"`
	Kind      string       `json:"kind"`
	Item      string       `json:"item,omitempty"`
	Fields    []ValueField `json:"fields,omitempty"`
	Choices   []string     `json:"choices,omitempty"`
	Length    int          `json:"length,omitempty"`
	Sample    any          `json:"sample"`
	Returns   []string     `json:"returns,omitempty"`
	Arguments []string     `json:"arguments,omitempty"`
}

type ValueField struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description Text            `json:"description"`
	Control     string          `json:"control,omitempty"`
	Choices     []string        `json:"choices,omitempty"`
	Hint        *Text           `json:"hint,omitempty"`
	When        *FieldCondition `json:"when,omitempty"`
	Slots       []string        `json:"slots,omitempty"`
}

type FieldCondition struct {
	Field string `json:"field"`
	Value string `json:"value"`
	Label Text   `json:"label"`
}

type Binding struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Fields     []ValueField `json:"fields,omitempty"`
	Defaults   any          `json:"defaults"`
	typ        reflect.Type
	start, end int
	literal    ast.Expr
}

func goType(t reflect.Type) string {
	if t.PkgPath() == "github.com/josexy/nuxt-ui-go" {
		return "nui." + strings.NewReplacer("github.com/josexy/nuxt-ui-go.", "nui.", "github.com/egoist/mygo/ui.", "ui.").Replace(t.Name())
	}
	if t.Name() != "" {
		return t.String()
	}
	switch t.Kind() {
	case reflect.Pointer:
		return "*" + goType(t.Elem())
	case reflect.Slice:
		return "[]" + goType(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), goType(t.Elem()))
	case reflect.Map:
		return "map[" + goType(t.Key()) + "]" + goType(t.Elem())
	case reflect.Func:
		var in, out []string
		for i := 0; i < t.NumIn(); i++ {
			in = append(in, goType(t.In(i)))
		}
		for i := 0; i < t.NumOut(); i++ {
			out = append(out, goType(t.Out(i)))
		}
		result := "func(" + strings.Join(in, ", ") + ")"
		if len(out) == 1 {
			result += " " + out[0]
		} else if len(out) > 1 {
			result += " (" + strings.Join(out, ", ") + ")"
		}
		return result
	}
	return t.String()
}

func optional(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && strings.HasPrefix(t.Name(), "Optional[") && t.PkgPath() == "github.com/josexy/nuxt-ui-go"
}

func publicFields(t reflect.Type) []reflect.StructField {
	var result []reflect.StructField
	if t.Kind() != reflect.Struct {
		return result
	}
	for _, field := range reflect.VisibleFields(t) {
		if field.IsExported() && !field.Anonymous {
			result = append(result, field)
		}
	}
	return result
}

func fieldSchemas(t reflect.Type) []ValueField {
	notes := APINotes()
	var fields []ValueField
	for _, field := range publicFields(t) {
		choices := enumNames(field.Type, t.Name(), field.Name)
		if field.Name == "FilterFields" {
			choices = []string{"label", "description", "value"}
			if t.Name() == "SelectMenuOptions" {
				choices = append(choices, "id", "group")
			}
		}
		if field.Type.Kind() == reflect.Func {
			choices = fieldFunctionModes(field.Type, field.Name)
		}
		description := notes[field.Name]
		if specific, ok := notes[t.Name()+"."+field.Name]; ok {
			description = specific
		}
		editor := ValueField{Name: field.Name, Type: goType(field.Type), Description: description, Choices: choices}
		if t == reflect.TypeFor[nui.InputDateOptions]() || t == reflect.TypeFor[nui.InputTimeOptions]() {
			switch field.Name {
			case "Placeholder", "Password", "MaxLength":
				editor.Hint = &Text{ZH: "分段输入不适用", EN: "Not used by segments"}
			case "UI":
				editor.Slots = []string{"root", "editor", "segment", "leading", "trailing"}
			case "Style":
				editor.Slots = []string{"root", "editor", "leading", "trailing"}
			}
		}
		if t == reflect.TypeFor[nui.PopoverOptions]() {
			switch field.Name {
			case "Label":
				editor.Hint = &Text{ZH: "无障碍名称", EN: "Accessible name"}
			case "OpenDelay", "CloseDelay":
				editor.When = &FieldCondition{Field: "Mode", Value: "nui.PopoverHover", Label: Text{ZH: "仅悬停", EN: "Hover only"}}
			case "UI", "Style":
				editor.Slots = []string{"panel"}
			}
		}
		fields = append(fields, editor)
	}
	return fields
}

func enumNames(t reflect.Type, owner, field string) []string {
	if t == reflect.TypeFor[nui.Side]() && owner == "SidebarOptions" {
		return []string{"nui.SideLeft", "nui.SideRight"}
	}
	var values []string
	for _, v := range enumValues[t] {
		values = append(values, v.Name)
	}
	if t == reflect.TypeFor[nui.Variant]() {
		var allowed []string
		switch owner {
		case "InputOptions", "TextareaOptions", "SelectOptions", "InputMenuOptions", "InputNumberOptions", "InputTagsOptions", "PinInputOptions", "SelectMenuOptions", "InputDateOptions", "InputTimeOptions":
			allowed = []string{"nui.VariantDefault", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.Ghost", "nui.None"}
		case "BadgeOptions", "AlertOptions", "CardOptions", "KbdOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle"}
		case "EmptyOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.None"}
		case "ButtonOptions", "ComponentTheme", "PaginationOptions":
			allowed = []string{"nui.VariantDefault", "nui.Solid", "nui.Outline", "nui.Soft", "nui.Subtle", "nui.Ghost", "nui.LinkVariant"}
		}
		if len(allowed) > 0 && (field == "Variant" || field == "ActiveVariant") {
			values = allowed
		}
	}
	return values
}

// Schemas uses reflection only for exported configuration values. Handles,
// bitmaps and callbacks have dedicated editors instead of arbitrary Go code.
func Schemas() map[string]ValueSchema {
	out := map[string]ValueSchema{}
	var visit func(reflect.Type)
	visit = func(t reflect.Type) {
		name := goType(t)
		if _, ok := out[name]; ok {
			return
		}
		s := ValueSchema{Type: name, Kind: "object", Sample: map[string]any{}}
		out[name] = s
		switch {
		case t == reflect.TypeFor[*mygo.Window]():
			s.Kind = "window"
			s.Sample = "current"
		case t == reflect.TypeFor[ui.Color]():
			s.Kind = "color"
			s.Sample = "#22c55e"
		case t == reflect.TypeFor[*ui.SVG]():
			s.Kind = "icon"
			s.Choices = iconNames()
			s.Sample = "check"
		case t == reflect.TypeFor[*ui.Bitmap]():
			s.Kind = "image"
			s.Sample = nil
		case t == reflect.TypeFor[*ui.Handle]():
			s.Kind = "handle"
			s.Sample = "target"
		case t == reflect.TypeFor[*ui.ScrollState]():
			s.Kind = "scroll"
			s.Sample = map[string]any{"X": float64(0), "Y": float64(0)}
		case t == reflect.TypeFor[nui.Date]():
			s.Kind = "date"
			s.Sample = "2026-11-02"
		case t == reflect.TypeFor[time.Duration]():
			s.Kind = "duration"
			s.Sample = float64(150)
		case t == reflect.TypeFor[ui.Modifiers]():
			s.Kind = "flags"
			s.Choices = []string{"ui.Ctrl", "ui.Shift", "ui.Alt", "ui.Super"}
			s.Sample = []any{}
		case optional(t):
			s.Kind = "optional"
			s.Item = goType(t.Field(0).Type)
			s.Sample = nil
			visit(t.Field(0).Type)
		case len(enumValues[t]) > 0:
			s.Kind = "enum"
			s.Choices = enumNames(t, "", "")
			s.Sample = s.Choices[0]
		default:
			switch t.Kind() {
			case reflect.String:
				s.Kind = "text"
				s.Sample = ""
			case reflect.Bool:
				s.Kind = "boolean"
				s.Sample = false
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				s.Kind = "integer"
				s.Sample = float64(0)
			case reflect.Float32, reflect.Float64:
				s.Kind = "number"
				s.Sample = float64(0)
			case reflect.Struct:
				s.Fields = fieldSchemas(t)
				for _, field := range publicFields(t) {
					visit(field.Type)
				}
			case reflect.Pointer:
				s.Kind = "pointer"
				s.Item = goType(t.Elem())
				s.Sample = nil
				visit(t.Elem())
			case reflect.Array, reflect.Slice:
				s.Kind = "array"
				s.Item = goType(t.Elem())
				s.Sample = []any{}
				if t.Kind() == reflect.Array {
					s.Length = t.Len()
				}
				visit(t.Elem())
			case reflect.Map:
				s.Kind = "map"
				s.Item = goType(t.Elem())
				for _, v := range enumValues[t.Key()] {
					s.Choices = append(s.Choices, fmt.Sprint(v.Value.Interface()))
				}
				visit(t.Elem())
			case reflect.Func:
				s.Kind = "function"
				s.Choices = functionModes(t)
				s.Sample = map[string]any{"mode": s.Choices[0], "text": "Example"}
				for i := 0; i < t.NumIn(); i++ {
					s.Arguments = append(s.Arguments, goType(t.In(i)))
				}
				for i := 0; i < t.NumOut(); i++ {
					if t.Out(i) != reflect.TypeFor[ui.Element]() {
						visit(t.Out(i))
						s.Returns = append(s.Returns, goType(t.Out(i)))
					}
				}
			case reflect.Interface:
				s.Kind = "text"
				s.Sample = ""
			default:
				panic("missing field editor for " + name)
			}
		}
		out[name] = s
	}
	for _, typ := range optionTypes {
		visit(typ)
		visit(reflect.SliceOf(typ))
	}
	return out
}

func nodeText(n ast.Node) string {
	var b bytes.Buffer
	_ = format.Node(&b, token.NewFileSet(), n)
	return b.String()
}

func expressionType(expr ast.Expr) reflect.Type {
	if nodeText(expr) == "string" {
		return reflect.TypeFor[string]()
	}
	if typ := optionTypes[nodeText(expr)]; typ != nil {
		return typ
	}
	if a, ok := expr.(*ast.ArrayType); ok && a.Len == nil {
		if typ := expressionType(a.Elt); typ != nil {
			return reflect.SliceOf(typ)
		}
	}
	return nil
}

func (d Demo) Bindings() []Binding {
	return append(d.bindingsFor(d.RenderBody(Config{}, true)), d.dataBindings()...)
}

func (d Demo) bindingsFor(body string) []Binding {
	prefix := "package main\nfunc example(){\n"
	set := token.NewFileSet()
	f, err := parser.ParseFile(set, "example.go", prefix+body+"\n}", 0)
	if err != nil {
		panic(err)
	}
	// Options can be shared with a trigger callback through a local variable.
	// Resolve only names with one definition/assignment in this example, so a
	// shadowed or reassigned name cannot patch an unrelated initializer.
	assignments := map[string]int{}
	literals := map[string]*ast.CompositeLit{}
	remember := func(name string, expr ast.Expr) {
		assignments[name]++
		literal, _ := expr.(*ast.CompositeLit)
		literals[name] = literal
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.AssignStmt:
			for i, left := range v.Lhs {
				if name, ok := left.(*ast.Ident); ok {
					var right ast.Expr
					if i < len(v.Rhs) {
						right = v.Rhs[i]
					}
					remember(name.Name, right)
				}
			}
		case *ast.ValueSpec:
			for i, name := range v.Names {
				var right ast.Expr
				if i < len(v.Values) {
					right = v.Values[i]
				}
				remember(name.Name, right)
			}
		case *ast.Field:
			for _, name := range v.Names {
				remember(name.Name, nil)
			}
		case *ast.RangeStmt:
			for _, expr := range []ast.Expr{v.Key, v.Value} {
				if name, ok := expr.(*ast.Ident); ok {
					remember(name.Name, nil)
				}
			}
		}
		return true
	})
	literalOf := func(expr ast.Expr) *ast.CompositeLit {
		if name, ok := expr.(*ast.Ident); ok && assignments[name.Name] == 1 {
			return literals[name.Name]
		}
		literal, _ := expr.(*ast.CompositeLit)
		return literal
	}
	var result []Binding
	counts := map[string]int{}
	bound := map[ast.Expr]bool{}
	add := func(expr ast.Expr, name string) {
		if bound[expr] {
			return
		}
		literal, ok := expr.(*ast.CompositeLit)
		if !ok {
			return
		}
		typ := expressionType(literal.Type)
		if typ == nil {
			return
		}
		bound[expr] = true
		counts[name]++
		id := name
		if counts[name] > 1 {
			id += strconv.Itoa(counts[name])
		}
		fields := fieldSchemas(typ)
		for i := range fields {
			expr := literalField(literal, typ, fields[i].Name)
			key := expressionControl(expr)
			if key != "" && fields[i].Name != "AccessibleLabel" {
				fields[i].Control = key
			}
		}
		cfg, _ := d.normalizeValues(Config{})
		defaults := literalDefaults(literal, typ, cfg)
		result = append(result, Binding{ID: id, Name: name, Type: goType(typ), Fields: fields, Defaults: defaults, typ: typ, start: set.Position(expr.Pos()).Offset - len(prefix), end: set.Position(expr.End()).Offset - len(prefix), literal: expr})
	}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		callee := call.Fun
		if generic, ok := callee.(*ast.IndexExpr); ok {
			callee = generic.X
		}
		if generic, ok := callee.(*ast.IndexListExpr); ok {
			callee = generic.X
		}
		fun, ok := callee.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := fun.X.(*ast.Ident)
		if !ok || pkg.Name != "nui" || !slices.Contains(d.Symbols, fun.Sel.Name) {
			return true
		}
		for _, arg := range call.Args {
			if lit := literalOf(arg); lit != nil {
				if typ := expressionType(lit.Type); typ != nil {
					name := typ.Name()
					if name == "" {
						name = fun.Sel.Name + "Items"
					}
					add(lit, name)
				}
			}
		}
		return true
	})
	if d.ID == "theme" {
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok && nodeText(lit.Type) == "nui.ComponentTheme" {
				add(lit, "ComponentTheme")
			}
			return true
		})
		result = append(result, Binding{ID: "Theme", Name: "Theme", Type: "nui.Theme", Fields: fieldSchemas(reflect.TypeFor[nui.Theme]()), Defaults: map[string]any{}, typ: reflect.TypeFor[nui.Theme](), start: -1, end: -1})
	}
	if d.ID == "app" {
		result = append(result, Binding{ID: "AppConfig", Name: "AppConfig", Type: "nui.AppConfig", Fields: fieldSchemas(reflect.TypeFor[nui.AppConfig]()), Defaults: map[string]any{}, typ: reflect.TypeFor[nui.AppConfig](), start: -1, end: -1})
	}
	return result
}

func literalField(lit *ast.CompositeLit, typ reflect.Type, name string) ast.Expr {
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for _, e := range lit.Elts {
		if kv, ok := e.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok {
				if key.Name == name {
					return kv.Value
				}
				if field, ok := typ.FieldByName(key.Name); ok && field.Anonymous {
					if nested, ok := kv.Value.(*ast.CompositeLit); ok {
						if found := literalField(nested, field.Type, name); found != nil {
							return found
						}
					}
				}
			}
		}
	}
	return nil
}

func expressionControl(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	key := ""
	count := 0
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := fn.X.(*ast.Ident)
		if !ok {
			return true
		}
		index := 0
		if pkg.Name == "showcase" && fn.Sel.Name == "Select" {
			index = 1
		} else if pkg.Name != "cfg" {
			return true
		}
		if len(call.Args) > index {
			if lit, ok := call.Args[index].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				key, _ = strconv.Unquote(lit.Value)
				count++
			}
		}
		return true
	})
	if count == 1 {
		return key
	}
	return ""
}

func literalDefaults(expr ast.Expr, typ reflect.Type, cfg Config) any {
	if expr == nil {
		return nil
	}
	if lit, ok := expr.(*ast.BasicLit); ok {
		if lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			return s
		}
		n, err := strconv.ParseFloat(lit.Value, 64)
		if err == nil {
			return n
		}
	}
	if id, ok := expr.(*ast.Ident); ok {
		if id.Name == "true" {
			return true
		}
		if id.Name == "false" {
			return false
		}
		return nil
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		for _, entry := range enumValues[typ] {
			if entry.Name == nodeText(sel) {
				return entry.Name
			}
		}
		return nil
	}
	if optional(typ) {
		if call, ok := expr.(*ast.CallExpr); ok && nodeText(call.Fun) == "nui.Set" && len(call.Args) == 1 {
			return literalDefaults(call.Args[0], typ.Field(0).Type, cfg)
		}
	}
	if call, ok := expr.(*ast.CallExpr); ok {
		name := nodeText(call.Fun)
		if strings.HasPrefix(name, "cfg.") && len(call.Args) > 0 {
			if s, ok := literalDefaults(call.Args[0], reflect.TypeFor[string](), cfg).(string); ok {
				return cfg.Values[s]
			}
		}
		if name == "showcase.Select" && len(call.Args) == 3 {
			key, _ := literalDefaults(call.Args[1], reflect.TypeFor[string](), cfg).(string)
			if m, ok := call.Args[2].(*ast.CompositeLit); ok {
				for _, e := range m.Elts {
					if kv, ok := e.(*ast.KeyValueExpr); ok {
						if literalDefaults(kv.Key, reflect.TypeFor[string](), cfg) == cfg.Values[key] {
							return literalDefaults(kv.Value, typ, cfg)
						}
					}
				}
			}
		}
		if name == "cmp.Or" && len(call.Args) > 0 {
			for _, arg := range call.Args {
				v := literalDefaults(arg, typ, cfg)
				if v != nil && v != "" && v != false && v != float64(0) {
					return v
				}
			}
			return nil
		}
		if len(call.Args) == 1 && (name == "float32" || name == "float64" || name == "int") {
			return literalDefaults(call.Args[0], typ, cfg)
		}
		if strings.HasPrefix(name, "lucide.") {
			for _, icon := range iconNames() {
				if iconFunction(icon) == name+"()" {
					return icon
				}
			}
		}
	}
	if lit, ok := expr.(*ast.CompositeLit); ok {
		if optional(typ) {
			return nil
		}
		if typ.Kind() == reflect.Struct {
			out := map[string]any{}
			for _, field := range publicFields(typ) {
				if e := literalField(lit, typ, field.Name); e != nil {
					out[field.Name] = literalDefaults(e, field.Type, cfg)
				}
			}
			return out
		}
		if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
			out := []any{}
			for _, e := range lit.Elts {
				out = append(out, literalDefaults(e, typ.Elem(), cfg))
			}
			return out
		}
	}
	if _, ok := expr.(*ast.FuncLit); ok {
		return map[string]any{"mode": "example"}
	}
	return nil
}

func cloneJSON(value any) any {
	b, _ := json.Marshal(value)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}
