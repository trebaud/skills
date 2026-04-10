# My Agent Config

A collection of agent skills I use for work.

## Installation

```bash
git clone https://github.com/moose/agents-config.git ~/.agents
```

Then create symbolic links to make skills available to your agents:

```bash
ln -s ~/.agents/skills ~/.claude/skills
ln -s ~/.agents/skills ~/.opencode/skills
```

## Skills

| Skill | Description |
|-------|-------------|
| [arch-diagram](skills/arch-diagram/SKILL.md) | Generates Mermaid.js architecture diagrams. Keeps visual docs in sync with code. |
| [code-reviewer](skills/code-reviewer/SKILL.md) | Comprehensive code review covering quality, security, and maintainability. |
| [create-pr](skills/create-pr/SKILL.md) | Creates pull requests with auto-generated title and description. |
| [debug](skills/debug/SKILL.md) | Test-first debugging. Creates reproducing tests, then uses subagents to implement fixes. |
| [extract-skill](skills/extract-skill/SKILL.md) | Extracts knowledge from web pages or files to create reusable skills. |
| [interview](skills/interview/SKILL.md) | Asks non-obvious technical questions about implementation plans, tradeoffs, and constraints. |
| [janitor](skills/janitor/SKILL.md) | Refactors code toward Clean Architecture and Clean Code principles. Fixes dependency violations, separates concerns, improves code quality. |
| [kiss-check](skills/kiss-check/SKILL.md) | Forces justification for complex solutions. Must explain why simpler won't work. |
| [macos-security-audit](skills/macos-security-audit/SKILL.md) | Runs a comprehensive security audit on macOS: processes, network, persistence, hardening, and more. |
| [prd](skills/prd/SKILL.md) | Generates Product Requirements Documents with clarifying questions, user stories, and acceptance criteria. |
| [refactor](skills/refactor/SKILL.md) | Safe refactoring with automated test verification after each step. |
| [rfc-generator](skills/rfc-generator/SKILL.md) | Creates RFC documents for new features through interactive questioning. |
| [security-analysis](skills/security-analysis/SKILL.md) | Identifies security vulnerabilities and analyzes security reports. |
| [semgrep-audit](skills/semgrep-audit/SKILL.md) | White-box security auditor with semgrep integration. Triages scan results and proposes concrete fixes. |
| [specs-generator](skills/specs-generator/SKILL.md) | Creates SPECS.md files for new features and design documentation. |
| [test-generator](skills/test-generator/SKILL.md) | Generates unit and integration tests following existing patterns. |
| [unslop](skills/unslop/SKILL.md) | Rewrites text to remove AI tropes and cliches, making it sound more natural and human. |

## Scripts

| Script | Description |
|--------|-------------|
| `scripts/sync-skills.sh` | Symlinks each skill into `~/.claude/skills` and `~/.opencode/skills`, removing stale links. |
| `scripts/claude-worktree.sh` | Creates a git worktree under `.claude/worktrees/<branch>`, installs deps, and launches Claude Code. |

## Skills in the SDLC

```mermaid
flowchart LR
    Plan["Plan<br/>prd · specs-generator<br/>rfc-generator · interview"]
    Design["Design<br/>arch-diagram · kiss-check"]
    Dev["Develop<br/>janitor · refactor"]
    Test["Test<br/>test-generator · debug"]
    Review["Review & Ship<br/>code-reviewer · security-analysis<br/>semgrep-audit · create-pr"]

    Plan --> Design --> Dev --> Test --> Review
    Review -.->|iterate| Plan
```

- **Plan** -- `prd` for product requirements, `specs-generator` for feature specs, `rfc-generator` for proposals, `interview` to validate plans
- **Design** -- `arch-diagram` to visualize the system, `kiss-check` to challenge complexity
- **Develop** -- `janitor` for architecture and code quality, `refactor` for safe simplification
- **Test** -- `test-generator` for coverage, `debug` for test-first bug fixing
- **Review & Ship** -- `code-reviewer` + `security-analysis` + `semgrep-audit` before merging, `create-pr` to ship
