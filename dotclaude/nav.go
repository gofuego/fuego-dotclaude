package dotclaude

import (
	"html"
	"sort"
	"strings"

	"github.com/gofuego/fuego/core"
)

// NavHook builds the filesystem-style navigation. It runs last in INDEX, after
// every other page exists and URLs are final. It does two things:
//
//  1. Generates a directory-index page for every intermediate route segment that
//     no real page claims (so /agents/, /agents/<plugin>/, /commands/git/, …
//     stop being empty and instead list what they contain).
//  2. Renders an IDE-style collapsible file tree of the whole artifact set and
//     stores it on every page so the sidebar is identical site-wide (the active
//     node is resolved client-side).
func NavHook() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		shells := makeSectionShells(pages)
		all := append(pages, shells...)

		for _, s := range shells {
			s.Envelope["children"] = childrenOf(all, baseRel(s.URL))
		}

		sidebar := renderSidebar(all)
		for _, p := range all {
			p.Envelope["sidebar"] = sidebar
		}
		return all, nil
	}
}

// makeSectionShells creates a "section" page for each intermediate directory URL
// that no page already claims.
func makeSectionShells(pages []*core.Page) []*core.Page {
	have := map[string]bool{}
	for _, p := range pages {
		have[baseRel(p.URL)] = true
	}

	needed := map[string]bool{}
	for _, p := range pages {
		if !inTree(p) {
			continue
		}
		segs := splitURL(baseRel(p.URL))
		for i := 1; i < len(segs); i++ {
			needed[strings.Join(segs[:i], "/")+"/"] = true
		}
	}

	dirs := make([]string, 0, len(needed))
	for dir := range needed {
		if !have[dir] {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)

	shells := make([]*core.Page, 0, len(dirs))
	for _, dir := range dirs {
		shells = append(shells, &core.Page{
			RelPath:  "_virtual/section/" + dir,
			Type:     "section",
			URL:      "/" + dir,
			Layout:   "section",
			Envelope: core.Envelope{"title": lastSegment(dir)},
		})
	}
	return shells
}

// childrenOf returns the immediate children of a directory URL, dirs first.
func childrenOf(all []*core.Page, dir string) []map[string]any {
	var kids []map[string]any
	for _, p := range all {
		if !inTree(p) {
			continue
		}
		u := baseRel(p.URL)
		if u == dir || parentURL(u) != dir {
			continue
		}
		kids = append(kids, map[string]any{
			"name":   lastSegment(u),
			"url":    u,
			"kind":   p.Type,
			"is_dir": hasChildren(all, u),
		})
	}
	sort.Slice(kids, func(i, j int) bool {
		di, dj := kids[i]["is_dir"].(bool), kids[j]["is_dir"].(bool)
		if di != dj {
			return di
		}
		return kids[i]["name"].(string) < kids[j]["name"].(string)
	})
	return kids
}

func hasChildren(all []*core.Page, u string) bool {
	for _, p := range all {
		if inTree(p) && parentURL(baseRel(p.URL)) == u {
			return true
		}
	}
	return false
}

// treeNode is a node in the rendered file tree.
type treeNode struct {
	url, label string
	children   []*treeNode
}

// renderSidebar builds the nested file tree from the artifact set and renders it
// as collapsible <details> elements.
func renderSidebar(pages []*core.Page) string {
	nodes := map[string]*treeNode{}
	for _, p := range pages {
		if !inTree(p) {
			continue
		}
		u := baseRel(p.URL)
		nodes[u] = &treeNode{url: u, label: lastSegment(u)}
	}

	var roots []*treeNode
	for u, n := range nodes {
		if parent, ok := nodes[parentURL(u)]; ok && parentURL(u) != "" {
			parent.children = append(parent.children, n)
		} else {
			roots = append(roots, n)
		}
	}

	var sortTree func(ns []*treeNode)
	sortTree = func(ns []*treeNode) {
		sort.Slice(ns, func(i, j int) bool {
			di, dj := len(ns[i].children) > 0, len(ns[j].children) > 0
			if di != dj {
				return di
			}
			return ns[i].label < ns[j].label
		})
		for _, n := range ns {
			sortTree(n.children)
		}
	}
	sortTree(roots)

	var sb strings.Builder
	sb.WriteString(`<nav class="filetree"><ul>`)
	sb.WriteString(`<li class="file"><a href="." data-home="1">Home</a></li>`)
	var render func(ns []*treeNode, depth int)
	render = func(ns []*treeNode, depth int) {
		for _, n := range ns {
			label := html.EscapeString(n.label)
			href := html.EscapeString(n.url)
			if len(n.children) > 0 {
				open := ""
				if depth == 0 {
					open = " open"
				}
				sb.WriteString(`<li class="dir"><details` + open + `><summary><a href="` + href + `">` + label + `</a></summary><ul>`)
				render(n.children, depth+1)
				sb.WriteString(`</ul></details></li>`)
			} else {
				sb.WriteString(`<li class="file"><a href="` + href + `">` + label + `</a></li>`)
			}
		}
	}
	render(roots, 0)
	sb.WriteString(`</ul></nav>`)
	return sb.String()
}

// inTree reports whether a page belongs in the filesystem sidebar (real
// artifacts and section folders; not the home page or derived views).
func inTree(p *core.Page) bool {
	if p.Skip || baseRel(p.URL) == "" {
		return false
	}
	switch p.Type {
	case "taxonomy-term", "taxonomy-index", "disambiguation", "reference-graph", "virtual":
		return false
	}
	return true
}

func splitURL(u string) []string {
	u = strings.Trim(u, "/")
	if u == "" {
		return nil
	}
	return strings.Split(u, "/")
}

func parentURL(u string) string {
	t := strings.TrimSuffix(u, "/")
	i := strings.LastIndex(t, "/")
	if i < 0 {
		return ""
	}
	return t[:i+1]
}

func lastSegment(u string) string {
	t := strings.TrimSuffix(u, "/")
	if i := strings.LastIndex(t, "/"); i >= 0 {
		return t[i+1:]
	}
	return t
}
