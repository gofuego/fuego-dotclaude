---
title: CLI Reference
layout: doc
summary: Every command and flag — build, serve, validate, and list — plus the path resolution and config lookup they share.
nav_section: "Reference"
nav_weight: 1
---

```
fuego-dotclaude <command> [path] [flags]
```

All four commands share the same optional `[path]` argument, resolved by the [scope detection rules](docs/concepts/scope-detection/) (default: the current directory), the same `--siblings`/`--no-siblings` override, and the same [configuration lookup](docs/reference/configuration/).

## build

Build the static site.

```bash
fuego-dotclaude build [path] [flags]
```

| Flag | Meaning |
|---|---|
| `-o, --output <dir>` | Output directory (default `build`, or `output_path` from config). |
| `--base-url <path>` | Base URL for deployment, e.g. `/my-repo`. |
| `--incremental` | Reuse cached parses for unchanged files. |
| `--siblings` / `--no-siblings` | Force or suppress root-level sibling injection. |

## serve

Build and serve with live reload, rebuilding on change.

```bash
fuego-dotclaude serve [path] [flags]
```

| Flag | Meaning |
|---|---|
| `--port <n>` | Dev server port (default: engine default). |
| `--base-url <path>` | Base URL, when previewing a subpath deployment. |
| `--siblings` / `--no-siblings` | As above. |

Output and cache are written to a stable scratch directory under the OS temp dir, keyed by the content path — never into the watched tree (which would loop the watcher), and the incremental cache survives restarts.

## validate

Report advisory coherence diagnostics; see [Validate in CI](docs/guides/validate-in-ci/) for the full list.

```bash
fuego-dotclaude validate [path] [--strict]
```

| Flag | Meaning |
|---|---|
| `--strict` | Exit non-zero if any problem is found (for CI gates). |

## list

Print every discovered artifact as a table.

```bash
fuego-dotclaude list [path]
```

Columns: `TYPE` (agent, skill, command, …), `NAME`, and `ROUTE` (the page URL it will get).

## version

```bash
fuego-dotclaude --version
```

Prints the version baked in at build time (`dev` for a source build).
