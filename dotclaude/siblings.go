package dotclaude

import (
	"os"
	"path/filepath"

	"github.com/gofuego/fuego/core"
)

// SiblingHook injects the root-level files that live one level above a .claude
// content directory — CLAUDE.md, CLAUDE.local.md, and .mcp.json — as pages, so a
// project repo's top-level memory and MCP config appear in the site even though
// they sit outside the content dir. siblingDir is that parent directory.
//
// Injected pages are flagged sibling=true: the classifier leaves them alone (they
// classify and route themselves here) and the home hook prefers a sibling
// CLAUDE.md as the site's home. The content dir stays the real .claude, so the
// dev server's live-reload still covers everything inside it.
func SiblingHook(siblingDir string) core.AfterParseHook {
	md := NewParser()
	mcp := &MCPParser{}

	memories := []struct{ name, slug, title string }{
		{"CLAUDE.md", "root-CLAUDE", "CLAUDE.md (project root)"},
		{"CLAUDE.local.md", "root-CLAUDE.local", "CLAUDE.local.md (project root)"},
	}

	return func(pages []*core.Page) ([]*core.Page, error) {
		for _, m := range memories {
			path := filepath.Join(siblingDir, m.name)
			raw, err := os.ReadFile(path)
			if err != nil {
				continue // absent sibling
			}
			env, nodes, perr := md.Parse(raw)
			if perr != nil {
				continue
			}
			if env == nil {
				env = core.Envelope{}
			}
			env["sibling"] = true
			env["source"] = "project"
			env["slug"] = m.slug
			env["title"] = m.title
			env["memory_path"] = "../" + m.name
			pages = append(pages, &core.Page{
				SourcePath: path,
				RelPath:    "../" + m.name,
				Ext:        "md",
				Envelope:   env,
				Nodes:      nodes,
				Type:       "memory",
				Layout:     "memory",
			})
		}

		if path := filepath.Join(siblingDir, ".mcp.json"); fileExists(path) {
			raw, err := os.ReadFile(path)
			if err == nil {
				env, nodes, _ := mcp.Parse(raw)
				env["sibling"] = true
				pages = append(pages, &core.Page{
					SourcePath: path,
					RelPath:    "../.mcp.json",
					Ext:        "json",
					Envelope:   env,
					Nodes:      nodes,
					Type:       "mcp",
					Layout:     "mcp",
				})
			}
		}

		return pages, nil
	}
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
