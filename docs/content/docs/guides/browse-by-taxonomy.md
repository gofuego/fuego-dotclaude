---
title: Browse by Taxonomy
layout: doc
summary: Three faceted hubs — tools, models, and sources — group artifacts by what they use and where they came from.
nav_section: "Guides"
nav_weight: 1
---

Beyond the file-tree sidebar, the site groups artifacts along three axes, each with a hub page and a page per value. The hubs are linked from the topbar of every page.

## `/tools/` — by tool

Groups agents and commands by the tools they declare — an agent's `tools:` frontmatter and a command's `allowed-tools`. Answers "which of my agents can run Bash?" and "what would be affected if I revoked WebFetch?"

## `/models/` — by model

Groups artifacts by their `model:` frontmatter (`opus`, `sonnet`, `haiku`, …). Useful for spotting which agents are pinned to which model tier before a model migration.

## `/sources/` — by provenance

Every artifact is stamped with where it came from: your project itself, or the name of the plugin that shipped it. The source hub is how you see at a glance what an installed plugin actually added to your workspace.

## The hubs always exist

The hub pages (`/tools/`, `/models/`, `/sources/`) are generated even when empty — a workspace where no agent declares tools still gets a `/tools/` page (with no entries) rather than a dead topbar link. Value pages (`/tools/bash/`, `/models/opus/`) exist only for values that actually appear.

## Where the values come from

The taxonomy fields are normalized from each artifact's frontmatter during the build — you don't annotate anything. If an agent declares no `tools` or `model`, it simply doesn't appear in those hubs (agents that inherit all tools by omitting `tools:` are not listed under every tool).
