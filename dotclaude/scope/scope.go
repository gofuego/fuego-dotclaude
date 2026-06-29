// Package scope resolves a CLI path argument and flags into what to build: the
// .claude content directory, whether to fold in root-level sibling files, and
// the usage mode. It supports the two cases from the design — a project repo
// (code + .claude/ + optional sibling CLAUDE.md/.mcp.json) and a .claude
// visualized in isolation — over a small filesystem abstraction so the decision
// logic is table-testable without touching disk.
package scope

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolution is the outcome of resolving a CLI invocation.
type Resolution struct {
	ContentDir string // the .claude directory to build
	SiblingDir string // directory one level up, where root-level siblings live
	Siblings   bool   // whether to inject the root-level siblings
	Mode       string // "isolated" or "project"
}

// FS is the minimal filesystem abstraction scope needs.
type FS interface {
	IsDir(path string) bool
}

// OSFS is the real filesystem.
type OSFS struct{}

func (OSFS) IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// Resolve maps an argument, the user's home directory, and an optional siblings
// override into a Resolution:
//
//   - no arg            → <home>/.claude, isolated (no siblings)
//   - arg is a .claude  → that directory, isolated (no siblings)
//   - arg has .claude/  → <arg>/.claude, project (siblings on)
//   - arg is some dir   → that directory treated as content, isolated
//
// A non-nil siblingsFlag (from --siblings/--no-siblings) overrides the default.
func Resolve(fsys FS, arg, home string, siblingsFlag *bool) (Resolution, error) {
	var r Resolution
	switch {
	case arg == "":
		r = Resolution{ContentDir: filepath.Join(home, ".claude"), Mode: "isolated"}
	case filepath.Base(arg) == ".claude" && fsys.IsDir(arg):
		r = Resolution{ContentDir: arg, Mode: "isolated"}
	case fsys.IsDir(filepath.Join(arg, ".claude")):
		r = Resolution{ContentDir: filepath.Join(arg, ".claude"), Mode: "project", Siblings: true}
	case fsys.IsDir(arg):
		r = Resolution{ContentDir: arg, Mode: "isolated"}
	default:
		return r, fmt.Errorf("no .claude directory found at %q", arg)
	}

	r.SiblingDir = filepath.Dir(r.ContentDir)
	if siblingsFlag != nil {
		r.Siblings = *siblingsFlag
	}
	return r, nil
}
