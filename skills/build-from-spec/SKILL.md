---
name: build-from-spec
description: >-
  Implement a spec in the current repo: split the work into ordered sections,
  track their status, and build section by section with one commit each. Use
  when the user says "implement this spec", "build the spec at <source>", or
  hands over a spec and asks for granular commits.
argument-hint: <spec source: a doc URL, a file path, a URL, or pasted text>
user-invocable: true
---

# Build from a spec

## 0. Preflight

Check whether you run in Claude Code: the `CLAUDECODE` environment variable is `1`, or the `plugin-authoring` skill is listed. Claude Code unlocks two extras below: Claude doc reading and the progress pane. Without it, skip both.

## 1. Read the spec

Read `$ARGUMENTS` with whatever tool fits the source.

In Claude Code, a claude.ai doc link is read with the docs skill and its tools. When the result overflows to a file, convert it with `python3 -I <skill>/scripts/doc2md.py <saved file> <out.md>`. Elsewhere, ask the user to export the doc as markdown.

## 2. Split into sections

Order the sections by dependency, each small enough for one commit with its tests. The last section is the full checks.

## 3. Set up tracking

**In Claude Code:** open the progress pane.

1. Load the `plugin-authoring` skill.
2. Copy `<skill>/mod/spec-progress/` into the session's mods folder it names, and run `claude plugin validate` on the copy.
3. End the turn so the mod loads.
4. Call `mcp__spec-progress__set_sections` with the section list.

**Elsewhere:** keep the sections as a checklist in your todo tool, or in the conversation.

## 4. Build each section

For each section, in order:

1. Mark it in progress (`update_section` → `in_progress`).
2. Implement it, with its tests, and check it.
3. Commit.
4. Mark it done with the commit sha (`update_section` → `done`, sha as `note`).

## 5. Finish

Report the commits, any deviations from the spec, and what was not run.
