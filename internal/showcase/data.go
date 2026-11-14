package showcase

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"

	nui "github.com/josexy/nuxt-ui-go/ui"
)

func (d Demo) dataBindings() []Binding {
	if !slices.Contains([]string{"tree", "command-palette", "chart"}, d.ID) {
		return nil
	}
	prefix := "package main\nfunc example(){\n"
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "setup.go", prefix+d.Setup+"\n}", 0)
	if err != nil {
		panic(err)
	}
	var bindings []Binding
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !slices.Contains([]string{"SetNodes", "NewCommandIndex", "SetCategories", "SetSlices"}, fn.Sel.Name) {
			return true
		}
		for _, arg := range call.Args {
			literal, ok := arg.(*ast.CompositeLit)
			if !ok {
				continue
			}
			typ := expressionType(literal.Type)
			if typ == nil || typ.Kind() != reflect.Slice {
				continue
			}
			name := typ.Elem().Name() + "Items"
			if typ.Elem().Kind() == reflect.String {
				name = "Categories"
			}
			defaults := literalDefaults(literal, typ, Config{})
			bindings = append(bindings, Binding{ID: name, Name: name, Type: goType(typ), Defaults: defaults, typ: typ, literal: literal, start: set.Position(literal.Pos()).Offset - len(prefix), end: set.Position(literal.End()).Offset - len(prefix)})
		}
		return true
	})
	return bindings
}

func (d Demo) configuredSetup(cfg Config, runtime bool, source *valueSource) (string, error) {
	body := d.Setup
	bindings := d.dataBindings()
	slices.SortFunc(bindings, func(a, b Binding) int { return b.start - a.start })
	for _, b := range bindings {
		original := body[b.start:b.end]
		next := original
		if runtime {
			next = "showcase.Options(cfg, " + strconv.Quote(b.ID) + ", " + original + ", nil, nui.Theme{}, &s.Live)"
		} else if patch, ok := cfg.Options[b.ID]; ok {
			var err error
			next, err = overlayExpression(b, original, patch, source)
			if err != nil {
				return "", err
			}
		}
		body = body[:b.start] + next + body[b.end:]
	}
	return body, nil
}

func (r *Runtime) ApplyData(cfg Config, ids []string, apply func() error) error {
	values := map[string]any{}
	for _, id := range ids {
		if v, ok := cfg.Options[id]; ok {
			values[id] = v
		}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return err
	}
	key := string(encoded)
	if r.dataKey == key {
		return nil
	}
	if err := apply(); err != nil {
		return err
	}
	r.dataKey = key
	return nil
}

func (d Demo) validateData(cfg Config) error {
	values := map[string]reflect.Value{}
	for _, b := range d.dataBindings() {
		base := reflect.New(b.typ).Elem()
		if _, err := applyPatch(base, b.Defaults, b.ID, nil, nil, nil, nui.Theme{}); err != nil {
			return err
		}
		if patch, ok := cfg.Options[b.ID]; ok {
			if _, err := applyPatch(base, patch, b.ID, nil, nil, nil, nui.Theme{}); err != nil {
				return err
			}
		}
		values[b.ID] = base
	}
	if d.ID == "chart" {
		var state nui.ChartState
		if err := state.SetCategories(values["Categories"].Interface().([]string), values["ChartSeriesItems"].Interface().([]nui.ChartSeries)); err != nil {
			return fmt.Errorf("ChartSeriesItems: %w", err)
		}
		if err := state.SetSlices(values["ChartSliceItems"].Interface().([]nui.ChartSlice)); err != nil {
			return fmt.Errorf("ChartSliceItems: %w", err)
		}
	}
	if d.ID == "tree" {
		seen := map[string]bool{}
		var visit func([]nui.TreeNode) error
		visit = func(nodes []nui.TreeNode) error {
			for _, n := range nodes {
				if n.ID == "" || seen[n.ID] {
					return fmt.Errorf("TreeNodeItems: IDs must be unique and non-empty")
				}
				seen[n.ID] = true
				if err := visit(n.Children); err != nil {
					return err
				}
			}
			return nil
		}
		if err := visit(values["TreeNodeItems"].Interface().([]nui.TreeNode)); err != nil {
			return err
		}
		if options, ok := cfg.Options["TreeOptions"].(map[string]any); ok {
			if callback, ok := options["OnLoad"].(map[string]any); ok && callback["mode"] == "load" {
				if items, ok := callback["items"]; ok {
					v, _, err := value(reflect.TypeFor[[]nui.TreeNode](), items, "TreeOptions.OnLoad.items", nil, nil, nil, nui.Theme{})
					if err != nil {
						return err
					}
					var validateLazy func([]nui.TreeNode) error
					validateLazy = func(nodes []nui.TreeNode) error {
						for _, node := range nodes {
							if node.Lazy && node.Children == nil {
								if err := visit(prefixTreeNodes(node.ID, v.Interface().([]nui.TreeNode))); err != nil {
									return err
								}
							}
							if err := validateLazy(node.Children); err != nil {
								return err
							}
						}
						return nil
					}
					if err := validateLazy(values["TreeNodeItems"].Interface().([]nui.TreeNode)); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (d Demo) configureMethod(name string) (string, error) {
	bindings := d.dataBindings()
	if len(bindings) == 0 {
		return "func(s *" + name + ") Configure(cfg showcase.Config) error { return nil }\n", nil
	}
	body, err := d.configuredSetup(Config{}, true, nil)
	if err != nil {
		return "", err
	}
	body = strings.ReplaceAll(body, "panic(err)", "return err")
	ids := []string{}
	for _, b := range bindings {
		ids = append(ids, strconv.Quote(b.ID))
	}
	return "func(s *" + name + ") Configure(cfg showcase.Config) error { return s.Live.ApplyData(cfg, []string{" + strings.Join(ids, ",") + "}, func() error { " + body + "; return nil }) }\n", nil
}
