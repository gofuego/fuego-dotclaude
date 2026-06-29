package dotclaude

import (
	"sort"

	"github.com/gofuego/fuego-dotclaude/dotclaude/jsonmodel"
	"github.com/gofuego/fuego/core"
)

// PluginParser renders a plugin's plugin.json manifest as the plugin's page.
// The AfterParse hook namespaces its route under the owning plugin.
type PluginParser struct{}

func (PluginParser) Type() string        { return "plugin" }
func (PluginParser) Filenames() []string { return []string{"plugin.json"} }

func (PluginParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env := core.Envelope{"layout": "plugin", "raw_json": jsonmodel.Pretty(raw)}
	pm, err := jsonmodel.ParsePluginManifest(raw)
	if err != nil {
		env["title"] = "Plugin"
		env["parse_error"] = err.Error()
		return env, nil, nil
	}
	env["title"] = pm.Name
	env["manifest_name"] = pm.Name
	env["version"] = pm.Version
	env["description"] = pm.Description
	env["author"] = pm.Author
	env["homepage"] = pm.Homepage
	env["license"] = pm.License
	env["keywords"] = pm.Keywords
	return env, nil, nil
}

// MarketplaceParser renders a marketplace.json as a marketplace page listing its
// plugins.
type MarketplaceParser struct{}

func (MarketplaceParser) Type() string        { return "marketplace" }
func (MarketplaceParser) Filenames() []string { return []string{"marketplace.json"} }

func (MarketplaceParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env := core.Envelope{"layout": "marketplace", "raw_json": jsonmodel.Pretty(raw)}
	mk, err := jsonmodel.ParseMarketplace(raw)
	if err != nil {
		env["title"] = "Marketplace"
		env["parse_error"] = err.Error()
		return env, nil, nil
	}
	env["title"] = mk.Name
	env["owner"] = mk.Owner
	plugins := make([]map[string]any, 0, len(mk.Plugins))
	for _, p := range mk.Plugins {
		plugins = append(plugins, map[string]any{
			"name":        p.Name,
			"source":      p.Source,
			"description": p.Description,
		})
	}
	env["plugins"] = plugins
	return env, nil, nil
}

// PluginHook attaches to each plugin page the list of components that plugin
// provides (its agents, skills, commands, MCP config, …), with resolved links.
// It runs in INDEX, after ROUTE, so the component URLs are final.
func PluginHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		components := map[string][]map[string]any{}
		var pluginPages []*core.Page

		for _, p := range pages {
			if p.Type == "plugin" {
				pluginPages = append(pluginPages, p)
				continue
			}
			plugin := stringOf(p.Envelope["plugin"])
			if plugin == "" {
				continue
			}
			components[plugin] = append(components[plugin], map[string]any{
				"name": displayName(p),
				"kind": p.Type,
				"url":  baseRel(p.URL),
			})
		}

		for _, pp := range pluginPages {
			plugin := stringOf(pp.Envelope["plugin"])
			comps := components[plugin]
			sort.Slice(comps, func(i, j int) bool {
				if comps[i]["kind"] != comps[j]["kind"] {
					return comps[i]["kind"].(string) < comps[j]["kind"].(string)
				}
				return comps[i]["name"].(string) < comps[j]["name"].(string)
			})
			if len(comps) > 0 {
				pp.Envelope["components"] = comps
			}
		}
		return pages, nil
	}
}

// namespacePluginConfig namespaces a plugin's JSON config/manifest pages (which
// the classifier doesn't reach, since they aren't Markdown) by plugin: routing
// slug, provenance, and the plugin display field. MCP config is retyped so it
// routes under the plugin instead of colliding with the root MCP page.
func namespacePluginConfig(p *core.Page, plugin string) {
	p.Envelope["plugin"] = plugin
	p.Envelope["source"] = plugin
	switch p.Type {
	case "plugin", "marketplace":
		p.Envelope["slug"] = plugin
	case "mcp":
		p.Type = "plugin-mcp"
		p.Layout = "mcp"
		p.Envelope["slug"] = plugin
	}
}

// isPluginConfigType reports whether a page's parser type is a JSON
// config/manifest the path classifier won't handle (so plugin namespacing must
// be applied to it explicitly).
func isPluginConfigType(t string) bool {
	switch t {
	case "plugin", "marketplace", "mcp", "settings", "settings-local":
		return true
	}
	return false
}
