---
name: audit-fix
description: Audit prod dependencies and fix vulnerabilities by tracing to root parents, using recursive updates, parent bumps, or replacements
tools: Bash, Read, Grep, Glob, WebSearch, WebFetch, Agent
metadata:
  version: 1.0.0
---

# Audit Fix

Run `pnpm audit`, trace each vulnerability to its root parent, select a fix strategy, and apply fixes in parallel subagents.

## Parameters

- **`<package-name>`** — only fix vulnerabilities for this package
- **`--level <severity>`** — fix all at this level and above (`low`, `moderate`, `high`, `critical`)
- **No args** — defaults to `--level critical`

## Workflow

### Step 0: Branch Check

Before making any changes, check the current branch:

```bash
git branch --show-current
```

If on `master` or `main`, create a new branch:

```bash
git checkout -b audit-fix-<short-descriptor>
```

Use a descriptor that reflects the scope (e.g. `audit-fix-critical`, `audit-fix-pbkdf2`).

### Step 1: Audit and Parse

```bash
pnpm audit --audit-level <level> --prod
```

For each vulnerability, extract: package, fix version, and full dependency path. If a specific package was passed, filter to that package only. Stop if no vulnerabilities found.

### Step 2: Group by Root Parent

The root parent is the first dep in the path (the direct dependency in package.json). Multiple CVEs from the same root parent are handled together.

### Step 3: Select Strategy Per Root Parent

For each root parent, check these in order — use the first that applies:

| Strategy | Condition | Check with |
|----------|-----------|------------|
| **A: Recursive update** | Parent's semver range allows the fix version | `pnpm view <parent>@<version> dependencies` |
| **B: Minor/patch bump** | Newer minor/patch of parent exists | `pnpm view <parent> version` |
| **C: Major bump** | Only a major bump fixes it | `pnpm view <parent> version` — **web search for breaking changes** |
| **D: Replace dep** | Parent is deprecated/abandoned | `pnpm view <parent>` deprecation notice, check for alternatives in package.json |
| **E: Accept risk** | No fix available | Exploitability analysis (see [workflow reference](references/upgrade-workflow.md#strategy-e)) |

### Step 4: Launch Subagents

One Agent per root parent, all in parallel. Each subagent must create its own commit scoped to that root parent so changes stay reviewable per-package:

```
Agent tool:
  subagent_type: "general-purpose"
  description: "Fix <root-parent> audit vulns"
  prompt: |
    Fix audit vulnerabilities for <root-parent>.
    Strategy: <A/B/C/D/E>
    Vulnerabilities: <CVE list with fix versions>
    Dependency chain: <full path>
    Read and follow: .claude/skills/audit-fix/references/upgrade-workflow.md

    After the fix is applied and verified, create a dedicated commit for this
    package only. Stage just the files changed by this fix (typically
    package.json and pnpm-lock.yaml) and commit with a message like:

      fix(audit): bump <root-parent> to <version> (<CVE ids>)

    Do NOT stage changes made by other parallel subagents.
```

Because subagents run in parallel on the same working tree, serialize the commit step if needed (e.g. via a lockfile check) or have each subagent stage only its own changed paths. If strategy E (accept risk) was chosen, skip the commit for that package.

### Step 5: Verify and Report

```bash
pnpm audit --audit-level <level> --prod
```

Report as:

```markdown
| Root Parent | CVEs | Strategy | Result |
|-------------|------|----------|--------|
| soap | 2 | A: Recursive update | FIXED |
| request | 2 | E: Accept risk | NOT EXPLOITABLE |
```

Include exploitability analysis for any remaining vulnerabilities.

## Anti-Patterns

- **Don't delete the lockfile** — use `pnpm update <pkg> --recursive` instead
- **Never use overrides**, overrides bypass semver intent
- **Don't do risky major bumps** for non-exploitable CVEs

## Example

```
/audit-fix                        # Fix all critical prod vulnerabilities
/audit-fix --level high           # Fix high and critical
/audit-fix pbkdf2                 # Fix only pbkdf2 vulnerabilities
```
