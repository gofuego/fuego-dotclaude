// Package dotclaude is the fuego-dotclaude format pack: it turns the Markdown
// and JSON artifacts inside a .claude directory into a navigable, cross-
// referenced documentation site. Register it on any Fuego engine with
// eng.Use(dotclaude.Pack()) — it brings its own parser, theme, routes, and the
// classification hook, so a vanilla Fuego project needs only a .claude
// directory and one line of wiring.
package dotclaude

import (
	"bytes"
	"fmt"

	"github.com/gofuego/fuego/core"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// MarkdownParser is the generic parser for every .md artifact in a .claude
// directory. Because Fuego dispatches by extension only, one parser handles
// agents, skills, commands, output styles, and memory files alike; the
// per-artifact distinction is recovered later by the classification hook from
// each page's RelPath (see package classify).
type MarkdownParser struct{}

// NewParser returns a new MarkdownParser.
func NewParser() *MarkdownParser { return &MarkdownParser{} }

// Type registers the parser under the "md" extension.
func (p *MarkdownParser) Type() string { return "md" }

// Parse splits YAML frontmatter from the body and renders the body to HTML,
// emitting it as a single raw "body" node. Frontmatter is returned in the
// envelope for the hook to classify and the theme to display.
//
// Parsing is tolerant: Claude Code agent frontmatter frequently contains
// unquoted descriptions with colons that aren't valid YAML. Rather than fail the
// whole build over one messy file, an unparseable frontmatter block degrades to
// an empty envelope, the raw block is preserved (frontmatter_raw) and flagged
// (frontmatter_error) so the page still renders and surfaces a warning.
func (p *MarkdownParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env, payload, warn, rawFM := splitFrontmatterTolerant(raw)

	var buf bytes.Buffer
	if err := md.Convert(payload, &buf); err != nil {
		return nil, nil, fmt.Errorf("rendering markdown: %w", err)
	}

	if warn != "" {
		env["frontmatter_error"] = warn
		env["frontmatter_raw"] = rawFM
	}

	nodes := []core.Node{{Type: "body", Content: buf.String(), Raw: true}}
	return env, nodes, nil
}

// splitFrontmatterTolerant tries strict YAML frontmatter parsing and, on
// failure, recovers the body without parsing — returning an empty envelope, a
// warning, and the raw frontmatter text. A file with no frontmatter parses
// cleanly with an empty envelope.
func splitFrontmatterTolerant(raw []byte) (env core.Envelope, body []byte, warn, rawFM string) {
	env, body, err := core.SplitFrontmatter(raw)
	if err == nil {
		if env == nil {
			env = make(core.Envelope)
		}
		return env, body, "", ""
	}

	fm, rest, ok := rawFrontmatter(raw)
	if !ok {
		// Couldn't even locate a frontmatter block; treat the whole file as body.
		return make(core.Envelope), raw, "", ""
	}
	return make(core.Envelope), rest, "frontmatter is not valid YAML: " + err.Error(), string(fm)
}

// rawFrontmatter locates a leading "---" frontmatter block and returns its raw
// text and the body, mirroring core.SplitFrontmatter's boundary logic but
// without YAML parsing.
func rawFrontmatter(raw []byte) (fm, body []byte, ok bool) {
	t := bytes.TrimLeft(raw, " \t\r\n")
	if !bytes.HasPrefix(t, []byte("---")) {
		return nil, nil, false
	}
	rest := t[3:]
	if i := bytes.IndexByte(rest, '\n'); i >= 0 {
		rest = rest[i+1:]
	} else {
		return nil, nil, false
	}
	closeIdx := bytes.Index(rest, []byte("---"))
	if closeIdx < 0 {
		return nil, nil, false
	}
	fm = rest[:closeIdx]
	body = rest[closeIdx+3:]
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	} else if len(body) > 1 && body[0] == '\r' && body[1] == '\n' {
		body = body[2:]
	}
	return fm, body, true
}
