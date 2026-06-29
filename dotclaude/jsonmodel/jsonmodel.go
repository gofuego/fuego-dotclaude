// Package jsonmodel decodes the JSON config artifacts of a .claude directory
// (.mcp.json, settings.json, settings.local.json) into structured forms the
// theme can render. It is tolerant: unknown keys are preserved and surfaced
// (Settings.Other) rather than dropped, and malformed input is reported via the
// returned error so callers can degrade to a raw view instead of failing the
// build. It is a deep module — pure decode, no engine types — table-testable in
// isolation.
package jsonmodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// MCPServer is one configured MCP server.
type MCPServer struct {
	Name      string
	Transport string // "stdio" (command-based), "http", or "sse"
	Command   string
	Args      []string
	Env       map[string]string
	URL       string
	Headers   map[string]string
}

// ParseMCP decodes a .mcp.json document into a name-sorted list of servers.
// Transport is inferred when absent: a URL implies "http", otherwise "stdio".
func ParseMCP(raw []byte) ([]MCPServer, error) {
	var f struct {
		MCPServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
			Type    string            `json:"type"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("decoding .mcp.json: %w", err)
	}

	out := make([]MCPServer, 0, len(f.MCPServers))
	for name, s := range f.MCPServers {
		transport := s.Type
		if transport == "" {
			if s.URL != "" {
				transport = "http"
			} else {
				transport = "stdio"
			}
		}
		out = append(out, MCPServer{
			Name:      name,
			Transport: transport,
			Command:   s.Command,
			Args:      s.Args,
			Env:       s.Env,
			URL:       s.URL,
			Headers:   s.Headers,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Permissions is the curated permissions section of settings.
type Permissions struct {
	Allow                 []string
	Deny                  []string
	Ask                   []string
	AdditionalDirectories []string
	DefaultMode           string
}

// Settings is a decoded settings.json / settings.local.json. Curated fields are
// surfaced explicitly; every non-curated key is kept in Other so the long tail
// is never silently dropped. Raw holds the full decoded document.
type Settings struct {
	Model                      string
	Permissions                *Permissions
	Env                        map[string]any
	Hooks                      map[string]any
	StatusLine                 map[string]any
	EnableAllProjectMCPServers *bool
	EnabledMCPJSONServers      []string
	DisabledMCPJSONServers     []string
	Other                      map[string]any
	Raw                        map[string]any
}

// curatedKeys are the settings keys rendered in dedicated sections; everything
// else flows into Settings.Other for the generic key/value table.
var curatedKeys = map[string]bool{
	"model":                      true,
	"permissions":                true,
	"env":                        true,
	"hooks":                      true,
	"statusLine":                 true,
	"enableAllProjectMcpServers": true,
	"enabledMcpjsonServers":      true,
	"disabledMcpjsonServers":     true,
}

// ParseSettings decodes a settings document, splitting it into curated sections
// and the preserved long tail (Other).
func ParseSettings(raw []byte) (*Settings, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decoding settings: %w", err)
	}

	s := &Settings{Raw: m, Other: map[string]any{}}
	s.Model, _ = m["model"].(string)
	s.Env, _ = m["env"].(map[string]any)
	s.Hooks, _ = m["hooks"].(map[string]any)
	s.StatusLine, _ = m["statusLine"].(map[string]any)
	s.EnabledMCPJSONServers = toStrings(m["enabledMcpjsonServers"])
	s.DisabledMCPJSONServers = toStrings(m["disabledMcpjsonServers"])
	if b, ok := m["enableAllProjectMcpServers"].(bool); ok {
		s.EnableAllProjectMCPServers = &b
	}
	if p, ok := m["permissions"].(map[string]any); ok {
		perm := &Permissions{
			Allow:                 toStrings(p["allow"]),
			Deny:                  toStrings(p["deny"]),
			Ask:                   toStrings(p["ask"]),
			AdditionalDirectories: toStrings(p["additionalDirectories"]),
		}
		perm.DefaultMode, _ = p["defaultMode"].(string)
		s.Permissions = perm
	}

	for k, v := range m {
		if !curatedKeys[k] {
			s.Other[k] = v
		}
	}
	return s, nil
}

// PluginManifest is a decoded plugin.json. Author is flattened to a display
// string; Raw holds the full document for the long-tail view.
type PluginManifest struct {
	Name        string
	Version     string
	Description string
	Author      string
	Homepage    string
	License     string
	Keywords    []string
	Raw         map[string]any
}

// ParsePluginManifest decodes a plugin.json, tolerating an author given as a
// string or as an object with a name field.
func ParsePluginManifest(raw []byte) (*PluginManifest, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decoding plugin.json: %w", err)
	}
	pm := &PluginManifest{Raw: m}
	pm.Name, _ = m["name"].(string)
	pm.Version, _ = m["version"].(string)
	pm.Description, _ = m["description"].(string)
	pm.Homepage, _ = m["homepage"].(string)
	pm.License, _ = m["license"].(string)
	pm.Keywords = toStrings(m["keywords"])
	pm.Author = flattenName(m["author"])
	return pm, nil
}

// MarketplacePlugin is one entry in a marketplace listing.
type MarketplacePlugin struct {
	Name        string
	Source      string
	Description string
}

// Marketplace is a decoded marketplace.json.
type Marketplace struct {
	Name    string
	Owner   string
	Plugins []MarketplacePlugin
	Raw     map[string]any
}

// ParseMarketplace decodes a marketplace.json, tolerating an owner given as a
// string or object and a plugins list of objects.
func ParseMarketplace(raw []byte) (*Marketplace, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decoding marketplace.json: %w", err)
	}
	mk := &Marketplace{Raw: m}
	mk.Name, _ = m["name"].(string)
	mk.Owner = flattenName(m["owner"])
	if list, ok := m["plugins"].([]any); ok {
		for _, item := range list {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			p := MarketplacePlugin{}
			p.Name, _ = obj["name"].(string)
			p.Description, _ = obj["description"].(string)
			p.Source = flattenName(obj["source"])
			mk.Plugins = append(mk.Plugins, p)
		}
	}
	return mk, nil
}

// flattenName renders a value that may be a string or an object with a "name"
// field as a display string.
func flattenName(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]any:
		if s, ok := val["name"].(string); ok {
			return s
		}
	}
	return ""
}

// Pretty reformats raw JSON with indentation, or returns the input unchanged if
// it can't be parsed (so a malformed document still shows its source text).
func Pretty(raw []byte) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}

func toStrings(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
