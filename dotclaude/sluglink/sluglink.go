// Package sluglink is the cross-reference engine: it builds a registry of named
// artifacts and rewrites rendered page HTML so every whole-word mention of an
// artifact's name becomes a link to that artifact's page.
//
// It is the project's hardest correctness surface, so it walks the HTML's token
// stream rather than naively regex-replacing the markup: matches are made only
// in prose and inline <code>, never inside fenced <pre> blocks or text already
// inside an <a>. Matching is whole-word and case-insensitive, guarded by a
// stopword list, and skips the page's own name. Ambiguous names (shared by more
// than one artifact) link to a disambiguation route. The package is pure (no
// engine types, no I/O) and table-testable in isolation.
package sluglink

import (
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// Target is one artifact a name can resolve to.
type Target struct {
	Name string // canonical display name
	URL  string // base-relative URL of the artifact page
	Kind string // artifact kind, for disambiguation display
}

// Link records a cross-reference emitted while rewriting a page, for callers
// that build backlinks.
type Link struct {
	Slug      string // normalized name that matched
	URL       string // resolved target (disambiguation URL when ambiguous)
	Ambiguous bool
}

// Registry holds the name → target(s) index.
type Registry struct {
	byKey     map[string][]Target
	stop      map[string]bool
	matchKeys []string // finalized, stopword-filtered, longest-first
}

// NewRegistry returns an empty registry guarded by the given stopwords (matched
// case-insensitively).
func NewRegistry(stopwords []string) *Registry {
	stop := make(map[string]bool, len(stopwords))
	for _, w := range stopwords {
		stop[normalize(w)] = true
	}
	return &Registry{byKey: map[string][]Target{}, stop: stop}
}

// Add registers a target under its normalized name. Empty or stopword names are
// ignored. Adding invalidates the finalized match set.
func (r *Registry) Add(t Target) {
	key := normalize(t.Name)
	if key == "" || r.stop[key] {
		return
	}
	r.byKey[key] = append(r.byKey[key], t)
	r.matchKeys = nil
}

// Resolve returns the targets registered under a normalized name.
func (r *Registry) Resolve(key string) ([]Target, bool) {
	t, ok := r.byKey[key]
	return t, ok
}

// Collisions returns the normalized names that resolve to more than one target,
// sorted.
func (r *Registry) Collisions() []string {
	var out []string
	for k, ts := range r.byKey {
		if len(ts) > 1 {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// finalize computes the match key order: longest first (so the longest slug
// wins an overlap), then lexical for determinism.
func (r *Registry) finalize() {
	if r.matchKeys != nil {
		return
	}
	keys := make([]string, 0, len(r.byKey))
	for k := range r.byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	r.matchKeys = keys
}

// Link rewrites an HTML fragment, linking whole-word slug mentions found in
// prose and inline code. selfName is the current page's own name (skipped to
// avoid self-links). It returns the rewritten HTML and the deduped set of links
// emitted.
func (r *Registry) Link(fragment, selfName string) (string, []Link, error) {
	r.finalize()
	selfKey := normalize(selfName)

	z := html.NewTokenizer(strings.NewReader(fragment))
	var b strings.Builder
	var links []Link
	aDepth, preDepth := 0, 0

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			}
			return "", nil, z.Err()
		}

		switch tt {
		case html.TextToken:
			if aDepth > 0 || preDepth > 0 {
				b.Write(z.Raw()) // inside a link or fenced block: pass through verbatim
				continue
			}
			out, ls := r.linkText(string(z.Text()), selfKey)
			b.WriteString(out)
			links = append(links, ls...)

		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := z.TagName()
			if tt == html.StartTagToken {
				switch string(name) {
				case "a":
					aDepth++
				case "pre":
					preDepth++
				}
			}
			b.Write(z.Raw())

		case html.EndTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "a":
				if aDepth > 0 {
					aDepth--
				}
			case "pre":
				if preDepth > 0 {
					preDepth--
				}
			}
			b.Write(z.Raw())

		default:
			b.Write(z.Raw())
		}
	}

	return b.String(), dedupe(links), nil
}

// linkText finds and wraps whole-word slug matches in a run of (unescaped) text,
// re-escaping the surrounding text. It resolves overlaps by preferring the
// earliest, then longest, match.
func (r *Registry) linkText(text, selfKey string) (string, []Link) {
	lower := strings.ToLower(text)

	type match struct {
		start, end int
		key        string
	}
	var matches []match
	for _, key := range r.matchKeys {
		if key == selfKey {
			continue
		}
		from := 0
		for {
			idx := strings.Index(lower[from:], key)
			if idx < 0 {
				break
			}
			s := from + idx
			e := s + len(key)
			if isWordBoundary(text, s, e) {
				matches = append(matches, match{s, e, key})
			}
			from = s + 1
		}
	}
	if len(matches) == 0 {
		return html.EscapeString(text), nil
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].start != matches[j].start {
			return matches[i].start < matches[j].start
		}
		return matches[i].end > matches[j].end // longer first on a tie
	})

	var b strings.Builder
	var links []Link
	pos := 0
	for _, m := range matches {
		if m.start < pos {
			continue // overlaps an already-picked match
		}
		b.WriteString(html.EscapeString(text[pos:m.start]))

		targets, _ := r.Resolve(m.key)
		ambiguous := len(targets) > 1
		url := ""
		if ambiguous {
			url = DisambiguationURL(m.key)
		} else if len(targets) == 1 {
			url = targets[0].URL
		}

		b.WriteString(`<a href="`)
		b.WriteString(html.EscapeString(url))
		b.WriteString(`" class="slug-link`)
		if ambiguous {
			b.WriteString(" slug-ambiguous")
		}
		b.WriteString(`" data-slug="`)
		b.WriteString(html.EscapeString(m.key))
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(text[m.start:m.end]))
		b.WriteString(`</a>`)

		links = append(links, Link{Slug: m.key, URL: url, Ambiguous: ambiguous})
		pos = m.end
	}
	b.WriteString(html.EscapeString(text[pos:]))
	return b.String(), links
}

// DisambiguationSlug sanitizes a normalized name into a URL-safe slug.
func DisambiguationSlug(key string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(key) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// DisambiguationURL is the base-relative route for a name's disambiguation page.
func DisambiguationURL(key string) string {
	return "disambiguation/" + DisambiguationSlug(key) + "/"
}

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func isWordBoundary(text string, s, e int) bool {
	before := true
	if s > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:s])
		before = !isWordRune(r)
	}
	after := true
	if e < len(text) {
		r, _ := utf8.DecodeRuneInString(text[e:])
		after = !isWordRune(r)
	}
	return before && after
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
}

func dedupe(links []Link) []Link {
	seen := map[string]bool{}
	var out []Link
	for _, l := range links {
		if seen[l.URL] {
			continue
		}
		seen[l.URL] = true
		out = append(out, l)
	}
	return out
}
