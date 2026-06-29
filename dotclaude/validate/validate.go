// Package validate runs advisory coherence checks over a .claude directory's
// artifacts: required frontmatter, skill name vs directory, dangling explicit
// references (subagent types, MCP servers), and malformed JSON. It works over a
// small artifact representation independent of the engine, so the checks are
// testable in isolation. Diagnostics are advisory — the site still renders; a
// caller may choose to treat any diagnostic as fatal (CI --strict).
package validate

import "sort"

// Artifact is the validator's view of one parsed artifact.
type Artifact struct {
	Kind         string
	RelPath      string
	Name         string   // frontmatter name (agents, skills, output styles)
	Description  string   // frontmatter description
	SkillDir     string   // a skill's directory name, to compare against Name
	ParseError   string   // non-empty if the source failed to parse
	SubagentType string   // an explicit reference to another agent, if declared
	MCPRefs      []string // MCP server names this artifact references
}

// Index is the known universe used to resolve references.
type Index struct {
	AgentNames map[string]bool
	MCPServers map[string]bool
}

// Diagnostic is one advisory finding, attributed to a source path.
type Diagnostic struct {
	RelPath string
	Message string
}

// Validate returns the diagnostics for the given artifacts, sorted by path then
// message for deterministic output.
func Validate(arts []Artifact, idx Index) []Diagnostic {
	var diags []Diagnostic
	add := func(relPath, msg string) { diags = append(diags, Diagnostic{relPath, msg}) }

	for _, a := range arts {
		if a.ParseError != "" {
			add(a.RelPath, "malformed JSON: "+a.ParseError)
		}

		switch a.Kind {
		case "agent":
			requireName(add, a)
			requireDescription(add, a)
		case "skill":
			requireName(add, a)
			requireDescription(add, a)
			if a.Name != "" && a.SkillDir != "" && a.Name != a.SkillDir {
				add(a.RelPath, "skill name "+quote(a.Name)+" does not match its directory "+quote(a.SkillDir))
			}
		case "command":
			requireDescription(add, a)
		}

		if a.SubagentType != "" && !idx.AgentNames[a.SubagentType] {
			add(a.RelPath, "references unknown subagent "+quote(a.SubagentType))
		}
		for _, ref := range a.MCPRefs {
			if !idx.MCPServers[ref] {
				add(a.RelPath, "references unknown MCP server "+quote(ref))
			}
		}
	}

	sort.Slice(diags, func(i, j int) bool {
		if diags[i].RelPath != diags[j].RelPath {
			return diags[i].RelPath < diags[j].RelPath
		}
		return diags[i].Message < diags[j].Message
	})
	return diags
}

func requireName(add func(string, string), a Artifact) {
	if a.Name == "" {
		add(a.RelPath, "missing required frontmatter: name")
	}
}

func requireDescription(add func(string, string), a Artifact) {
	if a.Description == "" {
		add(a.RelPath, "missing required frontmatter: description")
	}
}

func quote(s string) string { return "\"" + s + "\"" }
