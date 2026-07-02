---
title: Publish Your Workspace
layout: doc
summary: Build with a base URL, deploy to GitHub Pages, and review what the site exposes before making it public.
nav_section: "Guides"
nav_weight: 4
---

A built site is plain static files — host it anywhere. The two things to get right are the base URL and knowing what you're publishing.

## First: review what the site exposes

The site renders what your files contain. Before publishing anywhere public, check:

- **`.mcp.json`** — server definitions often carry environment variables and internal URLs.
- **`settings.local.json` and `CLAUDE.local.md`** — machine-local by convention, and often not meant to leave your machine.
- **Memory files** — `CLAUDE.md` content is rendered in full, including anything sensitive you wrote down.

There is no redaction pass. If a file shouldn't be public, don't publish the site publicly — or build from a tree that doesn't contain it.

## Build for a subpath

A GitHub Pages *project* site serves under `/<repo>/`, so links need a base URL:

```bash
fuego-dotclaude build --base-url /my-repo -o build
```

For a root domain, omit `--base-url`. To make the setting sticky, put it in [`fuego-dotclaude.yaml`](docs/reference/configuration/) instead of the flag.

## Deploy to GitHub Pages

A minimal workflow that builds on every push to `main` and publishes to a `gh-pages` branch:

```yaml
name: Publish workspace site
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - run: go run github.com/gofuego/fuego-dotclaude@latest build --base-url /${{ github.event.repository.name }}
      - uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: build
          publish_branch: gh-pages
```

Then point the repository's Pages settings at the `gh-pages` branch. On a **private** repository, Pages needs a paid GitHub plan — an alternative is any static host, or an authenticating host in front of the built output.

## Team workflow

Pair the deploy with [`validate --strict` on pull requests](docs/guides/validate-in-ci/): PRs gate on a coherent `.claude` tree, and the merged result publishes automatically. Since [scope detection](docs/concepts/scope-detection/) folds the root `CLAUDE.md` in, the published site is the same thing a teammate sees running `fuego-dotclaude serve` locally.
