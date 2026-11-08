package showcase

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
	"github.com/josexy/nuxt-ui-go/icons/lucide"
)

var icons = map[string]func() *ui.SVG{
	"check": lucide.Check, "plus": lucide.Plus, "search": lucide.Search, "settings": lucide.Settings,
	"x": lucide.X, "loader-circle": lucide.LoaderCircle, "chevron-down": lucide.ChevronDown, "chevron-up": lucide.ChevronUp,
	"chevron-left": lucide.ChevronLeft, "chevron-right": lucide.ChevronRight, "arrow-left": lucide.ArrowLeft, "arrow-right": lucide.ArrowRight,
	"info": lucide.Info, "circle-alert": lucide.CircleAlert, "folder": lucide.Folder, "file": lucide.File, "house": lucide.House,
	"user": lucide.User, "star": lucide.Star, "heart": lucide.Heart, "mail": lucide.Mail, "calendar": lucide.Calendar, "clock": lucide.Clock,
	"upload": lucide.Upload, "download": lucide.Download, "trash": lucide.Trash, "pencil": lucide.Pencil, "copy": lucide.Copy,
}

func iconNames() []string {
	names := make([]string, 0, len(icons))
	for name := range icons {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
func iconFunction(name string) string {
	var parts []string
	for p := range strings.SplitSeq(name, "-") {
		parts = append(parts, strings.ToUpper(p[:1])+p[1:])
	}
	return "lucide." + strings.Join(parts, "") + "()"
}

type cachedValue struct {
	key   string
	value reflect.Value
}

// Runtime owns persistent pointers only; build-scoped Context/Element values
// are never retained. Event sinks may be invoked from worker callbacks.
type Runtime struct {
	values   map[string]cachedValue
	Focus    ui.Handle
	Commands *nui.CommandPaletteState
	Tree     *nui.TreeState
	Window   *mygo.Window
	dataKey  string
	mu       sync.Mutex
	report   func(string, []string)
}

func (r *Runtime) SetReport(fn func(string, []string)) { r.mu.Lock(); r.report = fn; r.mu.Unlock() }
func (r *Runtime) emit(name string, args []reflect.Value) {
	values := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.CanInterface() {
			values = append(values, fmt.Sprint(arg.Interface()))
		}
	}
	r.mu.Lock()
	report := r.report
	r.mu.Unlock()
	if report != nil {
		report(name, values)
	} else {
		fmt.Println(name, strings.Join(values, ", "))
	}
}
func (r *Runtime) persistent(path string, spec any, create func() (reflect.Value, error)) (reflect.Value, error) {
	data, _ := json.Marshal(spec)
	key := string(data)
	if r.values == nil {
		r.values = map[string]cachedValue{}
	}
	if current, ok := r.values[path]; ok && current.key == key {
		return current.value, nil
	}
	value, err := create()
	if err == nil {
		r.values[path] = cachedValue{key, value}
	}
	return value, err
}

func (r *Runtime) Begin(c *ui.Context, t nui.Theme, cfg Config) {
	if cfg.focusTarget {
		nui.Button(c.Key("preview-return-focus"), t, "Focus target", nui.ButtonOptions{Variant: nui.Outline}).AlignSelf(ui.Start).Bind(&r.Focus)
	}
}

type sourcePointer struct{ name, typ, init string }
type valueSource struct {
	pointers map[string]sourcePointer
	focus    bool
	window   bool
}

func (s *valueSource) pointer(path, typ, init string) string {
	if s.pointers == nil {
		s.pointers = map[string]sourcePointer{}
	}
	name := "Option" + strconv.Itoa(len(s.pointers)+1)
	if p, ok := s.pointers[path]; ok {
		return "&s." + p.name
	}
	s.pointers[path] = sourcePointer{name, typ, init}
	return "&s." + name
}

func number(value any) (float64, bool) {
	n, ok := value.(float64)
	return n, ok && !math.IsNaN(n) && !math.IsInf(n, 0)
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)

// value converts the same JSON representation to a typed Go value and a Go
// expression. Keeping both together makes preview and copied code agree.
func value(typ reflect.Type, spec any, path string, runtime *Runtime, source *valueSource, c *ui.Context, theme nui.Theme) (reflect.Value, string, error) {
	zero := reflect.Zero(typ)
	fail := func() (reflect.Value, string, error) {
		return zero, "", fmt.Errorf("%s: invalid %s", path, goType(typ))
	}
	if spec == nil {
		return zero, zeroExpression(typ), nil
	}
	if typ == reflect.TypeFor[*mygo.Window]() {
		if spec != "current" {
			return fail()
		}
		if source != nil {
			source.window = true
		}
		if runtime == nil {
			return reflect.Zero(typ), "s.Window", nil
		}
		return reflect.ValueOf(runtime.Window), "s.Window", nil
	}
	if typ == reflect.TypeFor[ui.Color]() {
		s, ok := spec.(string)
		if !ok || !hexColor.MatchString(s) {
			return fail()
		}
		return reflect.ValueOf(ui.Hex(s)), "ui.Hex(" + strconv.Quote(s) + ")", nil
	}
	if typ == reflect.TypeFor[nui.Date]() {
		s, ok := spec.(string)
		if !ok {
			return fail()
		}
		date, err := time.Parse("2006-01-02", s)
		if err != nil {
			return fail()
		}
		d := nui.Date{Year: date.Year(), Month: date.Month(), Day: date.Day()}
		return reflect.ValueOf(d), fmt.Sprintf("nui.Date{Year: %d, Month: time.Month(%d), Day: %d}", d.Year, d.Month, d.Day), nil
	}
	if typ == reflect.TypeFor[time.Duration]() {
		n, ok := number(spec)
		if !ok || math.Abs(n) > float64(24*time.Hour/time.Millisecond)*365 {
			return fail()
		}
		duration := time.Duration(n * float64(time.Millisecond))
		code := fmt.Sprintf("time.Duration(%d)", duration)
		if duration%time.Millisecond == 0 {
			code = fmt.Sprintf("%d * time.Millisecond", duration/time.Millisecond)
		}
		return reflect.ValueOf(duration), code, nil
	}
	if optional(typ) {
		v, code, err := value(typ.Field(0).Type, spec, path, runtime, source, c, theme)
		if err != nil {
			return zero, "", err
		}
		out := reflect.New(typ).Elem()
		out.FieldByName("Value").Set(v)
		out.FieldByName("Set").SetBool(true)
		return out, "nui.Set(" + code + ")", nil
	}
	if typ == reflect.TypeFor[ui.Modifiers]() {
		items, ok := spec.([]any)
		if !ok {
			return fail()
		}
		mods := ui.Modifiers(0)
		var expr []string
		choices := map[string]ui.Modifiers{"ui.Ctrl": ui.Ctrl, "ui.Shift": ui.Shift, "ui.Alt": ui.Alt, "ui.Super": ui.Super}
		for _, it := range items {
			name, ok := it.(string)
			mod, valid := choices[name]
			if !ok || !valid {
				return fail()
			}
			mods |= mod
			expr = append(expr, name)
		}
		code := "ui.Modifiers(0)"
		if len(expr) > 0 {
			code = strings.Join(expr, " | ")
		}
		return reflect.ValueOf(mods), code, nil
	}
	if choices := enumValues[typ]; len(choices) > 0 {
		s, ok := spec.(string)
		if !ok {
			return fail()
		}
		for _, choice := range choices {
			if choice.Name == s {
				return choice.Value, choice.Name, nil
			}
		}
		return fail()
	}
	if typ == reflect.TypeFor[*ui.SVG]() {
		s, ok := spec.(string)
		if !ok {
			return fail()
		}
		if s == "" || s == "none" {
			return zero, "nil", nil
		}
		if icon, ok := icons[s]; ok {
			return reflect.ValueOf(icon()), iconFunction(s), nil
		}
		if len(s) > 65536 || !strings.Contains(s, "<svg") {
			return fail()
		}
		v, err := ui.ParseSVG([]byte(s))
		if err != nil {
			return zero, "", fmt.Errorf("%s: %w", path, err)
		}
		init := "func() *ui.SVG { icon, err := ui.ParseSVG([]byte(" + strconv.Quote(s) + ")); if err != nil { panic(err) }; return icon }()"
		code := init
		if source != nil {
			code = "*" + source.pointer(path, "*ui.SVG", init)
		}
		if runtime != nil {
			cached, err := runtime.persistent(path, s, func() (reflect.Value, error) { return reflect.ValueOf(v), nil })
			return cached, code, err
		}
		return reflect.ValueOf(v), code, nil
	}
	if typ == reflect.TypeFor[*ui.Bitmap]() {
		s, ok := spec.(string)
		if !ok || len(s) > 3*1024*1024 {
			return fail()
		}
		_, encoded, ok := strings.Cut(s, ",")
		if !ok || !strings.HasPrefix(s, "data:image/") {
			return fail()
		}
		bytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fail()
		}
		create := func() (reflect.Value, error) {
			bitmap, err := ui.DecodeBitmap(bytes)
			if err != nil {
				return zero, err
			}
			w, h := bitmap.Size()
			if w > 4096 || h > 4096 {
				return zero, fmt.Errorf("image exceeds 4096 pixels")
			}
			return reflect.ValueOf(bitmap), nil
		}
		var image reflect.Value
		if runtime != nil {
			image, err = runtime.persistent(path, s, create)
		} else {
			image, err = create()
		}
		if err != nil {
			return zero, "", fmt.Errorf("%s: %w", path, err)
		}
		init := "func() *ui.Bitmap { data, err := base64.StdEncoding.DecodeString(" + strconv.Quote(encoded) + "); if err != nil { panic(err) }; image, err := ui.DecodeBitmap(data); if err != nil { panic(err) }; return image }()"
		code := init
		if source != nil {
			code = "*" + source.pointer(path, "*ui.Bitmap", init)
		}
		return image, code, nil
	}
	if typ == reflect.TypeFor[*ui.Handle]() {
		if spec != "target" {
			return fail()
		}
		if source != nil {
			source.focus = true
		}
		if runtime == nil {
			runtime = &Runtime{}
		}
		return reflect.ValueOf(&runtime.Focus), "&s.ReturnFocus", nil
	}
	if typ == reflect.TypeFor[*ui.ScrollState]() {
		m, ok := spec.(map[string]any)
		if !ok {
			return fail()
		}
		x, y := float64(0), float64(0)
		for key, raw := range m {
			n, ok := number(raw)
			if !ok || n < 0 {
				return fail()
			}
			switch key {
			case "X":
				x = n
			case "Y":
				y = n
			default:
				return fail()
			}
		}
		init := fmt.Sprintf("ui.ScrollState{X: %s, Y: %s}", strconv.FormatFloat(x, 'f', -1, 64), strconv.FormatFloat(y, 'f', -1, 64))
		create := func() (reflect.Value, error) {
			return reflect.ValueOf(&ui.ScrollState{X: float32(x), Y: float32(y)}), nil
		}
		v, err := create()
		if runtime != nil {
			v, err = runtime.persistent(path, spec, create)
		}
		code := "&" + init
		if source != nil {
			code = source.pointer(path, "ui.ScrollState", init)
		}
		return v, code, err
	}
	if typ.Kind() == reflect.Func {
		return functionValue(typ, spec, path, runtime, source, c, theme)
	}
	switch typ.Kind() {
	case reflect.String:
		s, ok := spec.(string)
		if !ok || len(s) > 65536 {
			return fail()
		}
		v := reflect.New(typ).Elem()
		v.SetString(s)
		return v, typedExpression(typ, strconv.Quote(s)), nil
	case reflect.Bool:
		b, ok := spec.(bool)
		if !ok {
			return fail()
		}
		return reflect.ValueOf(b).Convert(typ), strconv.FormatBool(b), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := number(spec)
		v := reflect.New(typ).Elem()
		if !ok || math.Trunc(n) != n || math.Abs(n) > 9007199254740991 || v.OverflowInt(int64(n)) {
			return fail()
		}
		v.SetInt(int64(n))
		return v, typedExpression(typ, strconv.FormatFloat(n, 'f', -1, 64)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := number(spec)
		v := reflect.New(typ).Elem()
		if !ok || n < 0 || math.Trunc(n) != n || n > 9007199254740991 || v.OverflowUint(uint64(n)) {
			return fail()
		}
		v.SetUint(uint64(n))
		return v, typedExpression(typ, strconv.FormatFloat(n, 'f', -1, 64)), nil
	case reflect.Float32, reflect.Float64:
		n, ok := number(spec)
		v := reflect.New(typ).Elem()
		if !ok || v.OverflowFloat(n) {
			return fail()
		}
		v.SetFloat(n)
		return v, typedExpression(typ, strconv.FormatFloat(n, 'f', -1, 64)), nil
	case reflect.Pointer:
		v, code, err := value(typ.Elem(), spec, path, runtime, source, c, theme)
		if err != nil {
			return zero, "", err
		}
		ptr := reflect.New(typ.Elem())
		ptr.Elem().Set(v)
		expr := "func() " + goType(typ) + " { v := " + typedExpression(typ.Elem(), code) + "; return &v }()"
		if typ.Elem().Kind() == reflect.Bool {
			if runtime != nil {
				ptr, err = runtime.persistent(path, spec, func() (reflect.Value, error) { return ptr, nil })
			}
			if source != nil {
				expr = source.pointer(path, goType(typ.Elem()), code)
			}
		}
		return ptr, expr, err
	case reflect.Struct:
		m, ok := spec.(map[string]any)
		if !ok {
			return fail()
		}
		out := reflect.New(typ).Elem()
		var assignments []string
		keys := sortedKeys(m)
		for _, key := range keys {
			field, ok := typ.FieldByName(key)
			if !ok || !field.IsExported() || field.Anonymous {
				return fail()
			}
			v, code, err := value(field.Type, m[key], path+"."+key, runtime, source, c, theme)
			if err != nil {
				return zero, "", err
			}
			out.FieldByIndex(field.Index).Set(v)
			assignments = append(assignments, "v."+key+" = "+code)
		}
		code := goType(typ) + "{}"
		if len(assignments) > 0 {
			code = "func() " + goType(typ) + " { v := " + goType(typ) + "{}; " + strings.Join(assignments, "; ") + "; return v }()"
		}
		return out, code, nil
	case reflect.Array, reflect.Slice:
		items, ok := spec.([]any)
		if !ok || len(items) > 100 {
			return fail()
		}
		if typ.Kind() == reflect.Array && len(items) != typ.Len() {
			return fail()
		}
		var out reflect.Value
		if typ.Kind() == reflect.Array {
			out = reflect.New(typ).Elem()
		} else {
			out = reflect.MakeSlice(typ, len(items), len(items))
		}
		codes := make([]string, len(items))
		for i, item := range items {
			v, code, err := value(typ.Elem(), item, fmt.Sprintf("%s[%d]", path, i), runtime, source, c, theme)
			if err != nil {
				return zero, "", err
			}
			out.Index(i).Set(v)
			codes[i] = code
		}
		return out, goType(typ) + "{" + strings.Join(codes, ",") + "}", nil
	case reflect.Map:
		m, ok := spec.(map[string]any)
		if !ok || len(m) > 100 {
			return fail()
		}
		out := reflect.MakeMap(typ)
		var codes []string
		for _, key := range sortedKeys(m) {
			if typ.Key().Kind() != reflect.String {
				return fail()
			}
			k := reflect.New(typ.Key()).Elem()
			k.SetString(key)
			v, code, err := value(typ.Elem(), m[key], path+"."+key, runtime, source, c, theme)
			if err != nil {
				return zero, "", err
			}
			out.SetMapIndex(k, v)
			codes = append(codes, typedExpression(typ.Key(), strconv.Quote(key))+":"+code)
		}
		return out, goType(typ) + "{" + strings.Join(codes, ",") + "}", nil
	}
	return fail()
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
func zeroExpression(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Func, reflect.Interface:
		return "nil"
	case reflect.String:
		return typedExpression(t, `""`)
	case reflect.Bool:
		return "false"
	case reflect.Struct, reflect.Array:
		return goType(t) + "{}"
	default:
		return typedExpression(t, "0")
	}
}
func typedExpression(t reflect.Type, expr string) string {
	if t == reflect.TypeFor[string]() || t == reflect.TypeFor[int]() {
		return expr
	}
	switch t.Kind() {
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.String:
		if t.Name() != "" {
			return goType(t) + "(" + expr + ")"
		}
	}
	return expr
}

// Apply overlays only explicitly supplied fields. Nested objects preserve the
// example's existing defaults; null and explicit false/zero remain distinct.
func applyPatch(target reflect.Value, spec any, path string, runtime *Runtime, source *valueSource, c *ui.Context, theme nui.Theme) ([]string, error) {
	return patchValue(target, spec, path, "o", runtime, source, c, theme)
}

func patchValue(target reflect.Value, spec any, path, expression string, runtime *Runtime, source *valueSource, c *ui.Context, theme nui.Theme) ([]string, error) {
	if m, ok := spec.(map[string]any); ok && target.Kind() == reflect.Struct && !optional(target.Type()) && target.Type() != reflect.TypeFor[ui.Color]() && target.Type() != reflect.TypeFor[nui.Date]() {
		var lines []string
		for _, key := range sortedKeys(m) {
			field, ok := target.Type().FieldByName(key)
			if !ok || !field.IsExported() || field.Anonymous {
				return nil, fmt.Errorf("%s.%s: unknown field", path, key)
			}
			child := target.FieldByIndex(field.Index)
			statements, err := patchValue(child, m[key], path+"."+key, expression+"."+key, runtime, source, c, theme)
			if err != nil {
				return nil, err
			}
			lines = append(lines, statements...)
		}
		return lines, nil
	}
	if items, ok := spec.([]any); ok && target.Kind() == reflect.Slice && target.Type().Elem().Kind() == reflect.Struct {
		if len(items) > 100 {
			return nil, fmt.Errorf("%s: at most 100 items", path)
		}
		result := reflect.MakeSlice(target.Type(), len(items), len(items))
		var lines []string
		lines = append(lines, fmt.Sprintf("result := make(%s, %d)", goType(target.Type()), len(items)))
		_, hasID := target.Type().Elem().FieldByName("ID")
		for i, item := range items {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			targetExpr := fmt.Sprintf("result[%d]", i)
			// Keyed data inherits callbacks only from the same business item.
			// A replacement at the same index must not activate the old item.
			matchedID := false
			if m, ok := item.(map[string]any); ok && hasID {
				if id, ok := m["ID"].(string); ok && id != "" {
					matchedID = true
					itemPath = path + "[" + id + "]"
					for j := range target.Len() {
						if target.Index(j).FieldByName("ID").String() == id {
							result.Index(i).Set(target.Index(j))
							break
						}
					}
					lines = append(lines, "for _, previous := range base { if previous.ID == "+strconv.Quote(id)+" { "+targetExpr+" = previous; break } }")
				}
			}
			if !matchedID {
				lines = append(lines, fmt.Sprintf("if len(base) > %d { %s = base[%d] }", i, targetExpr, i))
				if i < target.Len() {
					result.Index(i).Set(target.Index(i))
				}
			}
			statements, err := patchValue(result.Index(i), item, itemPath, targetExpr, runtime, source, c, theme)
			if err != nil {
				return nil, err
			}
			lines = append(lines, statements...)
		}
		target.Set(result)
		code := "func(base " + goType(target.Type()) + ") " + goType(target.Type()) + " { " + strings.Join(lines, "; ") + "; return result }(" + expression + ")"
		return []string{expression + " = " + code}, nil
	}
	if callback, ok := spec.(map[string]any); ok && target.Kind() == reflect.Func && callback["mode"] == "example" {
		return nil, nil
	}
	v, code, err := value(target.Type(), spec, path, runtime, source, c, theme)
	if err != nil {
		return nil, err
	}
	if callback, ok := spec.(map[string]any); ok && target.Kind() == reflect.Func && callback["mode"] == "record" {
		previous := reflect.ValueOf(target.Interface())
		next := v
		v = reflect.MakeFunc(target.Type(), func(args []reflect.Value) []reflect.Value {
			if !previous.IsNil() {
				previous.Call(args)
			}
			return next.Call(args)
		})
		signature, args := functionSignature(target.Type())
		code = "func(previous " + goType(target.Type()) + ") " + goType(target.Type()) + " { next := " + code + "; return " + signature + " { if previous != nil { previous(" + strings.Join(args, ", ") + ") }; next(" + strings.Join(args, ", ") + ") } }(" + expression + ")"
	}
	target.Set(v)
	return []string{expression + " = " + code}, nil
}

func Options[T any](cfg Config, id string, base T, c *ui.Context, t nui.Theme, runtime *Runtime) T {
	patch, ok := cfg.Options[id]
	if !ok {
		return base
	}
	value := reflect.ValueOf(&base).Elem()
	if _, err := applyPatch(value, patch, id, runtime, nil, c, t); err != nil {
		panic(err)
	} // Normalize rejects invalid values before the UI thread.
	return base
}
