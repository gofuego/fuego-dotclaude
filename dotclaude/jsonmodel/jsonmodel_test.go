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

func TestParsePluginManifest(t *testing.T) {
	raw := []byte(`{
	  "name": "acme-tools",
	  "version": "1.2.0",
	  "description": "Handy tools.",
	  "author": { "name": "Acme", "email": "x@acme.dev" },
	  "keywords": ["dev", "tools"],
	  "license": "MIT"
	}`)
	pm, err := ParsePluginManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if pm.Name != "acme-tools" || pm.Version != "1.2.0" {
		t.Errorf("name/version not decoded: %+v", pm)
	}
	if pm.Author != "Acme" {
		t.Errorf("author object should flatten to name, got %q", pm.Author)
	}
	if len(pm.Keywords) != 2 {
		t.Errorf("keywords not decoded: %v", pm.Keywords)
	}
}

func TestParseMarketplace(t *testing.T) {
	raw := []byte(`{
	  "name": "acme-market",
	  "owner": "Acme Inc",
	  "plugins": [
	    { "name": "acme-tools", "source": "./acme-tools", "description": "Tools." },
	    { "name": "acme-lint" }
	  ]
	}`)
	mk, err := ParseMarketplace(raw)
	if err != nil {
		t.Fatal(err)
	}
	if mk.Name != "acme-market" || mk.Owner != "Acme Inc" {
		t.Errorf("name/owner not decoded: %+v", mk)
	}
	if len(mk.Plugins) != 2 || mk.Plugins[0].Name != "acme-tools" {
		t.Errorf("plugins not decoded: %+v", mk.Plugins)
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
