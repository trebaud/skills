# Skills

A collection of agent skills I use for work.

## Installation

```bash
mkdir -p ~/.agents
git clone https://github.com/trebaud/skills.git ~/.agents
```

Then create symbolic links to make skills available to your agents:

```bash
ln -s ~/.agents/skills ~/.claude/skills
ln -s ~/.agents/skills ~/.opencode/skills
```

## Skills in the SDLC

```
   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────────┐
   │   Plan    │──▶│  Design   │──▶│  Develop  │──▶│   Test    │──▶│ Review & Ship │
   └───────────┘   └───────────┘   └───────────┘   └───────────┘   └───────────────┘
         ▲                                                                  │
         └──────────────────────────── iterate ─────────────────────────────┘

   Plan:           prd · specs-generator · rfc-generator · interview
   Design:         arch-diagram · kiss-check
   Develop:        janitor · refactor
   Test:           test-generator · debug
   Review & Ship:  code-reviewer · security-analysis · semgrep-audit
                   gha-audit · audit-fix · create-pr · releaser
```

- **Plan** -- `prd` for product requirements, `specs-generator` for feature specs, `rfc-generator` for proposals, `interview` to validate plans
- **Design** -- `arch-diagram` to visualize the system, `kiss-check` to challenge complexity
- **Develop** -- `janitor` for architecture and code quality, `refactor` for safe simplification
- **Test** -- `test-generator` for coverage, `debug` for test-first bug fixing
- **Review & Ship** -- `code-reviewer` + `security-analysis` + `semgrep-audit` for code review, `gha-audit` for CI security, `audit-fix` for dependency vulnerabilities, `create-pr` to open the PR, `releaser` to cut a release

| Skill | Description |
|-------|-------------|
| [arch-diagram](skills/arch-diagram/SKILL.md) | Generates Mermaid.js architecture diagrams. Keeps visual docs in sync with code. |
| [audit-fix](skills/audit-fix/SKILL.md) | Audits prod dependencies and fixes vulnerabilities by tracing to root parents, using recursive updates, parent bumps, or replacements. |
| [code-reviewer](skills/code-reviewer/SKILL.md) | Comprehensive code review covering quality, security, and maintainability. |
| [create-pr](skills/create-pr/SKILL.md) | Creates pull requests with auto-generated title and description. |
| [debug](skills/debug/SKILL.md) | Test-first debugging. Creates reproducing tests, then uses subagents to implement fixes. |
| [gha-audit](skills/gha-audit/SKILL.md) | Audits GitHub Actions workflows for security issues — dangerous triggers, script injection, unpinned actions, over-permissioned tokens. |
| [interview](skills/interview/SKILL.md) | Asks non-obvious technical questions about implementation plans, tradeoffs, and constraints. |
| [janitor](skills/janitor/SKILL.md) | Refactors code toward Clean Architecture and Clean Code principles. Fixes dependency violations, separates concerns, improves code quality. |
| [kiss-check](skills/kiss-check/SKILL.md) | Forces justification for complex solutions. Must explain why simpler won't work. |
| [prd](skills/prd/SKILL.md) | Generates Product Requirements Documents with clarifying questions, user stories, and acceptance criteria. |
| [refactor](skills/refactor/SKILL.md) | Safe refactoring with automated test verification after each step. |
| [releaser](skills/releaser/SKILL.md) | Cuts a new release of a bun package: bumps version, runs typecheck and tests, pushes the git tag, creates a GitHub release with curated changelog. |
| [rfc-generator](skills/rfc-generator/SKILL.md) | Creates RFC documents for new features through interactive questioning. |
| [security-analysis](skills/security-analysis/SKILL.md) | Identifies security vulnerabilities and analyzes security reports. |
| [semgrep-audit](skills/semgrep-audit/SKILL.md) | White-box security auditor with semgrep integration. Triages scan results and proposes concrete fixes. |
| [specs-generator](skills/specs-generator/SKILL.md) | Creates SPECS.md files for new features and design documentation. |
| [test-generator](skills/test-generator/SKILL.md) | Generates unit and integration tests following existing patterns. |

## Misc

Skills that live outside the SDLC loop — situational tools for security monitoring, PR triage, and writing.

| Skill | Description |
|-------|-------------|
| [gh-feed](skills/gh-feed/SKILL.md) | Shows a rich feed of active open pull requests across all GitHub repos — links, review status, CI, comments, and latest activity. |
| [macos-security-audit](skills/macos-security-audit/SKILL.md) | Runs a comprehensive security audit on macOS: processes, network, persistence, hardening, and more. |
| [unslop](skills/unslop/SKILL.md) | Rewrites text to remove AI tropes and cliches, making it sound more natural and human. |
| [vuln-feed-scan](skills/vuln-feed-scan/SKILL.md) | Scans security feeds for recent high-signal vulnerabilities affecting our stack (nodejs, typescript, mongo, npm, aws, crypto). |

## Scripts

| Script | Description |
|--------|-------------|
| `scripts/sync-skills.sh` | Symlinks each skill into `~/.claude/skills` and `~/.opencode/skills`, removing stale links. |
| `scripts/claude-worktree.sh` | Creates a git worktree under `.claude/worktrees/<branch>`, installs deps, and launches Claude Code. |
