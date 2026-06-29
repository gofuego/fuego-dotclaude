package dotclaude

import (
	"path/filepath"
	"sort"

	"github.com/gofuego/fuego-dotclaude/dotclaude/classify"
	"github.com/gofuego/fuego/core"
)

// catalogOrder fixes the artifact sections shown in the catalog and their
// labels. Kinds not listed here (e.g. skill docs) are sub-pages and stay out of
// the top-level catalog.
var catalogOrder = []struct{ kind, label string }{
	{"agent", "Agents"},
	{"command", "Commands"},
	{"skill", "Skills"},
	{"output-style", "Output Styles"},
	{"memory", "Memory"},
	{"mcp", "MCP"},
	{"settings", "Settings"},
	{"settings-local", "Local Settings"},
	{"plugin", "Plugins"},
	{"marketplace", "Marketplaces"},
}

// HomeHook makes the site's home page the top-level CLAUDE.md (followed by a
// generated catalog), or a pure generated dashboard when no CLAUDE.md exists. It
// runs in INDEX so the home/dashboard page is collision-checked alongside every
// other page.
func HomeHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		home := pickHome(pages)
		catalog := buildCatalog(pages, home)

		if home != nil {
			// Re-home the chosen CLAUDE.md to "/" so it renders once, with the
			// catalog appended. The non-chosen memory pages keep their routes.
			home.URL = "/"
			home.Layout = "home"
			home.Envelope["is_home"] = true
			home.Envelope["catalog"] = catalog
			return pages, nil
		}

		// No CLAUDE.md anywhere: generate a dashboard at "/".
		dashboard := &core.Page{
			RelPath: "_virtual/home",
			Type:    "virtual",
			URL:     "/",
			Layout:  "home",
			Envelope: core.Envelope{
				"title":        "Home",
				"is_dashboard": true,
				"catalog":      catalog,
			},
		}
		return append(pages, dashboard), nil
	}
}

// pickHome selects the CLAUDE.md to use as the home page. A root-level sibling
// CLAUDE.md (injected one level above the content dir) wins over the content
// dir's own CLAUDE.md; deeper nested CLAUDE.md files are never the home. Returns
// nil when there is no candidate.
func pickHome(pages []*core.Page) *core.Page {
	var best *core.Page
	bestPriority := 0
	for _, p := range pages {
		if classify.Kind(p.Type) != classify.KindMemory {
			continue
		}
		if filepath.Base(p.RelPath) != "CLAUDE.md" {
			continue // CLAUDE.local.md and the like are not home candidates
		}
		priority := homePriority(p)
		if priority > bestPriority {
			best, bestPriority = p, priority
		}
	}
	return best
}

// homePriority ranks a CLAUDE.md as a home candidate: an injected root sibling
// (2) beats the content dir's top-level CLAUDE.md (1); anything deeper is not a
// candidate (0).
func homePriority(p *core.Page) int {
	if sib, _ := p.Envelope["sibling"].(bool); sib {
		return 2
	}
	if p.RelPath == "CLAUDE.md" {
		return 1
	}
	return 0
}

// buildCatalog groups real artifact pages by kind, in catalogOrder, with counts
// and base-relative links. The home page and virtual pages are excluded.
func buildCatalog(pages []*core.Page, home *core.Page) []map[string]any {
	byKind := map[string][]map[string]any{}
	for _, p := range pages {
		if p == home || p.Skip || p.Type == "virtual" {
			continue
		}
		name := displayName(p)
		if name == "" {
			continue
		}
		byKind[p.Type] = append(byKind[p.Type], map[string]any{
			"name": name,
			"url":  baseRel(p.URL),
		})
	}

	var sections []map[string]any
	for _, sec := range catalogOrder {
		items := byKind[sec.kind]
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i]["name"].(string) < items[j]["name"].(string)
		})
		sections = append(sections, map[string]any{
			"label": sec.label,
			"count": len(items),
			"items": items,
		})
	}
	return sections
}

// displayName is an artifact's catalog label.
func displayName(p *core.Page) string {
	switch classify.Kind(p.Type) {
	case classify.KindAgent, classify.KindSkill, classify.KindOutputStyle:
		if n := stringOf(p.Envelope["name"]); n != "" {
			return n
		}
		return stem(p.RelPath)
	case classify.KindCommand:
		return stringOf(p.Envelope["command_name"])
	case classify.KindMemory:
		return stringOf(p.Envelope["memory_path"])
	}
	// JSON config pages and anything else: use the title.
	return stringOf(p.Envelope["title"])
}

func stem(relPath string) string {
	base := filepath.Base(relPath)
	return base[:len(base)-len(filepath.Ext(base))]
}
