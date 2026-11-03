package showcase

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	nui "github.com/josexy/nuxt-ui-go"
)

func (d Demo) Normalize(input Config) (Config, error) {
	cfg, err := d.normalizeValues(input)
	if err != nil {
		return Config{}, err
	}
	cfg.Options = map[string]any{}
	bindings := d.Bindings()
	source := &valueSource{}
	for _, id := range sortedKeys(input.Options) {
		index := slices.IndexFunc(bindings, func(b Binding) bool { return b.ID == id })
		if index < 0 {
			return Config{}, fmt.Errorf("unknown options %s", id)
		}
		b := bindings[index]
		patch := cloneJSON(input.Options[id])
		if err := validateFields(b.typ, patch, id); err != nil {
			return Config{}, err
		}
		target := reflect.New(b.typ).Elem()
		if _, err := applyPatch(target, patch, id, nil, source, nil, nui.Theme{}); err != nil {
			return Config{}, err
		}
		defaults := literalDefaults(b.literal, b.typ, cfg)
		if d.ID == "button" && b.ID == "ButtonOptions" && cfg.String("label") == "" {
			if m, ok := mergePatch(defaults, patch).(map[string]any); ok {
				if label, _ := m["AccessibleLabel"].(string); label == "" {
					return Config{}, fmt.Errorf("ButtonOptions.AccessibleLabel is required for an icon-only button")
				}
			}
		}
		if err := validateResolved(b.typ, mergePatch(defaults, patch), id); err != nil {
			return Config{}, err
		}
		cfg.Options[id] = patch
	}
	cfg.focusTarget = source.focus
	if err := d.validateData(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Reject unsupported enum combinations before a component can panic. More
// specific numeric rules belong here instead of silently changing user input.
func validateFields(t reflect.Type, spec any, path string) error {
	if spec == nil {
		return nil
	}
	if optional(t) {
		return validateFields(t.Field(0).Type, spec, path)
	}
	if t.Kind() == reflect.Struct {
		if m, ok := spec.(map[string]any); ok {
			for _, key := range sortedKeys(m) {
				field, ok := t.FieldByName(key)
				if !ok || !field.IsExported() || field.Anonymous {
					return fmt.Errorf("%s.%s: unknown field", path, key)
				}
				raw := m[key]
				if key == "FilterFields" {
					if items, ok := raw.([]any); ok {
						allowed := []string{"label", "description", "value"}
						if t.Name() == "SelectMenuOptions" {
							allowed = append(allowed, "id", "group")
						}
						for _, item := range items {
							v, ok := item.(string)
							if !ok || !slices.Contains(allowed, v) {
								return fmt.Errorf("%s.%s: unsupported filter field", path, key)
							}
						}
					}
				}
				if field.Type == reflect.TypeFor[nui.Variant]() && raw != nil {
					v, ok := raw.(string)
					if !ok || !slices.Contains(enumNames(field.Type, t.Name(), key), v) {
						return fmt.Errorf("%s.%s: unsupported variant", path, key)
					}
				}
				if n, ok := number(raw); ok {
					if slices.Contains([]string{"Width", "Height", "MinWidth", "MaxWidth", "MinHeight", "MaxHeight", "Gap", "Padding", "PaddingX", "PaddingY", "Radius", "BorderWidth", "FontSize", "LineHeight", "ThumbSize", "IconSize", "Step", "Duration", "Delay", "Interval", "MaxFiles", "MaxSize", "MaxLength", "Length", "Count", "Rows", "Columns", "PageSize", "VisibleCount", "Overscan"}, key) && n < 0 {
						return fmt.Errorf("%s.%s must be non-negative", path, key)
					}
					if (key == "Opacity" && (n < 0 || n > 1)) || (key == "Length" && t.Name() == "PinInputOptions" && n > 32) || (key == "Max" && t.Name() == "InputRatingOptions" && n > 20) {
						return fmt.Errorf("%s.%s is out of range", path, key)
					}
				}
				if err := validateFields(field.Type, raw, path+"."+key); err != nil {
					return err
				}
			}
		}
	}
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		if items, ok := spec.([]any); ok {
			for i, v := range items {
				if err := validateFields(t.Elem(), v, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}
	if t.Kind() == reflect.Map {
		if m, ok := spec.(map[string]any); ok {
			for key, v := range m {
				if err := validateFields(t.Elem(), v, path+"."+key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func mergePatch(base, patch any) any {
	m, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	result := map[string]any{}
	if defaults, ok := base.(map[string]any); ok {
		for k, v := range defaults {
			result[k] = v
		}
	}
	for k, v := range m {
		result[k] = mergePatch(result[k], v)
	}
	return result
}

func validateResolved(t reflect.Type, spec any, path string) error {
	if spec == nil {
		return nil
	}
	if optional(t) {
		return validateResolved(t.Field(0).Type, spec, path)
	}
	if t.Kind() == reflect.Struct {
		m, ok := spec.(map[string]any)
		if !ok {
			return nil
		}
		num := func(key string, fallback float64) float64 {
			if n, ok := number(m[key]); ok {
				return n
			}
			return fallback
		}
		switch t.Name() {
		case "SliderOptions", "RangeSliderOptions":
			if num("Max", 100) <= num("Min", 0) {
				return fmt.Errorf("%s: Max must exceed Min", path)
			}
		case "InputNumberOptions", "SplitterOptions":
			if num("Max", 1e7) < num("Min", 0) {
				return fmt.Errorf("%s: Max must not be less than Min", path)
			}
			if t.Name() == "SplitterOptions" && m["Percent"] == true && num("Min", 0) > 100 {
				return fmt.Errorf("%s: percentage Min exceeds 100", path)
			}
		case "CalendarOptions":
			if num("WeekStart", 0) < 0 || num("WeekStart", 0) > 6 {
				return fmt.Errorf("%s: WeekStart must be 0..6", path)
			}
			lo, _ := m["Min"].(string)
			hi, _ := m["Max"].(string)
			if lo != "" && hi != "" && lo > hi {
				return fmt.Errorf("%s: Max date precedes Min", path)
			}
		case "DrawerOptions":
			if points, ok := m["SnapPoints"].([]any); ok {
				previous := float64(0)
				for _, p := range points {
					n, ok := number(p)
					if !ok || n <= previous || n > 1 {
						return fmt.Errorf("%s: SnapPoints must increase within (0, 1]", path)
					}
					previous = n
				}
			}
		case "ProgressOptions":
			if max, ok := number(m["Max"]); ok && max <= 0 {
				return fmt.Errorf("%s: Max must be positive", path)
			}
		}
		for _, f := range publicFields(t) {
			if err := validateResolved(f.Type, m[f.Name], path+"."+f.Name); err != nil {
				return err
			}
		}
	}
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		if items, ok := spec.([]any); ok {
			ids := map[string]bool{}
			if slices.Contains([]string{"MenuItem", "NavigationItem", "CommandAction", "TreeNode"}, t.Elem().Name()) {
				if err := validateTreeIDs(items, map[string]bool{}, 0, path); err != nil {
					return err
				}
			}
			values := map[string]bool{}
			for i, item := range items {
				if m, ok := item.(map[string]any); ok {
					name := t.Elem().Name()
					if strings.HasPrefix(name, "TableColumn[") || slices.Contains([]string{"AccordionItem", "AvatarItem", "ProgressSegment", "ListboxItem", "TreeNode", "TabItem", "StepItem", "TimelineItem", "NavigationItem", "CarouselItem", "FooterColumn", "FooterLink", "FieldGroupItem", "MenuItem", "CommandAction", "BreadcrumbItem"}, name) {
						id, _ := m["ID"].(string)
						if id == "" || ids[id] {
							return fmt.Errorf("%s: IDs must be unique and non-empty", path)
						}
						ids[id] = true
					}
					if name == "FieldGroupItem" && m["Build"] == nil {
						return fmt.Errorf("%s: each field-group item requires Build", path)
					}
					if strings.HasPrefix(name, "Option[") {
						key := fmt.Sprint(m["Value"])
						if values[key] {
							return fmt.Errorf("%s: option values must be unique", path)
						}
						values[key] = true
					}
				}
				if err := validateResolved(t.Elem(), item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateTreeIDs(items []any, seen map[string]bool, depth int, path string) error {
	if depth > 32 {
		return fmt.Errorf("%s: nesting exceeds 32 levels", path)
	}
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			id, _ := m["ID"].(string)
			if id == "" || seen[id] {
				return fmt.Errorf("%s: IDs must be unique and non-empty", path)
			}
			seen[id] = true
			if children, ok := m["Children"].([]any); ok {
				if err := validateTreeIDs(children, seen, depth+1, path); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func overlayExpression(b Binding, original string, patch any, source *valueSource) (string, error) {
	target := reflect.New(b.typ).Elem()
	lines, err := applyPatch(target, patch, b.ID, nil, source, nil, nui.Theme{})
	if err != nil {
		return "", err
	}

	if len(lines) == 0 {
		return original, nil
	}
	return "func() " + b.Type + " { o := " + original + "; " + strings.Join(lines, "; ") + "; return o }()", nil
}

func (d Demo) configuredBody(cfg Config, runtime bool, source *valueSource) (string, error) {
	body := d.RenderBody(cfg, runtime)
	bindings := d.bindingsFor(body)
	slices.SortFunc(bindings, func(a, b Binding) int { return b.start - a.start })
	for _, b := range bindings {
		if b.start < 0 {
			continue
		}
		original := body[b.start:b.end]
		replacement := original
		if runtime {
			replacement = "showcase.Options(cfg, " + strconv.Quote(b.ID) + ", " + original + ", c, t, &s.Live)"
		} else if patch, ok := cfg.Options[b.ID]; ok {
			var err error
			replacement, err = overlayExpression(b, original, patch, source)
			if err != nil {
				return "", err
			}
		}
		body = body[:b.start] + replacement + body[b.end:]
	}
	if d.ID == "theme" {
		if runtime {
			body = "t = showcase.Options(cfg, \"Theme\", t, c, t, &s.Live)\n" + body
		} else if patch, ok := cfg.Options["Theme"]; ok {
			for _, b := range bindings {
				if b.ID == "Theme" {
					expr, err := overlayExpression(b, "t", patch, source)
					if err != nil {
						return "", err
					}
					body = "t = " + expr + "\n" + body
				}
			}
		}
	}
	if runtime {
		body = "s.Live.Begin(c, t, cfg)\n" + body
	}
	return body, nil
}
