package validate

import (
	"strings"
	"testing"
)

func hasDiag(diags []Diagnostic, relPath, substr string) bool {
	for _, d := range diags {
		if d.RelPath == relPath && strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func TestValidate(t *testing.T) {
	idx := Index{
		AgentNames: map[string]bool{"reviewer": true},
		MCPServers: map[string]bool{"github": true},
	}
	arts := []Artifact{
		{Kind: "agent", RelPath: "agents/a.md"},                                              // missing name + description
		{Kind: "agent", RelPath: "agents/b.md", Name: "b", Description: "ok"},                 // clean
		{Kind: "skill", RelPath: "skills/x/SKILL.md", Name: "y", Description: "d", SkillDir: "x"}, // name != dir
		{Kind: "command", RelPath: "commands/c.md"},                                          // missing description
		{Kind: "agent", RelPath: "agents/d.md", Name: "d", Description: "d", SubagentType: "ghost"}, // dangling subagent
		{Kind: "mcp", RelPath: ".mcp.json", ParseError: "unexpected end of input"},           // malformed JSON
		{Kind: "settings", RelPath: "settings.json", MCPRefs: []string{"github", "missing"}}, // one dangling MCP ref
	}

	diags := Validate(arts, idx)

	cases := []struct{ path, substr string }{
		{"agents/a.md", "missing required frontmatter: name"},
		{"agents/a.md", "missing required frontmatter: description"},
		{"skills/x/SKILL.md", "does not match its directory"},
		{"commands/c.md", "missing required frontmatter: description"},
		{"agents/d.md", "references unknown subagent"},
		{".mcp.json", "malformed JSON"},
		{"settings.json", "references unknown MCP server \"missing\""},
	}
	for _, c := range cases {
		if !hasDiag(diags, c.path, c.substr) {
			t.Errorf("expected diagnostic for %s containing %q", c.path, c.substr)
		}
	}

	// The clean agent and the resolvable MCP ref produce no diagnostics.
	if hasDiag(diags, "agents/b.md", "") {
		t.Error("clean agent should produce no diagnostics")
	}
	if hasDiag(diags, "settings.json", "\"github\"") {
		t.Error("resolvable MCP ref should not be flagged")
	}
}

func TestValidateClean(t *testing.T) {
	arts := []Artifact{
		{Kind: "agent", RelPath: "agents/a.md", Name: "a", Description: "d"},
		{Kind: "skill", RelPath: "skills/s/SKILL.md", Name: "s", Description: "d", SkillDir: "s"},
	}
	if diags := Validate(arts, Index{}); len(diags) != 0 {
		t.Errorf("expected no diagnostics, got %v", diags)
	}
}
