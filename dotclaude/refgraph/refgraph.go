// Package refgraph turns a set of pages and their forward links (who mentions
// whom, already resolved by the slug linker) into reverse navigation: per-target
// backlinks and the global edge list. It is a pure deep module over an in-memory
// representation — no engine types, no I/O — so the graph logic is testable in
// isolation.
package refgraph

import "sort"

// Node is a page and the target URLs its body links to.
type Node struct {
	URL   string
	Title string
	Kind  string
	Links []string
}

// Ref is a lightweight reference to a page.
type Ref struct {
	URL   string
	Title string
	Kind  string
}

// Edge is one page and the pages it references.
type Edge struct {
	Source  Ref
	Targets []Ref
}

// Graph is the computed reference structure.
type Graph struct {
	Backlinks map[string][]Ref // target URL -> sources that reference it
	Edges     []Edge           // source -> targets, for the global index
}

// Build computes backlinks and edges from the nodes' forward links. Self-links
// are dropped, targets and backlinks are deduped, and everything is sorted by
// URL for deterministic output. Targets resolve their title/kind from the node
// set when known.
func Build(nodes []Node) *Graph {
	byURL := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		byURL[n.URL] = n
	}

	sorted := append([]Node(nil), nodes...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].URL < sorted[j].URL })

	g := &Graph{Backlinks: map[string][]Ref{}}
	for _, n := range sorted {
		seen := map[string]bool{}
		var targets []Ref
		for _, l := range n.Links {
			if l == n.URL || seen[l] {
				continue
			}
			seen[l] = true

			ref := Ref{URL: l}
			if t, ok := byURL[l]; ok {
				ref.Title, ref.Kind = t.Title, t.Kind
			}
			targets = append(targets, ref)
			g.Backlinks[l] = append(g.Backlinks[l], Ref{URL: n.URL, Title: n.Title, Kind: n.Kind})
		}
		if len(targets) > 0 {
			sort.Slice(targets, func(i, j int) bool { return targets[i].URL < targets[j].URL })
			g.Edges = append(g.Edges, Edge{
				Source:  Ref{URL: n.URL, Title: n.Title, Kind: n.Kind},
				Targets: targets,
			})
		}
	}

	for url, refs := range g.Backlinks {
		sort.Slice(refs, func(i, j int) bool { return refs[i].URL < refs[j].URL })
		g.Backlinks[url] = refs
	}
	return g
}
