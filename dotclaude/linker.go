package dotclaude

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gofuego/fuego-dotclaude/dotclaude/classify"
	"github.com/gofuego/fuego-dotclaude/dotclaude/sluglink"
	"github.com/gofuego/fuego/core"
)

// LinkHook builds the slug registry from every named artifact and rewrites each
// page's rendered body so whole-word mentions of an artifact's name become links
// to it. It runs in BEFORE-RENDER, after ROUTE has resolved page URLs, so the
// targets it links to are final. Ambiguous names are warned once and linked to a
// disambiguation route (the page itself is built in a later slice).
func LinkHook() core.BeforeRenderHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		reg := buildRegistry(pages)

		for _, key := range reg.Collisions() {
			targets, _ := reg.Resolve(key)
			fmt.Printf("fuego-dotclaude: warning: ambiguous name %q matches %d artifacts; mentions link to %s\n",
				key, len(targets), sluglink.DisambiguationURL(key))
		}

		for _, p := range pages {
			selfName := pageName(p)
			var links []sluglink.Link
			for i := range p.Nodes {
				n := &p.Nodes[i]
				if !n.Raw || n.Type != "body" {
					continue
				}
				out, ls, err := reg.Link(n.Content, selfName)
				if err != nil {
					continue // degrade: leave the body unmodified
				}
				n.Content = out
				links = append(links, ls...)
			}
			if rels := linkTargets(links); len(rels) > 0 {
				p.Envelope["links_to"] = rels // consumed by the backlink builder
			}
		}
		return pages, nil
	}
}

// buildRegistry collects link targets from named artifacts: agents, skills, and
// output styles by name; commands by namespaced name; MCP servers by name
// (anchored on the MCP page).
func buildRegistry(pages []*core.Page) *sluglink.Registry {
	reg := sluglink.NewRegistry(stopwords)
	for _, p := range pages {
		switch classify.Kind(p.Type) {
		case classify.KindAgent, classify.KindSkill, classify.KindOutputStyle:
			if name := pageName(p); name != "" {
				reg.Add(sluglink.Target{Name: name, URL: baseRel(p.URL), Kind: p.Type})
			}
		case classify.KindCommand:
			if name := pageName(p); name != "" {
				reg.Add(sluglink.Target{Name: name, URL: baseRel(p.URL), Kind: "command"})
			}
		default:
			if p.Type == "mcp" {
				for _, sv := range serverNames(p) {
					reg.Add(sluglink.Target{
						Name: sv,
						URL:  baseRel(p.URL) + "#server-" + sv,
						Kind: "mcp-server",
					})
				}
			}
		}
	}
	return reg
}

// pageName returns the matchable name of an artifact page: frontmatter name for
// agents/skills/output styles, the namespaced name (sans leading slash) for
// commands. Empty for pages that aren't link targets.
func pageName(p *core.Page) string {
	switch classify.Kind(p.Type) {
	case classify.KindCommand:
		return strings.TrimPrefix(stringOf(p.Envelope["command_name"]), "/")
	case classify.KindAgent, classify.KindSkill, classify.KindOutputStyle:
		return stringOf(p.Envelope["name"])
	}
	return ""
}

func serverNames(p *core.Page) []string {
	switch v := p.Envelope["mcp_server_names"].(type) {
	case []string:
		return v
	default:
		return nil
	}
}

// linkTargets projects a page's emitted links into a sorted, unique slice of
// {slug, url} maps for the backlink builder.
func linkTargets(links []sluglink.Link) []map[string]any {
	if len(links) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []map[string]any
	for _, l := range links {
		if seen[l.URL] {
			continue
		}
		seen[l.URL] = true
		out = append(out, map[string]any{"slug": l.Slug, "url": l.URL, "ambiguous": l.Ambiguous})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["url"].(string) < out[j]["url"].(string) })
	return out
}

func stringOf(v any) string {
	s, _ := v.(string)
	return s
}

// stopwords guards the linker against common words being treated as slugs.
// Kept to high-frequency English function words that are unlikely artifact
// names, so legitimate names like "test" or "build" still link.
var stopwords = []string{
	"a", "an", "and", "are", "as", "at", "be", "but", "by", "for", "from",
	"if", "in", "into", "is", "it", "its", "of", "on", "or", "out", "over",
	"so", "than", "that", "the", "then", "this", "to", "up", "use", "via",
	"was", "when", "with", "you", "your",
}
