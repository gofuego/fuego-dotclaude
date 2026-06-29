package sluglink

import (
	"strings"
	"testing"
)

func reg(t *testing.T) *Registry {
	r := NewRegistry([]string{"the", "code"})
	r.Add(Target{Name: "code-reviewer", URL: "agents/code-reviewer/", Kind: "agent"})
	r.Add(Target{Name: "doc-writer", URL: "agents/doc-writer/", Kind: "agent"})
	return r
}

func TestLinkTable(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		self     string
		contains []string // substrings that must appear
		absent   []string // substrings that must NOT appear
	}{
		{
			name:     "prose match",
			in:       "<p>Ask the code-reviewer to look.</p>",
			contains: []string{`<a href="agents/code-reviewer/"`, `data-slug="code-reviewer"`, ">code-reviewer</a>"},
		},
		{
			name:     "inline code match",
			in:       "<p>Run <code>code-reviewer</code> now.</p>",
			contains: []string{`<code><a href="agents/code-reviewer/"`},
		},
		{
			name:   "fenced block skipped",
			in:     "<pre><code>code-reviewer here</code></pre>",
			absent: []string{"<a href"},
		},
		{
			name:   "existing link skipped",
			in:     `<p><a href="x">code-reviewer</a></p>`,
			absent: []string{"slug-link"},
		},
		{
			name:   "self slug skipped",
			in:     "<p>I am code-reviewer.</p>",
			self:   "code-reviewer",
			absent: []string{"<a href"},
		},
		{
			name:     "case insensitive",
			in:       "<p>The Code-Reviewer agent.</p>",
			contains: []string{`<a href="agents/code-reviewer/"`, ">Code-Reviewer</a>"},
		},
		{
			name:   "whole word boundary",
			in:     "<p>code-reviewerish should not match.</p>",
			absent: []string{"<a href"},
		},
		{
			name:   "stopword excluded",
			in:     "<p>the code is here.</p>",
			absent: []string{"<a href"},
		},
		{
			name:     "html not corrupted",
			in:       `<p class="x" data-y="z">doc-writer &amp; more</p>`,
			contains: []string{`<p class="x" data-y="z">`, "&amp;", `<a href="agents/doc-writer/"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, err := reg(t).Link(tt.in, tt.self)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
			for _, bad := range tt.absent {
				if strings.Contains(out, bad) {
					t.Errorf("output %q should not contain %q", out, bad)
				}
			}
		})
	}
}

func TestAmbiguousLinksToDisambiguation(t *testing.T) {
	r := NewRegistry(nil)
	r.Add(Target{Name: "review", URL: "agents/review/", Kind: "agent"})
	r.Add(Target{Name: "review", URL: "commands/review/", Kind: "command"})

	cols := r.Collisions()
	if len(cols) != 1 || cols[0] != "review" {
		t.Fatalf("collisions = %v, want [review]", cols)
	}

	out, links, err := r.Link("<p>Run review now.</p>", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `href="disambiguation/review/"`) {
		t.Errorf("ambiguous match should link to disambiguation route: %q", out)
	}
	if !strings.Contains(out, "slug-ambiguous") {
		t.Error("ambiguous link should carry the slug-ambiguous class")
	}
	if len(links) != 1 || !links[0].Ambiguous {
		t.Errorf("expected one ambiguous link, got %+v", links)
	}
}

func TestLongestMatchWins(t *testing.T) {
	r := NewRegistry(nil)
	r.Add(Target{Name: "review", URL: "a/review/", Kind: "agent"})
	r.Add(Target{Name: "code-review", URL: "skills/code-review/", Kind: "skill"})

	out, _, err := r.Link("<p>Use code-review here.</p>", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `href="skills/code-review/"`) {
		t.Errorf("longest slug should win the overlap: %q", out)
	}
}

func TestDisambiguationSlug(t *testing.T) {
	cases := map[string]string{
		"git:commit":    "git-commit",
		"code-reviewer": "code-reviewer",
		"a/b":           "a-b",
		"  weird !! ":   "weird",
	}
	for in, want := range cases {
		if got := DisambiguationSlug(in); got != want {
			t.Errorf("DisambiguationSlug(%q) = %q, want %q", in, got, want)
		}
	}
}
