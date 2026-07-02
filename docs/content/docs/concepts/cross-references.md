---
title: Cross-References
layout: doc
summary: Whole-word mentions of artifact names become links; backlinks, disambiguation pages, and the global reference graph make the web navigable in both directions.
nav_section: "Concepts"
nav_weight: 2
---

The signature feature: your artifacts already reference each other by name — an agent's prompt mentions a skill, a skill's workflow tells Claude to invoke `docs-writer`, a settings file allowlists an MCP server. fuego-dotclaude turns those mentions into a navigable web.

## Forward links

Every *named* artifact is a link target: frontmatter `name`s (agents, skills, output styles), command paths, and MCP server keys. On every rendered page, whole-word occurrences of those names become links to the artifact's page.

The linker is deliberately conservative about **where** it links:

| Links | Never links |
|---|---|
| Prose | Fenced code blocks (`<pre>`) |
| Inline `` `code` `` spans | Text that is already a link |
| | The page's own name (no self-links) |

Matching is whole-word and case-insensitive, and names that collide with common words are stopword-guarded. The linker rewrites the rendered HTML by walking its token stream — not regex over markup — so it cannot corrupt your content.

## Ambiguous names

If two artifacts share a name (an agent and a skill both called `review`), mentions link to a generated **disambiguation page** listing every artifact with that name, each with its kind, so the reader picks.

## Backlinks: "Referenced by"

Every page that is mentioned elsewhere gets a **Referenced by** section listing the pages that mention it — the reverse direction of every auto-link. This is how you answer "what still uses this agent?" before deleting it.

## The global graph

`/references/` renders the whole reference graph as an edge list: every source page and everything it mentions. Together with per-page backlinks, this gives you both the zoomed-in and zoomed-out view of how your workspace hangs together.

## Validation catches dangling references

Auto-linking only links names that exist. For *explicit* references — an agent naming a `subagent_type`, an artifact referencing an MCP server — the [validate command](docs/guides/validate-in-ci/) reports mentions of names that don't resolve to anything.
