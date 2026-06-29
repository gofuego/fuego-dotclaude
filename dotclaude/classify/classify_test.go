package classify

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    Kind
	}{
		{"top-level agent", "agents/code-reviewer.md", KindAgent},
		{"agent windows separators", `agents\code-reviewer.md`, KindAgent},
		{"agent uppercase extension", "agents/Reviewer.MD", KindAgent},
		{"nested agent", "agents/team/security.md", KindAgent},

		{"skill main file", "skills/code-review/SKILL.md", KindSkill},
		{"skill bundled doc", "skills/code-review/REFERENCE.md", KindSkillDoc},
		{"skill nested doc", "skills/code-review/refs/api.md", KindSkillDoc},
		{"skill bundled asset is not a kind", "skills/code-review/run.sh", KindUnknown},
		{"skill without name dir", "skills/SKILL.md", KindUnknown},

		{"top-level command", "commands/deploy.md", KindCommand},
		{"namespaced command", "commands/git/commit.md", KindCommand},

		{"output style", "output-styles/concise.md", KindOutputStyle},

		{"root memory", "CLAUDE.md", KindMemory},
		{"local memory", "CLAUDE.local.md", KindMemory},
		{"nested memory", "projects/api/CLAUDE.md", KindMemory},

		{"plugin agent", "plugins/acme/agents/foo.md", KindAgent},
		{"plugin skill", "plugins/acme/skills/s/SKILL.md", KindSkill},
		{"plugin skill doc", "plugins/acme/skills/s/REF.md", KindSkillDoc},
		{"plugin command", "plugins/acme/commands/git/commit.md", KindCommand},
		{"plugin memory", "plugins/acme/CLAUDE.md", KindMemory},

		{"agents dir itself", "agents", KindUnknown},
		{"non-md under agents", "agents/notes.txt", KindUnknown},
		{"unrelated md", "notes/scratch.md", KindUnknown},
		{"agents substring trap", "subagents/foo.md", KindUnknown},
		{"empty", "", KindUnknown},
		{"dot", ".", KindUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.relPath); got != tt.want {
				t.Errorf("Classify(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		relPath string
		want    string
	}{
		{"agents/code-reviewer.md", "code-reviewer"},
		{"agents/team/security.md", "team/security"},
		{"skills/code-review/SKILL.md", "code-review"},
		{"skills/code-review/REFERENCE.md", "code-review/REFERENCE"},
		{"skills/code-review/refs/api.md", "code-review/refs/api"},
		{"commands/deploy.md", "deploy"},
		{"commands/git/commit.md", "git/commit"},
		{"output-styles/concise.md", "concise"},
		{"CLAUDE.md", "CLAUDE"},
		{"CLAUDE.local.md", "CLAUDE.local"},
		{"projects/api/CLAUDE.md", "projects/api/CLAUDE"},
		{"notes/scratch.md", ""},
	}
	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			if got := Slug(tt.relPath); got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestCommandName(t *testing.T) {
	tests := []struct {
		relPath string
		want    string
	}{
		{"commands/deploy.md", "/deploy"},
		{"commands/git/commit.md", "/git:commit"},
		{"commands/a/b/c.md", "/a:b:c"},
		{"agents/foo.md", ""},
	}
	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			if got := CommandName(tt.relPath); got != tt.want {
				t.Errorf("CommandName(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestSkillRoot(t *testing.T) {
	tests := []struct {
		relPath string
		want    string
	}{
		{"skills/code-review/SKILL.md", "skills/code-review"},
		{"skills/code-review/REFERENCE.md", "skills/code-review"},
		{"skills/code-review/refs/api.md", "skills/code-review"},
		{"plugins/acme/skills/s/SKILL.md", "plugins/acme/skills/s"},
		{"plugins/acme/skills/s/REF.md", "plugins/acme/skills/s"},
		{"agents/foo.md", ""},
	}
	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			if got := SkillRoot(tt.relPath); got != tt.want {
				t.Errorf("SkillRoot(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestPluginName(t *testing.T) {
	cases := map[string]string{
		"plugins/acme/agents/foo.md":     "acme",
		"plugins/acme/plugin.json":       "acme",
		"plugins/acme/.claude-plugin/plugin.json": "acme",
		"agents/foo.md":                  "",
		"plugins/acme":                   "", // needs at least one inner segment
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			if got := PluginName(in); got != want {
				t.Errorf("PluginName(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestPluginSlugIsSectionRelative(t *testing.T) {
	// The plugin namespace is added by the caller; Slug stays section-relative.
	cases := map[string]string{
		"plugins/acme/agents/foo.md":          "foo",
		"plugins/acme/skills/s/SKILL.md":       "s",
		"plugins/acme/skills/s/REF.md":         "s/REF",
		"plugins/acme/commands/git/commit.md":  "git/commit",
	}
	for in, want := range cases {
		t.Run(in, func(t *testing.T) {
			if got := Slug(in); got != want {
				t.Errorf("Slug(%q) = %q, want %q", in, got, want)
			}
		})
	}
}
