package showcase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	nui "github.com/josexy/nuxt-ui-go"
)

func functionModes(t reflect.Type) []string {
	for i := 0; i < t.NumIn(); i++ {
		if t.In(i) == reflect.TypeFor[ui.Element]() {
			return []string{"style"}
		}
		if t.In(i) == reflect.TypeFor[*ui.Context]() {
			return []string{"content"}
		}
	}
	if t.NumOut() == 0 {
		return []string{"record"}
	}
	if t.NumOut() == 1 {
		switch t.Out(0).Kind() {
		case reflect.Bool:
			return []string{"allow", "deny", "weekends"}
		case reflect.String:
			return []string{"format"}
		}
	}
	return []string{"result"}
}
func fieldFunctionModes(t reflect.Type, name string) []string {
	if name == "Search" {
		return []string{"search"}
	}
	if name == "OnLoad" {
		return []string{"load"}
	}
	if t.NumIn() == 0 && t.NumOut() == 0 && slices.Contains([]string{"Header", "Footer", "Actions", "Content", "Leading", "Trailing", "Body"}, name) {
		return []string{"content"}
	}
	modes := functionModes(t)
	if slices.Contains(modes, "weekends") && (t.NumIn() != 1 || t.In(0) != reflect.TypeFor[nui.Date]()) {
		return []string{"allow", "deny"}
	}
	return modes
}
func functionSignature(t reflect.Type) (string, []string) {
	in, out, names := []string{}, []string{}, []string{}
	for i := 0; i < t.NumIn(); i++ {
		name := fmt.Sprintf("arg%d", i)
		names = append(names, name)
		in = append(in, name+" "+goType(t.In(i)))
	}
	for i := 0; i < t.NumOut(); i++ {
		out = append(out, goType(t.Out(i)))
	}
	signature := "func(" + strings.Join(in, ", ") + ")"
	if len(out) == 1 {
		signature += " " + out[0]
	} else if len(out) > 1 {
		signature += " (" + strings.Join(out, ", ") + ")"
	}
	return signature, names
}

func functionValue(t reflect.Type, spec any, path string, runtime *Runtime, source *valueSource, c *ui.Context, theme nui.Theme) (reflect.Value, string, error) {
	zero := reflect.Zero(t)
	fail := func() (reflect.Value, string, error) { return zero, "", fmt.Errorf("%s: invalid callback", path) }
	m, ok := spec.(map[string]any)
	if !ok {
		return fail()
	}
	mode, _ := m["mode"].(string)
	name := path[strings.LastIndex(path, ".")+1:]
	if !slices.Contains(fieldFunctionModes(t, name), mode) {
		return fail()
	}
	signature, args := functionSignature(t)
	text, _ := m["text"].(string)
	if len(text) > 2000 {
		return fail()
	}
	if mode == "search" || mode == "load" {
		typ := reflect.TypeFor[[]nui.CommandAction]()
		if mode == "load" {
			typ = reflect.TypeFor[[]nui.TreeNode]()
		}
		items := m["items"]
		if items == nil {
			items = []any{}
		}
		if err := validateResolved(typ, items, path+".items"); err != nil {
			return zero, "", err
		}
		v, code, err := value(typ, items, path+".items", runtime, source, c, theme)
		if err != nil {
			return zero, "", err
		}
		if mode == "search" {
			actions := v.Interface().([]nui.CommandAction)
			fn := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
				ctx := args[0].Interface().(context.Context)
				ticket := args[1].Interface().(nui.QueryTicket)
				query, parent := args[2].String(), args[3].String()
				matches := nui.NewCommandIndex(actions).Search(ctx, query, parent, 100)
				runtime.Commands.Results.Complete(ticket, matches, nil)
				runtime.emit(path, args[2:])
				return nil
			})
			return fn, signature + " { s.Commands.Results.Complete(arg1, nui.NewCommandIndex(" + code + ").Search(arg0, arg2, arg3, 100), nil); fmt.Println(" + strconv.Quote(path) + ", arg2, arg3) }", nil
		}
		nodes := v.Interface().([]nui.TreeNode)
		fn := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			runtime.Tree.CompleteLoad(args[1].Interface().(nui.TreeTicket), prefixTreeNodes(args[0].String(), nodes), nil)
			runtime.emit(path, args[:1])
			return nil
		})
		return fn, signature + " { var children func(string, []nui.TreeNode) []nui.TreeNode; children = func(parent string, nodes []nui.TreeNode) []nui.TreeNode { if nodes == nil { return nil }; out := append([]nui.TreeNode{}, nodes...); for i := range out { out[i].ID = parent + \"/\" + out[i].ID; out[i].Children = children(out[i].ID, out[i].Children) }; return out }; s.Tree.CompleteLoad(arg1, children(arg0, " + code + "), nil); fmt.Println(" + strconv.Quote(path) + ", arg0) }", nil
	}
	// Event functions capture only the persistent event sink, never a Context.
	if mode == "record" {
		return reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
				if runtime != nil {
					runtime.emit(path, args)
				}
				return nil
			}), signature + " { fmt.Println(" + strconv.Quote(path) + func() string {
				if len(args) > 0 {
					return ", " + strings.Join(args, ", ")
				}
				return ""
			}() + ") }", nil
	}
	if mode == "allow" || mode == "deny" || mode == "weekends" {
		expr := strconv.FormatBool(mode == "allow")
		if mode == "weekends" {
			expr = "time.Date(arg0.Year, arg0.Month, arg0.Day, 0, 0, 0, 0, time.UTC).Weekday() == time.Saturday || time.Date(arg0.Year, arg0.Month, arg0.Day, 0, 0, 0, 0, time.UTC).Weekday() == time.Sunday"
		}
		return reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			result := mode == "allow"
			if mode == "weekends" {
				d := args[0].Interface().(nui.Date)
				day := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Weekday()
				result = day == time.Saturday || day == time.Sunday
			}
			return []reflect.Value{reflect.ValueOf(result)}
		}), signature + " { return " + expr + " }", nil
	}
	if mode == "format" {
		suffix, _ := m["suffix"].(string)
		precision := float64(2)
		if raw, ok := m["precision"]; ok {
			var valid bool
			precision, valid = number(raw)
			if !valid || precision < 0 || precision > 12 || precision != float64(int(precision)) {
				return fail()
			}
		}
		layout, _ := m["layout"].(string)
		if layout == "" {
			layout = "2006-01-02"
		}
		format := "%." + strconv.Itoa(int(precision)) + "f"
		expression := strconv.Quote(text) + " + fmt.Sprint(arg0) + " + strconv.Quote(suffix)
		isTime := t.In(0) == reflect.TypeFor[time.Time]()
		isNumber := t.In(0).Kind() == reflect.Float32 || t.In(0).Kind() == reflect.Float64
		if isTime {
			expression = strconv.Quote(text) + " + arg0.Format(" + strconv.Quote(layout) + ") + " + strconv.Quote(suffix)
		} else if isNumber {
			expression = strconv.Quote(text) + " + fmt.Sprintf(" + strconv.Quote(format) + ", arg0) + " + strconv.Quote(suffix)
		}
		fn := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			v := fmt.Sprint(args[0].Interface())
			if isTime {
				v = args[0].Interface().(time.Time).Format(layout)
			} else if isNumber {
				v = fmt.Sprintf(format, args[0].Interface())
			}
			return []reflect.Value{reflect.ValueOf(text + v + suffix)}
		})
		return fn, signature + " { return " + expression + " }", nil
	}
	if mode == "content" {
		contextArg := -1
		for i := 0; i < t.NumIn(); i++ {
			if t.In(i) == reflect.TypeFor[*ui.Context]() {
				contextArg = i
			}
		}
		contextName := "c"
		if contextArg >= 0 {
			contextName = args[contextArg]
		}
		expr := "ui.Text(" + contextName + ".Key(" + strconv.Quote("configured-"+path) + "), " + strconv.Quote(text) + ")"
		statement := expr
		if t.NumOut() > 0 {
			statement = "return " + expr
		}
		// Builder callbacks run synchronously during construction. c is captured
		// only for these documented build callbacks, never stored in Runtime.
		fn := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			ctx := c
			if contextArg >= 0 {
				ctx = args[contextArg].Interface().(*ui.Context)
			}
			element := ui.Text(ctx.Key("configured-"+path), text)
			if t.NumOut() > 0 {
				return []reflect.Value{reflect.ValueOf(element)}
			}
			return nil
		})
		return fn, signature + " { " + statement + " }", nil
	}
	if mode == "style" {
		raw := m["parts"]
		if raw == nil {
			raw = map[string]any{}
		}
		v, code, err := value(reflect.TypeFor[nui.Styles](), raw, path+".parts", runtime, source, c, theme)
		if err != nil {
			return zero, "", err
		}
		styles := v.Interface().(nui.Styles)
		slots := []nui.Slot{nui.Root, nui.Label, nui.Leading, nui.Trailing}
		if strings.Contains(path, "Input") || strings.Contains(path, "Textarea") {
			slots[1] = nui.Editor
		}
		var lines []string
		if t.NumIn() == 2 && t.In(0) == reflect.TypeFor[nui.Slot]() {
			lines = append(lines, styleStatements("arg1", "styles[arg0]")...)
		} else {
			for i := range t.NumIn() {
				if t.In(i) == reflect.TypeFor[ui.Element]() {
					lines = append(lines, styleStatements(args[i], "styles[nui.Slot("+strconv.Quote(string(slots[i]))+")]")...)
				}
			}
		}
		fn := reflect.MakeFunc(t, func(args []reflect.Value) []reflect.Value {
			if t.NumIn() == 2 && t.In(0) == reflect.TypeFor[nui.Slot]() {
				applyPartStyle(args[1].Interface().(ui.Element), styles[args[0].Interface().(nui.Slot)])
			} else {
				for i, arg := range args {
					if t.In(i) == reflect.TypeFor[ui.Element]() {
						applyPartStyle(arg.Interface().(ui.Element), styles[slots[i]])
					}
				}
			}
			return nil
		})
		return fn, signature + " { styles := " + code + "; " + strings.Join(lines, "; ") + " }", nil
	}
	// Return values use the same typed editors as ordinary fields, including
	// FormErrors and FileUpload's worker result. Errors use a message string.
	results, _ := m["results"].([]any)
	if len(results) != t.NumOut() {
		return fail()
	}
	values := make([]reflect.Value, t.NumOut())
	codes := make([]string, t.NumOut())
	for i := range t.NumOut() {
		if t.Out(i) == reflect.TypeFor[error]() {
			s, _ := results[i].(string)
			values[i] = reflect.Zero(t.Out(i))
			codes[i] = "nil"
			if s != "" {
				values[i] = reflect.ValueOf(errors.New(s))
				codes[i] = "errors.New(" + strconv.Quote(s) + ")"
			}
			continue
		}
		v, code, err := value(t.Out(i), results[i], fmt.Sprintf("%s.results[%d]", path, i), runtime, source, c, theme)
		if err != nil {
			return zero, "", err
		}
		values[i], codes[i] = v, code
	}
	return reflect.MakeFunc(t, func([]reflect.Value) []reflect.Value {
		out := make([]reflect.Value, len(values))
		for i, v := range values {
			out[i] = copyResult(v)
		}
		return out
	}), signature + " { return " + strings.Join(codes, ", ") + " }", nil
}

func prefixTreeNodes(parent string, nodes []nui.TreeNode) []nui.TreeNode {
	if nodes == nil {
		return nil
	}
	out := slices.Clone(nodes)
	for i := range out {
		out[i].ID = parent + "/" + out[i].ID
		out[i].Children = prefixTreeNodes(out[i].ID, out[i].Children)
	}
	return out
}

// Return fresh collection values on every callback invocation, just like the
// literal in copied Go code. Worker consumers must not share mutable slices.
func copyResult(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Slice:
		if v.IsNil() {
			return v
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := range v.Len() {
			out.Index(i).Set(copyResult(v.Index(i)))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return v
		}
		out := reflect.MakeMap(v.Type())
		iter := v.MapRange()
		for iter.Next() {
			out.SetMapIndex(iter.Key(), copyResult(iter.Value()))
		}
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v)
		for _, f := range publicFields(v.Type()) {
			out.FieldByIndex(f.Index).Set(copyResult(v.FieldByIndex(f.Index)))
		}
		return out
	}
	return v
}

// The statements mirror the library's public PartStyle fields, and are also
// emitted into standalone examples so copying requires no website helper.
func styleStatements(element, style string) []string {
	var lines []string
	for _, pair := range [][2]string{{"Background", "Background"}, {"Foreground", "TextColor"}, {"Radius", "Radius"}, {"Opacity", "Opacity"}, {"Font", "Font"}, {"FontSize", "FontSize"}, {"LineHeight", "LineHeight"}, {"Weight", "FontWeight"}, {"PaddingX", "PaddingX"}, {"PaddingY", "PaddingY"}, {"Gap", "Gap"}} {
		lines = append(lines, "if "+style+"."+pair[0]+".Set { "+element+"."+pair[1]+"("+style+"."+pair[0]+".Value) }")
	}
	lines = append(lines, "if "+style+".BorderWidth.Set { "+element+".InsetShadow(0, 0, 0, "+style+".BorderWidth.Value, "+style+".BorderColor.Value) }", "if "+style+".Shadow.Set { shadow := "+style+".Shadow.Value; "+element+".Shadow(0, shadow.OffsetY, shadow.Blur, 0, shadow.Color) }")
	return []string{"if " + element + ".Valid() { " + strings.Join(lines, "; ") + " }"}
}
func applyPartStyle(e ui.Element, s nui.PartStyle) {
	// Optional slots (for example an absent leading icon) have no element.
	if !e.Valid() {
		return
	}
	if s.Background.Set {
		e.Background(s.Background.Value)
	}
	if s.Foreground.Set {
		e.TextColor(s.Foreground.Value)
	}
	if s.Radius.Set {
		e.Radius(s.Radius.Value)
	}
	if s.Opacity.Set {
		e.Opacity(s.Opacity.Value)
	}
	if s.Font.Set {
		e.Font(s.Font.Value)
	}
	if s.FontSize.Set {
		e.FontSize(s.FontSize.Value)
	}
	if s.LineHeight.Set {
		e.LineHeight(s.LineHeight.Value)
	}
	if s.Weight.Set {
		e.FontWeight(s.Weight.Value)
	}
	if s.PaddingX.Set {
		e.PaddingX(s.PaddingX.Value)
	}
	if s.PaddingY.Set {
		e.PaddingY(s.PaddingY.Value)
	}
	if s.Gap.Set {
		e.Gap(s.Gap.Value)
	}
	if s.BorderWidth.Set {
		e.InsetShadow(0, 0, 0, s.BorderWidth.Value, s.BorderColor.Value)
	}
	if s.Shadow.Set {
		v := s.Shadow.Value
		e.Shadow(0, v.OffsetY, v.Blur, 0, v.Color)
	}
}
