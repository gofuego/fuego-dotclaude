---
title: Configuration
layout: doc
summary: The optional fuego-dotclaude.yaml — its three fields, where it's looked up, and the defaults that apply without it.
nav_section: "Reference"
nav_weight: 2
---

fuego-dotclaude is zero-config by design — the routes, theme, and parsers all come from the pack. One small optional file, `fuego-dotclaude.yaml`, overrides the site-level settings:

```yaml
site_name: "Acme Engineering Workspace"   # default: "Claude Code Workspace"
base_url: "/acme-claude"                  # default: "" (root)
output_path: "dist"                       # default: "build"
```

All fields are optional; anything unset keeps its default. CLI flags (`--base-url`, `-o`) win over the file for a single run.

## Lookup order

The first file found wins:

1. `<content-dir>/fuego-dotclaude.yaml` — inside the `.claude` directory itself
2. one directory above the content dir — the project root, next to `.claude/`
3. the current working directory

The project-root location (2) is the natural home: the config travels with the repo without cluttering the `.claude` directory Claude Code reads.

## What is deliberately *not* configurable

Routes, taxonomies, layouts, and the theme are fixed by the pack so every rendered workspace looks and navigates the same. If you need to change those, consume the pack from your own Fuego project instead — see [Use as a Pack](docs/reference/use-as-a-pack/), where your config and theme files override the pack's.
