package jsonmodel

import "testing"

func TestParseMCP(t *testing.T) {
	raw := []byte(`{
	  "mcpServers": {
	    "github": { "command": "npx", "args": ["-y", "server-github"], "env": {"TOKEN": "x"} },
	    "api": { "url": "https://example.com/mcp", "type": "sse" },
	    "fetch": { "url": "https://fetch.example" }
	  }
	}`)
	servers, err := ParseMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 3 {
		t.Fatalf("got %d servers, want 3", len(servers))
	}
	// Sorted by name: api, fetch, github.
	if servers[0].Name != "api" || servers[2].Name != "github" {
		t.Errorf("servers not sorted by name: %v", []string{servers[0].Name, servers[1].Name, servers[2].Name})
	}
	if servers[2].Transport != "stdio" {
		t.Errorf("github transport = %q, want stdio", servers[2].Transport)
	}
	if servers[0].Transport != "sse" {
		t.Errorf("api transport = %q, want sse (explicit type)", servers[0].Transport)
	}
	if servers[1].Transport != "http" {
		t.Errorf("fetch transport = %q, want http (inferred from url)", servers[1].Transport)
	}
	if servers[2].Env["TOKEN"] != "x" {
		t.Errorf("github env not decoded: %v", servers[2].Env)
	}
}

func TestParseMCPMalformed(t *testing.T) {
	if _, err := ParseMCP([]byte(`{not json`)); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseSettings(t *testing.T) {
	raw := []byte(`{
	  "model": "opus",
	  "permissions": { "allow": ["Bash(ls)"], "deny": ["Read(./secret)"], "defaultMode": "acceptEdits" },
	  "env": { "FOO": "bar" },
	  "enableAllProjectMcpServers": true,
	  "enabledMcpjsonServers": ["github"],
	  "cleanupPeriodDays": 30,
	  "includeCoAuthoredBy": false
	}`)
	s, err := ParseSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Model != "opus" {
		t.Errorf("model = %q, want opus", s.Model)
	}
	if s.Permissions == nil || len(s.Permissions.Allow) != 1 || s.Permissions.DefaultMode != "acceptEdits" {
		t.Errorf("permissions not decoded: %+v", s.Permissions)
	}
	if s.EnableAllProjectMCPServers == nil || !*s.EnableAllProjectMCPServers {
		t.Error("enableAllProjectMcpServers not decoded")
	}
	if len(s.EnabledMCPJSONServers) != 1 || s.EnabledMCPJSONServers[0] != "github" {
		t.Errorf("enabledMcpjsonServers not decoded: %v", s.EnabledMCPJSONServers)
	}
	// Unknown keys are preserved, not dropped.
	if _, ok := s.Other["cleanupPeriodDays"]; !ok {
		t.Error("cleanupPeriodDays (long-tail key) should be preserved in Other")
	}
	if _, ok := s.Other["includeCoAuthoredBy"]; !ok {
		t.Error("includeCoAuthoredBy (long-tail key) should be preserved in Other")
	}
	// Curated keys do not leak into Other.
	if _, ok := s.Other["model"]; ok {
		t.Error("curated key model should not appear in Other")
	}
}

func TestParseSettingsMalformed(t *testing.T) {
	if _, err := ParseSettings([]byte(`nope`)); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestPretty(t *testing.T) {
	if got := Pretty([]byte(`{"a":1}`)); got == `{"a":1}` {
		t.Errorf("Pretty did not reformat: %q", got)
	}
	// Malformed input is returned unchanged.
	if got := Pretty([]byte(`{bad`)); got != `{bad` {
		t.Errorf("Pretty(malformed) = %q, want unchanged", got)
	}
}
