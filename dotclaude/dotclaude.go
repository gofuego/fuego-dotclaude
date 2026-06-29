package dotclaude

import (
	"embed"
	"io/fs"

	"github.com/gofuego/fuego/core"
)

//go:embed theme
var themeFS embed.FS

//go:embed config-defaults.yaml
var configDefaults []byte

// Pack returns the fuego-dotclaude format pack: a generic Markdown parser, the
// embedded theme, the artifact route defaults, and the classification hook.
func Pack() core.Pack {
	theme, _ := fs.Sub(themeFS, "theme")
	return core.Pack{
		Name: "dotclaude",
		Parsers: []core.Parser{
			NewParser(),
			&MCPParser{},
			&SettingsParser{},
			&SettingsParser{local: true},
			PluginParser{},
			MarketplaceParser{},
		},
		Theme:          theme,
		ConfigDefaults: configDefaults,
		Hooks: core.Hooks{
			AfterParse:   []core.AfterParseHook{AfterParseHook()},
			Index:        []core.IndexHook{IndexHook(), PluginHook(), HomeHook(), ReferenceHook()},
			BeforeRender: []core.BeforeRenderHook{LinkHook()},
		},
	}
}
