package refgraph

import (
	"reflect"
	"testing"
)

func TestBuildBacklinks(t *testing.T) {
	nodes := []Node{
		{URL: "a/", Title: "A", Kind: "agent", Links: []string{"b/"}},
		{URL: "b/", Title: "B", Kind: "skill", Links: nil},
		{URL: "c/", Title: "C", Kind: "agent", Links: []string{"b/", "a/", "b/"}}, // dup b/
		{URL: "d/", Title: "D", Kind: "agent", Links: []string{"d/"}},             // self-link
	}
	g := Build(nodes)

	// B is referenced by A and C (sorted by URL, deduped).
	wantB := []Ref{{URL: "a/", Title: "A", Kind: "agent"}, {URL: "c/", Title: "C", Kind: "agent"}}
	if !reflect.DeepEqual(g.Backlinks["b/"], wantB) {
		t.Errorf("Backlinks[b/] = %+v, want %+v", g.Backlinks["b/"], wantB)
	}
	// A is referenced only by C.
	if got := g.Backlinks["a/"]; len(got) != 1 || got[0].URL != "c/" {
		t.Errorf("Backlinks[a/] = %+v, want [C]", got)
	}
	// Self-link produces no backlink.
	if _, ok := g.Backlinks["d/"]; ok {
		t.Error("self-link should not create a backlink")
	}
	// Target metadata resolved from the node set.
	if g.Backlinks["b/"][0].Title != "A" {
		t.Error("backlink source title should be resolved")
	}
}

func TestBuildEdges(t *testing.T) {
	nodes := []Node{
		{URL: "c/", Title: "C", Links: []string{"b/", "a/"}},
		{URL: "a/", Title: "A", Links: nil},
		{URL: "b/", Title: "B", Links: nil},
	}
	g := Build(nodes)
	// Only C has outgoing links; its targets are sorted by URL.
	if len(g.Edges) != 1 {
		t.Fatalf("got %d edges, want 1", len(g.Edges))
	}
	if g.Edges[0].Source.URL != "c/" {
		t.Errorf("edge source = %q, want c/", g.Edges[0].Source.URL)
	}
	if g.Edges[0].Targets[0].URL != "a/" || g.Edges[0].Targets[1].URL != "b/" {
		t.Errorf("edge targets not sorted: %+v", g.Edges[0].Targets)
	}
}
