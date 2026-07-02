---
title: "See your Claude Code workspace"
layout: home
cta:
  - label: "Get started"
    url: "/docs/getting-started/"
  - label: "CLI reference"
    url: "/docs/reference/cli/"
    ghost: true
---

A `.claude/` directory accumulates agents, skills, slash commands, output styles, memory, MCP servers, settings, and plugins — and no way to see them as a whole. **fuego-dotclaude renders it as a navigable, cross-referenced site**: every artifact becomes a page, every whole-word mention of an artifact's name becomes a link, and every page lists what references it.

```bash
go install github.com/gofuego/fuego-dotclaude@latest
cd my-project && fuego-dotclaude serve
```

One command, zero config. The tool detects whether you pointed it at a project (builds `./.claude` and folds in the root `CLAUDE.md`/`.mcp.json`) or at a `.claude` directory itself. Your top-level `CLAUDE.md` becomes the home page, followed by a generated catalog of everything else — including the agents, skills, and commands inside each installed plugin, rendered as first-class pages.
