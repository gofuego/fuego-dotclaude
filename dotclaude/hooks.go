package dotclaude

import (
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofuego/fuego-dotclaude/dotclaude/classify"
	"github.com/gofuego/fuego/core"
)

// AfterParseHook recovers each page's Claude Code artifact kind from its
// RelPath and stamps the type, layout, routing slug, and kind-specific display
// metadata onto the page. It runs after PARSE and before ROUTE, so the type and
// slug it assigns drive both URL routing and layout selection. Pages it doesn't
// recognize keep their default type and render generically.
func AfterParseHook() core.AfterParseHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		for _, p := range pages {
			// Injected siblings classify and enrich themselves; don't reprocess.
			if sib, _ := p.Envelope["sibling"].(bool); sib {
				continue
			}
			kind := classify.Classify(p.RelPath)
			if kind == classify.KindUnknown {
				continue
			}

			p.Type = string(kind)
			if p.Layout == "" {
				p.Layout = string(kind)
			}
			if slug := classify.Slug(p.RelPath); slug != "" {
				p.Envelope["slug"] = slug
			}
			enrich(kind, p)
		}
		return pages, nil
	}
}

// enrich adds kind-specific display fields the theme reads, plus the normalized
// taxonomy fields (tools, source) and a title fallback used by taxonomy and
// catalog listings.
func enrich(kind classify.Kind, p *core.Page) {
	env := p.Envelope

	// Provenance: every artifact is project-scoped for now; plugin artifacts
	// override this in the plugins slice.
	if _, ok := env["source"]; !ok {
		env["source"] = "project"
	}

	switch kind {
	case classify.KindAgent:
		normalizeTools(env)
	case classify.KindSkill:
		if name, _ := env["name"].(string); name == "" {
			env["name"] = path.Base(classify.SkillRoot(p.RelPath))
		}
	case classify.KindCommand:
		env["command_name"] = classify.CommandName(p.RelPath)
		normalizeTools(env)
	case classify.KindMemory:
		env["memory_path"] = filepath.ToSlash(p.RelPath)
	}

	setTitle(kind, p)
}

// normalizeTools coerces an artifact's tool list into the "tools" taxonomy
// field. Agents declare "tools", commands declare "allowed-tools"; both may be a
// comma-separated string or a YAML list. A missing field leaves no tools.
func normalizeTools(env core.Envelope) {
	if v, ok := env["tools"]; ok {
		env["tools"] = toStringSlice(v)
		return
	}
	if v, ok := env["allowed-tools"]; ok {
		env["tools"] = toStringSlice(v)
	}
}

// setTitle gives every artifact a "title" (used by taxonomy/catalog listings and
// the manifest) when one isn't already present.
func setTitle(kind classify.Kind, p *core.Page) {
	if t, _ := p.Envelope["title"].(string); t != "" {
		return
	}
	switch kind {
	case classify.KindCommand:
		p.Envelope["title"] = stringOf(p.Envelope["command_name"])
	case classify.KindMemory:
		p.Envelope["title"] = stringOf(p.Envelope["memory_path"])
	default:
		if n := stringOf(p.Envelope["name"]); n != "" {
			p.Envelope["title"] = n
		}
	}
}

// IndexHook attaches each skill page's bundled files: its sibling Markdown docs
// (as resolved sub-page links) and its non-Markdown assets (copied verbatim to
// output, linked at their natural path). It runs during INDEX, after ROUTE has
// resolved real-page URLs, so the doc links point at the docs' final routes.
func IndexHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		skills := map[string]*core.Page{}
		docs := map[string][]*core.Page{}

		for _, p := range pages {
			switch classify.Kind(p.Type) {
			case classify.KindSkill:
				skills[classify.SkillRoot(p.RelPath)] = p
			case classify.KindSkillDoc:
				root := classify.SkillRoot(p.RelPath)
				docs[root] = append(docs[root], p)
			}
		}

		for root, skill := range skills {
			attachBundledDocs(skill, docs[root], root)
			attachBundledAssets(skill, root)
		}
		return pages, nil
	}
}

// attachBundledDocs records the skill's sibling Markdown docs as links and
// back-links each doc to its skill.
func attachBundledDocs(skill *core.Page, docPages []*core.Page, root string) {
	sort.Slice(docPages, func(i, j int) bool { return docPages[i].URL < docPages[j].URL })

	var docs []map[string]any
	for _, d := range docPages {
		docs = append(docs, map[string]any{
			"name": docName(d),
			"url":  baseRel(d.URL),
		})
		d.Envelope["skill_name"] = skill.Envelope["name"]
		d.Envelope["skill_url"] = baseRel(skill.URL)
	}
	if len(docs) > 0 {
		skill.Envelope["bundled_docs"] = docs
	}
}

// attachBundledAssets lists the skill directory's non-Markdown files. They are
// copied to output by the engine's STATIC phase at their content-relative path,
// so the link is just that path (base-relative under the site's <base href>).
func attachBundledAssets(skill *core.Page, root string) {
	dir := filepath.Dir(skill.SourcePath)

	var assets []map[string]any
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		base := d.Name()
		if base == "SKILL.md" || strings.HasSuffix(strings.ToLower(base), ".md") {
			return nil // SKILL.md and docs are pages, not assets
		}
		rel, relErr := filepath.Rel(dir, p)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		assets = append(assets, map[string]any{
			"name": rel,
			"url":  root + "/" + rel,
		})
		return nil
	})

	sort.Slice(assets, func(i, j int) bool {
		return assets[i]["name"].(string) < assets[j]["name"].(string)
	})
	if len(assets) > 0 {
		skill.Envelope["bundled_assets"] = assets
	}
}

// baseRel converts an engine page URL ("/skills/x/") to a base-relative link
// ("skills/x/") so it resolves correctly under the site's <base href> for any
// deployment base URL.
func baseRel(url string) string { return strings.TrimPrefix(url, "/") }

// docName is a skill doc's display label: its frontmatter title/name, else its
// filename stem.
func docName(p *core.Page) string {
	if t, ok := p.Envelope["title"].(string); ok && t != "" {
		return t
	}
	if n, ok := p.Envelope["name"].(string); ok && n != "" {
		return n
	}
	base := filepath.Base(p.RelPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// toStringSlice normalizes a tools value to []string. Agent frontmatter writes
// tools either as a comma-separated string ("Read, Write, Bash") or a YAML
// list; an omitted field means "all tools".
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case string:
		var out []string
		for _, s := range strings.Split(val, ",") {
			if s = strings.TrimSpace(s); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return val
	default:
		return nil
	}
}
