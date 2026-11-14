// Package showcase owns the website's examples, configurable values and code.
// Templates are compiled into preview views by internal/cmd/genwebsite.
package showcase

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	nui "github.com/josexy/nuxt-ui-go/ui"
)

//go:embed catalog.json
var catalogJSON []byte

//go:embed api-notes.json
var notesJSON []byte

func APINotes() map[string]Text {
	var notes map[string]Text
	if err := json.Unmarshal(notesJSON, &notes); err != nil {
		panic(err)
	}
	return notes
}

type Text struct {
	ZH string `json:"zh"`
	EN string `json:"en"`
}

type Choice struct {
	Value string `json:"value"`
	Go    string `json:"go"`
}

type Control struct {
	Key     string   `json:"key"`
	Label   Text     `json:"label"`
	Kind    string   `json:"kind"`
	Default any      `json:"default"`
	Choices []Choice `json:"choices,omitempty"`
	GoType  string   `json:"goType,omitempty"`
	Min     float64  `json:"min,omitempty"`
	Max     float64  `json:"max,omitempty"`
}

type Demo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description Text      `json:"description"`
	Guide       Text      `json:"guide"`
	Note        Text      `json:"note"`
	Symbols     []string  `json:"symbols"`
	Types       []string  `json:"types"`
	Controls    []Control `json:"controls"`
	State       string    `json:"state"`
	Setup       string    `json:"setup"`
	Body        string    `json:"body"`
	Height      int       `json:"height"`
}

func Catalog() []Demo {
	var demos []Demo
	if err := json.Unmarshal(catalogJSON, &demos); err != nil {
		panic(err)
	}
	return demos
}

func Find(id string) (Demo, bool) {
	for _, demo := range Catalog() {
		if demo.ID == id {
			return demo, true
		}
	}
	return Demo{}, false
}

type Config struct {
	Values      map[string]any `json:"values"`
	Options     map[string]any `json:"options,omitempty"`
	Primary     string         `json:"primary"`
	Neutral     string         `json:"neutral"`
	Mode        string         `json:"mode"`
	Locale      string         `json:"locale"`
	focusTarget bool
}

// PrimaryColors includes the mode-dependent Black option before the color ramps.
func PrimaryColors() []string {
	return append([]string{"black"}, nui.PrimaryPaletteNames()...)
}

// Normalize validates external messages before they reach component APIs.
// Nil optional values mean inheritance; false and zero remain explicit values.
func (d Demo) normalizeValues(input Config) (Config, error) {
	cfg := input
	cfg.Values = make(map[string]any, len(d.Controls))
	if cfg.Primary == "" {
		cfg.Primary = "green"
	}
	if cfg.Neutral == "" {
		cfg.Neutral = "slate"
	}
	if cfg.Mode == "" {
		cfg.Mode = "light"
	}
	if cfg.Locale == "" {
		cfg.Locale = "zh-Hans"
	}
	if !slices.Contains(PrimaryColors(), cfg.Primary) || !slices.Contains(nui.NeutralPaletteNames(), cfg.Neutral) {
		return Config{}, fmt.Errorf("unknown palette")
	}
	if !slices.Contains([]string{"light", "dark", "system"}, cfg.Mode) || !slices.Contains(nui.SupportedLocales(), cfg.Locale) {
		return Config{}, fmt.Errorf("unknown appearance or locale")
	}
	for _, control := range d.Controls {
		value, exists := input.Values[control.Key]
		if !exists {
			value = control.Default
		}
		valid := false
		switch control.Kind {
		case "text":
			s, ok := value.(string)
			valid = ok && len(s) <= 2000
		case "boolean":
			_, valid = value.(bool)
		case "optionalBoolean":
			_, valid = value.(bool)
			valid = valid || value == nil
		case "number", "integer":
			n, ok := value.(float64)
			valid = ok && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= control.Min && n <= control.Max
			if control.Kind == "integer" {
				valid = valid && math.Trunc(n) == n
			}
		case "select":
			s, ok := value.(string)
			valid = ok && slices.ContainsFunc(control.Choices, func(c Choice) bool { return c.Value == s })
		}
		if !valid {
			return Config{}, fmt.Errorf("invalid value for %s", control.Key)
		}
		cfg.Values[control.Key] = value
	}
	for key := range input.Values {
		if _, ok := cfg.Values[key]; !ok {
			return Config{}, fmt.Errorf("unknown parameter %s", key)
		}
	}
	return cfg, nil
}

func (c Config) String(key string) string  { value, _ := c.Values[key].(string); return value }
func (c Config) Bool(key string) bool      { value, _ := c.Values[key].(bool); return value }
func (c Config) Number(key string) float64 { value, _ := c.Values[key].(float64); return value }
func (c Config) Int(key string) int        { return int(c.Number(key)) }
func (c Config) OptionalBool(key string) nui.Optional[bool] {
	if c.Values[key] == nil {
		return nui.Optional[bool]{}
	}
	return nui.Set(c.Bool(key))
}

func (c Config) Theme() nui.AppConfig {
	palettes := nui.DefaultPalettes()
	if c.Primary != "black" {
		palettes.Primary, _ = nui.PaletteByName(c.Primary)
	}
	palettes.Neutral, _ = nui.PaletteByName(c.Neutral)
	mode := nui.LightMode
	if c.Mode == "dark" {
		mode = nui.DarkMode
	}
	if c.Mode == "system" {
		mode = nui.SystemMode
	}
	base := nui.AppConfig{Mode: mode, Light: nui.NewTheme(false, palettes), Dark: nui.NewTheme(true, palettes)}
	if c.Primary == "black" {
		base.Light = base.Light.WithBlackPrimary()
		base.Dark = base.Dark.WithBlackPrimary()
	}
	return Options(c, "AppConfig", base, nil, base.Light, nil)
}

// Select keeps the runtime and copied enum expressions in the same descriptor.
func Select[T any](cfg Config, key string, choices map[string]T) T { return choices[cfg.String(key)] }

func (c Control) expression(cfg Config, runtime bool) string {
	key := strconv.Quote(c.Key)
	if runtime {
		switch c.Kind {
		case "text":
			return "cfg.String(" + key + ")"
		case "boolean":
			return "cfg.Bool(" + key + ")"
		case "optionalBoolean":
			return "cfg.OptionalBool(" + key + ")"
		case "number":
			return "cfg.Number(" + key + ")"
		case "integer":
			return "cfg.Int(" + key + ")"
		case "select":
			var entries []string
			for _, choice := range c.Choices {
				entries = append(entries, strconv.Quote(choice.Value)+":"+choice.Go)
			}
			return "showcase.Select(cfg, " + key + ", map[string]" + c.GoType + "{" + strings.Join(entries, ",") + "})"
		}
	}
	value := cfg.Values[c.Key]
	switch c.Kind {
	case "text":
		return strconv.Quote(value.(string))
	case "boolean":
		return strconv.FormatBool(value.(bool))
	case "optionalBoolean":
		if value == nil {
			return "nui.Optional[bool]{}"
		}
		return "nui.Set(" + strconv.FormatBool(value.(bool)) + ")"
	case "number", "integer":
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	case "select":
		for _, choice := range c.Choices {
			if choice.Value == value {
				return choice.Go
			}
		}
	}
	panic("unvalidated showcase configuration")
}

func (d Demo) RenderBody(cfg Config, runtime bool) string {
	pairs := make([]string, 0, len(d.Controls)*2)
	for _, control := range d.Controls {
		pairs = append(pairs, "{{"+control.Key+"}}", control.expression(cfg, runtime))
	}
	// Replace only tokens from the template, never token-like text entered by
	// the user inside an already quoted Go string.
	return strings.NewReplacer(pairs...).Replace(d.Body)
}
