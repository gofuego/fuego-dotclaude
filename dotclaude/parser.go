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
// emitting it as a single raw "body" node. Frontmatter is returned verbatim in
// the envelope for the hook to classify and the theme to display.
func (p *MarkdownParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env, payload, err := core.SplitFrontmatter(raw)
	if err != nil {
		return nil, nil, err
	}
	if env == nil {
		env = make(core.Envelope)
	}

	var buf bytes.Buffer
	if err := md.Convert(payload, &buf); err != nil {
		return nil, nil, fmt.Errorf("rendering markdown: %w", err)
	}

	nodes := []core.Node{{Type: "body", Content: buf.String(), Raw: true}}
	return env, nodes, nil
}
