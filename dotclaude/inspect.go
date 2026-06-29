package dotclaude

import (
	"path"
	"sort"

	"github.com/gofuego/fuego-dotclaude/dotclaude/classify"
	"github.com/gofuego/fuego-dotclaude/dotclaude/validate"
	"github.com/gofuego/fuego/core"
)

// Capture records the page set during INDEX so the validate and list commands
// can inspect what the build produced without rendering it.
type Capture struct {
	pages []*core.Page
}

// NewCapture returns a fresh Capture.
func NewCapture() *Capture { return &Capture{} }

// Index returns an IndexHook that records the page set.
func (c *Capture) Index() core.IndexHook {
	return func(pages []*core.Page) ([]*core.Page, error) {
		c.pages = pages
		return pages, nil
	}
}

// Pages returns the captured pages.
func (c *Capture) Pages() []*core.Page { return c.pages }

// Diagnostics runs the validator over a captured page set.
func Diagnostics(pages []*core.Page) []validate.Diagnostic {
	idx := validate.Index{
		AgentNames: map[string]bool{},
		MCPServers: map[string]bool{},
	}
	for _, p := range pages {
		switch p.Type {
		case "agent":
			if n := stringOf(p.Envelope["name"]); n != "" {
				idx.AgentNames[n] = true
			}
		case "mcp", "plugin-mcp":
			for _, s := range serverNames(p) {
				idx.MCPServers[s] = true
			}
		}
	}

	var arts []validate.Artifact
	for _, p := range pages {
		if a, ok := toArtifact(p); ok {
			arts = append(arts, a)
		}
	}
	return validate.Validate(arts, idx)
}

// toArtifact projects a page into the validator's view, or reports false for
// pages that aren't validated (virtual/taxonomy/etc.).
func toArtifact(p *core.Page) (validate.Artifact, bool) {
	a := validate.Artifact{
		RelPath:      p.RelPath,
		Name:         stringOf(p.Envelope["name"]),
		Description:  stringOf(p.Envelope["description"]),
		ParseError:   stringOf(p.Envelope["parse_error"]),
		SubagentType: stringOf(p.Envelope["subagent_type"]),
	}
	switch p.Type {
	case "agent", "command":
		a.Kind = p.Type
	case "skill":
		a.Kind = "skill"
		a.SkillDir = path.Base(classify.SkillRoot(p.RelPath))
	case "mcp", "plugin-mcp", "settings", "settings-local", "plugin", "marketplace":
		a.Kind = "json"
		a.MCPRefs = settingsMCPRefs(p)
	default:
		if a.ParseError == "" {
			return validate.Artifact{}, false
		}
		a.Kind = "json"
	}
	return a, true
}

// settingsMCPRefs returns the MCP server names a settings page references.
func settingsMCPRefs(p *core.Page) []string {
	ctrl, ok := p.Envelope["mcp_controls"].(map[string]any)
	if !ok {
		return nil
	}
	var refs []string
	for _, key := range []string{"enabled", "disabled"} {
		if list, ok := ctrl[key].([]string); ok {
			refs = append(refs, list...)
		}
	}
	return refs
}

// ArtifactInfo is a one-line listing of a discovered artifact.
type ArtifactInfo struct {
	Type  string
	Name  string
	Route string
}

// ListArtifacts returns the real (non-virtual) artifacts, sorted by type then
// route, for the list command.
func ListArtifacts(pages []*core.Page) []ArtifactInfo {
	var out []ArtifactInfo
	for _, p := range pages {
		if p.Skip || p.Type == "virtual" || isGeneratedType(p.Type) {
			continue
		}
		out = append(out, ArtifactInfo{
			Type:  p.Type,
			Name:  displayName(p),
			Route: p.URL,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return out[i].Route < out[j].Route
	})
	return out
}

// isGeneratedType reports engine/pack-generated page types that aren't source
// artifacts.
func isGeneratedType(t string) bool {
	switch t {
	case "taxonomy-term", "taxonomy-index", "disambiguation", "reference-graph":
		return true
	}
	return false
}
