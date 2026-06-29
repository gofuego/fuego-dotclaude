package dotclaude

import (
	"sort"
	"strings"

	"github.com/gofuego/fuego-dotclaude/dotclaude/refgraph"
	"github.com/gofuego/fuego-dotclaude/dotclaude/sluglink"
	"github.com/gofuego/fuego/core"
)

// ReferenceHook turns the forward links discovered by the slug linker into
// reverse navigation: a "Referenced by" list on each artifact, one
// disambiguation page per ambiguous name (the target of ambiguous auto-links),
// and a global reference-graph index. It runs in INDEX (after ROUTE, so URLs are
// final) and appends its pages there so they are collision-checked. It scans
// page bodies to derive the graph but does not rewrite them — the actual
// link rewriting happens later, in BEFORE-RENDER.
func ReferenceHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		reg := buildRegistry(pages)

		nodes := make([]refgraph.Node, 0, len(pages))
		byURL := map[string]*core.Page{}
		for _, p := range pages {
			if p.Skip {
				continue
			}
			url := baseRel(p.URL)
			byURL[url] = p
			nodes = append(nodes, refgraph.Node{
				URL:   url,
				Title: stringOf(p.Envelope["title"]),
				Kind:  p.Type,
				Links: forwardLinks(reg, p),
			})
		}

		g := refgraph.Build(nodes)
		for url, refs := range g.Backlinks {
			if p, ok := byURL[url]; ok {
				p.Envelope["referenced_by"] = refsToMaps(refs)
			}
		}

		var added []*core.Page
		added = append(added, disambiguationPages(reg)...)
		added = append(added, graphIndexPage(g))
		return append(pages, added...), nil
	}
}

// forwardLinks scans a page's body nodes and returns the page URLs it links to,
// expanding ambiguous mentions to every candidate and stripping anchors so MCP
// server mentions attribute to the MCP page.
func forwardLinks(reg *sluglink.Registry, p *core.Page) []string {
	self := pageName(p)
	var links []string
	for _, n := range p.Nodes {
		if !n.Raw || n.Type != "body" {
			continue
		}
		_, ls, err := reg.Link(n.Content, self)
		if err != nil {
			continue
		}
		for _, l := range ls {
			if l.Ambiguous {
				targets, _ := reg.Resolve(l.Slug)
				for _, t := range targets {
					links = append(links, pageURL(t.URL))
				}
				continue
			}
			links = append(links, pageURL(l.URL))
		}
	}
	return links
}

// pageURL strips a #fragment so a link resolves to its page.
func pageURL(u string) string {
	if i := strings.IndexByte(u, '#'); i >= 0 {
		return u[:i]
	}
	return u
}

// disambiguationPages builds one virtual page per ambiguous name, listing every
// artifact that shares it. Their URLs match the links the linker emits.
func disambiguationPages(reg *sluglink.Registry) []*core.Page {
	sets := reg.DisambiguationSets()
	keys := make([]string, 0, len(sets))
	for k := range sets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pages []*core.Page
	for _, key := range keys {
		targets := sets[key]
		candidates := make([]map[string]any, 0, len(targets))
		for _, t := range targets {
			candidates = append(candidates, map[string]any{
				"title": t.Name,
				"url":   t.URL,
				"kind":  t.Kind,
			})
		}
		pages = append(pages, &core.Page{
			RelPath: "_virtual/disambiguation/" + sluglink.DisambiguationSlug(key),
			Type:    "disambiguation",
			URL:     "/" + sluglink.DisambiguationURL(key),
			Layout:  "disambiguation",
			Envelope: core.Envelope{
				"title":      key,
				"term":       key,
				"candidates": candidates,
			},
		})
	}
	return pages
}

// graphIndexPage builds the global reference-graph index page.
func graphIndexPage(g *refgraph.Graph) *core.Page {
	var edges []map[string]any
	for _, e := range g.Edges {
		targets := make([]map[string]any, 0, len(e.Targets))
		for _, t := range e.Targets {
			targets = append(targets, map[string]any{"url": t.URL, "title": refTitle(t)})
		}
		edges = append(edges, map[string]any{
			"source_url":   e.Source.URL,
			"source_title": refTitle(e.Source),
			"targets":      targets,
		})
	}
	return &core.Page{
		RelPath: "_virtual/references",
		Type:    "reference-graph",
		URL:     "/references/",
		Layout:  "reference-graph",
		Envelope: core.Envelope{
			"title": "Reference graph",
			"edges": edges,
		},
	}
}

func refsToMaps(refs []refgraph.Ref) []map[string]any {
	out := make([]map[string]any, 0, len(refs))
	for _, r := range refs {
		out = append(out, map[string]any{"url": r.URL, "title": refTitle(r), "kind": r.Kind})
	}
	return out
}

// refTitle falls back to the URL when a target has no resolved title.
func refTitle(r refgraph.Ref) string {
	if r.Title != "" {
		return r.Title
	}
	return r.URL
}
