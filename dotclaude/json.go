package dotclaude

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/gofuego/fuego-dotclaude/dotclaude/jsonmodel"
	"github.com/gofuego/fuego/core"
)

// MCPParser renders a .mcp.json document as a server-per-card page. Each server
// name is published in the envelope (mcp_server_names) so the slug registry can
// turn mentions of it elsewhere into links to this page.
type MCPParser struct{}

func (p *MCPParser) Type() string        { return "mcp" }
func (p *MCPParser) Filenames() []string { return []string{".mcp.json"} }

func (p *MCPParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	env := core.Envelope{
		"title":    "MCP Servers",
		"layout":   "mcp",
		"source":   "project",
		"raw_json": jsonmodel.Pretty(raw),
	}

	servers, err := jsonmodel.ParseMCP(raw)
	if err != nil {
		env["parse_error"] = err.Error()
		return env, nil, nil
	}

	view := make([]map[string]any, 0, len(servers))
	names := make([]string, 0, len(servers))
	for _, s := range servers {
		view = append(view, map[string]any{
			"name":      s.Name,
			"transport": s.Transport,
			"command":   s.Command,
			"args":      s.Args,
			"url":       s.URL,
			"env":       sortedStringKV(s.Env),
			"headers":   sortedStringKV(s.Headers),
		})
		names = append(names, s.Name)
	}
	env["mcp_servers"] = view
	env["mcp_server_names"] = names
	return env, nil, nil
}

// SettingsParser renders a settings.json or settings.local.json document. The
// local flag selects the filename, route type, and title; both share the
// "settings" layout.
type SettingsParser struct{ local bool }

func (p *SettingsParser) Type() string {
	if p.local {
		return "settings-local"
	}
	return "settings"
}

func (p *SettingsParser) Filenames() []string {
	if p.local {
		return []string{"settings.local.json"}
	}
	return []string{"settings.json"}
}

func (p *SettingsParser) Parse(raw []byte) (core.Envelope, []core.Node, error) {
	title := "Settings"
	if p.local {
		title = "Local Settings"
	}
	env := core.Envelope{
		"title":    title,
		"layout":   "settings",
		"source":   "project",
		"raw_json": jsonmodel.Pretty(raw),
	}

	s, err := jsonmodel.ParseSettings(raw)
	if err != nil {
		env["parse_error"] = err.Error()
		return env, nil, nil
	}

	if s.Model != "" {
		env["model"] = s.Model
	}
	if s.Permissions != nil {
		env["permissions"] = map[string]any{
			"allow":        s.Permissions.Allow,
			"deny":         s.Permissions.Deny,
			"ask":          s.Permissions.Ask,
			"additional":   s.Permissions.AdditionalDirectories,
			"default_mode": s.Permissions.DefaultMode,
		}
	}
	env["env_vars"] = sortedAnyKV(s.Env)
	env["mcp_controls"] = map[string]any{
		"enable_all": s.EnableAllProjectMCPServers,
		"enabled":    s.EnabledMCPJSONServers,
		"disabled":   s.DisabledMCPJSONServers,
	}
	if len(s.Hooks) > 0 {
		env["hooks_json"] = marshalPretty(s.Hooks)
	}
	if len(s.StatusLine) > 0 {
		env["status_line_json"] = marshalPretty(s.StatusLine)
	}
	env["other_settings"] = sortedAnyKV(s.Other)
	return env, nil, nil
}

// kv is one row of a key/value display table.
type kv struct {
	Key   string
	Value string
}

// sortedStringKV turns a string map into key-sorted rows.
func sortedStringKV(m map[string]string) []kv {
	if len(m) == 0 {
		return nil
	}
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{Key: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// sortedAnyKV turns an arbitrary map into key-sorted rows, rendering non-scalar
// values as compact JSON so the long-tail table can show anything.
func sortedAnyKV(m map[string]any) []kv {
	if len(m) == 0 {
		return nil
	}
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{Key: k, Value: valueString(v)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func valueString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		// JSON numbers decode as float64; print integers without a trailing .0.
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func marshalPretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
