// Package classify maps a content-dir-relative path inside a .claude directory
// to the Claude Code artifact kind it represents, plus the routing identity
// (slug, command namespace, skill root) derived from that path.
//
// Fuego dispatches parsers by file extension or basename glob only, and a
// parser's Parse receives no path — so every Markdown artifact (agents, skills,
// commands, output styles, memory) is indistinguishable at parse time. This
// package is the single source of truth that recovers an artifact's kind and
// identity from its location in the tree, consumed by the pack's hooks after
// PARSE but before ROUTE. Everything here is a pure function of the path: no
// I/O, no engine types, table-testable in isolation.
package classify

import (
	"path"
	"strings"
)

// Kind is a Claude Code artifact kind. Its string value doubles as the page's
// content type (driving route + layout selection) once the hook assigns it.
type Kind string

const (
	// KindAgent is a subagent definition: agents/<name>.md.
	KindAgent Kind = "agent"

	// KindSkill is a skill's main definition file: skills/<name>/SKILL.md.
	KindSkill Kind = "skill"

	// KindSkillDoc is a bundled Markdown file inside a skill directory
	// (anything under skills/<name>/ that is not SKILL.md). It renders as a
	// sub-page of its skill.
	KindSkillDoc Kind = "skill-doc"

	// KindCommand is a slash command: commands/**/*.md, namespaced by
	// subdirectory.
	KindCommand Kind = "command"

	// KindOutputStyle is an output style: output-styles/<name>.md.
	KindOutputStyle Kind = "output-style"

	// KindMemory is a memory file: CLAUDE.md or CLAUDE.local.md at any depth.
	KindMemory Kind = "memory"

	// KindUnknown is any path this package does not recognize. The hook leaves
	// such pages with their default type so they still render generically.
	KindUnknown Kind = "unknown"
)

// Classify returns the artifact Kind for a content-dir-relative path (relative
// to the .claude directory). Paths inside plugins/<name>/ classify by their
// inner artifact, so a plugin's agents/skills/commands reuse the same kinds
// (and layouts) as project artifacts. An unrecognized path returns KindUnknown.
func Classify(relPath string) Kind {
	segments, _, ok := segs(relPath)
	if !ok {
		return KindUnknown
	}
	segments, _ = effective(segments)
	if len(segments) == 0 {
		return KindUnknown
	}
	base := segments[len(segments)-1]

	switch segments[0] {
	case "agents":
		if len(segments) >= 2 && isMarkdown(base) {
			return KindAgent
		}
	case "skills":
		if len(segments) >= 3 {
			if base == "SKILL.md" {
				return KindSkill
			}
			if isMarkdown(base) {
				return KindSkillDoc
			}
		}
	case "commands":
		if len(segments) >= 2 && isMarkdown(base) {
			return KindCommand
		}
	case "output-styles":
		if len(segments) >= 2 && isMarkdown(base) {
			return KindOutputStyle
		}
	}

	if isMemory(base) {
		return KindMemory
	}
	return KindUnknown
}

// PluginName returns the plugin a path belongs to (plugins/<name>/…), or "" for
// paths outside a plugin.
func PluginName(relPath string) string {
	segments, _, ok := segs(relPath)
	if !ok {
		return ""
	}
	_, plugin := effective(segments)
	return plugin
}

// effective strips a leading plugins/<name>/ prefix, returning the inner path
// segments and the plugin name ("" when not under a plugin).
func effective(segments []string) (inner []string, plugin string) {
	if len(segments) >= 3 && segments[0] == "plugins" {
		return segments[2:], segments[1]
	}
	return segments, ""
}

// Slug returns the section-relative routing slug for a recognized path, or "" for
// KindUnknown. Slugs may contain "/" so artifacts nest under their section route.
// For plugin artifacts this is the inner slug; callers prepend the plugin
// namespace. Derived from the path, so files in different directories never
// collide.
func Slug(relPath string) string {
	segments, _, ok := segs(relPath)
	if !ok {
		return ""
	}
	segments, _ = effective(segments)
	switch Classify(relPath) {
	case KindAgent, KindSkillDoc, KindCommand, KindOutputStyle:
		// Path under the section directory, extension dropped.
		return dropMD(strings.Join(segments[1:], "/"))
	case KindSkill:
		// The skill directory path between skills/ and SKILL.md.
		return strings.Join(segments[1:len(segments)-1], "/")
	case KindMemory:
		// Full path so nested CLAUDE.md files stay distinct.
		return dropMD(strings.Join(segments, "/"))
	}
	return ""
}

// CommandName returns a command's display name in /namespace:command form,
// derived from its subdirectories. Returns "" for non-command paths.
func CommandName(relPath string) string {
	if Classify(relPath) != KindCommand {
		return ""
	}
	segments, _, _ := segs(relPath)
	segments, _ = effective(segments)
	parts := segments[1:] // drop the "commands" prefix
	parts[len(parts)-1] = dropMD(parts[len(parts)-1])
	return "/" + strings.Join(parts, ":")
}

// SkillRoot returns the directory a skill or skill-doc belongs to, used to group
// a skill with its bundled docs and assets. Plugin skills keep their plugin
// prefix so same-named skills in different plugins stay distinct. Returns "" when
// the path is not under a skill.
func SkillRoot(relPath string) string {
	segments, _, ok := segs(relPath)
	if !ok {
		return ""
	}
	inner, plugin := effective(segments)
	if len(inner) < 2 || inner[0] != "skills" {
		return ""
	}
	root := "skills/" + inner[1]
	if plugin != "" {
		return "plugins/" + plugin + "/" + root
	}
	return root
}

// segs normalizes a path and splits it into segments, also returning the base
// name. ok is false for empty or "." paths.
func segs(relPath string) (segments []string, base string, ok bool) {
	rel := path.Clean(strings.ReplaceAll(relPath, "\\", "/"))
	if rel == "." || rel == "" {
		return nil, "", false
	}
	segments = strings.Split(rel, "/")
	return segments, segments[len(segments)-1], true
}

func isMarkdown(name string) bool { return strings.HasSuffix(strings.ToLower(name), ".md") }

func isMemory(base string) bool { return base == "CLAUDE.md" || base == "CLAUDE.local.md" }

// dropMD trims a trailing .md (case-insensitive) from a path or name.
func dropMD(s string) string {
	if strings.HasSuffix(strings.ToLower(s), ".md") {
		return s[:len(s)-len(".md")]
	}
	return s
}
