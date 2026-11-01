package nui

import (
	"embed"
	"io/fs"
	"strings"
	"sync"

	"github.com/egoist/mygo/i18n"
	_ "github.com/egoist/mygo/locales/zh-Hans"
	"github.com/egoist/mygo/ui"
)

//go:embed locales/en.json locales/zh-Hans.json
var localeFiles embed.FS

// A library catalog follows the application's locale without changing its
// supported languages. Applications can override entries with nuxtui.<key>.
var componentMessages = sync.OnceValue(func() *i18n.Catalog {
	files, err := fs.Sub(localeFiles, "locales")
	if err != nil {
		panic(err)
	}
	catalog, err := i18n.NewLibraryCatalog(files, "en", "nuxtui")
	if err != nil {
		panic(err)
	}
	return catalog
})

// SupportedLocales returns the bundled component languages, English first.
// An app without its own catalogs can declare these with
// mygo.App.SetLanguages(SupportedLocales()...) and switch all its windows with
// mygo.App.SetLocale("zh-Hans"), SetLocale("en"), or SetLocale("") for the system.
// Applications with catalogs declare their own languages instead.
func SupportedLocales() []string { return componentMessages().Languages() }

func message(c *ui.Context, key string) string {
	return componentMessages().Format(c.Locale(), key, nil)
}

func formatMessage(c *ui.Context, key string, args map[string]any) string {
	return componentMessages().Format(c.Locale(), key, args)
}

func namedMessage(c *ui.Context, key, name string) string {
	return strings.TrimSpace(formatMessage(c, key, map[string]any{"name": name}))
}

// Keep the identity of built-in errors so a language change can reformat an
// already displayed error. External errors remain application-owned text.
type localizedError struct {
	key  string
	args map[string]any
}

func (e *localizedError) Error() string { return e.in(i18n.Locale()) }
func (e *localizedError) in(locale string) string {
	return componentMessages().Format(locale, e.key, e.args)
}
