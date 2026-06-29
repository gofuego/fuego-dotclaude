// Package scope resolves a CLI path argument and flags into what to build: the
// .claude content directory, whether to fold in root-level sibling files, and
// the usage mode. It supports the two cases from the design — a project repo
// (code + .claude/ + optional sibling CLAUDE.md/.mcp.json) and a .claude
// visualized in isolation — over a small filesystem abstraction so the decision
// logic is table-testable without touching disk. The default (empty argument)
// is the current directory: the common case is a user cd'ing into a project and
// running the tool there.
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

// Resolve maps an argument and an optional siblings override into a Resolution:
//
//   - arg contains .claude/  → <arg>/.claude, project (siblings on)
//   - arg is a .claude dir   → that directory, isolated (no siblings)
//   - arg is any other dir   → that directory treated as the content root,
//                              isolated (the "dedicated .claude repo under a
//                              different name" case, e.g. an ai/ repo)
//   - arg is not a directory  → error
//
// An empty arg means the current directory ("."). A non-nil siblingsFlag (from
// --siblings/--no-siblings) overrides the default.
func Resolve(fsys FS, arg string, siblingsFlag *bool) (Resolution, error) {
	if arg == "" {
		arg = "."
	}

	var r Resolution
	switch {
	case fsys.IsDir(filepath.Join(arg, ".claude")):
		r = Resolution{ContentDir: filepath.Join(arg, ".claude"), Mode: "project", Siblings: true}
	case filepath.Base(arg) == ".claude" && fsys.IsDir(arg):
		r = Resolution{ContentDir: arg, Mode: "isolated"}
	case fsys.IsDir(arg):
		r = Resolution{ContentDir: arg, Mode: "isolated"}
	default:
		return r, fmt.Errorf("%q is not a directory", arg)
	}

	r.SiblingDir = filepath.Dir(r.ContentDir)
	if siblingsFlag != nil {
		r.Siblings = *siblingsFlag
	}
	return r, nil
}
