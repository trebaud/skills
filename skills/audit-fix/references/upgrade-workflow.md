# Upgrade Workflow Reference

Execution steps for subagents. Each strategy section is self-contained — read only the one assigned to you.

## Strategy A: Recursive Update

```bash
pnpm update <vulnerable-package> --recursive
pnpm why <vulnerable-package>                    # confirm version >= fix version
pnpm audit --audit-level critical --prod         # confirm CVE resolved
```

`--recursive` is required — without it pnpm only updates direct deps, not transitive ones.

## Strategy B: Minor/Patch Parent Bump

```bash
pnpm view <parent> version                       # confirm latest is minor/patch bump
```

Edit package.json to update the parent version, then:

```bash
pnpm install
pnpm why <vulnerable-package>                    # confirm transitive dep updated
pnpm typecheck
```

## Strategy C: Major Parent Bump

### 1. Research (mandatory)

```bash
pnpm view <parent> repository.url
```

Web search: `"<parent> breaking changes v<new-major>"` and `"<parent> migration guide"`.

Check: CHANGELOG.md, GitHub releases, migration guides, issues tagged "breaking-change".

### 2. Implement

Find all usages with Grep (`from '<parent>'` and `require('<parent>')`), then:

```bash
# Edit package.json with new major version
pnpm install
# Fix breaking changes per research
pnpm typecheck
pnpm test <relevant-test-files>
```

**Fallback:** If too complex, try latest minor/patch of current major first. If that doesn't fix it, document and recommend the major bump as a separate task.

## Strategy D: Replace Dependency

1. Find all usages with Grep (`from '<parent>'` and `require('<parent>')`)
2. Check if an alternative is already in package.json
3. Map each used function to its equivalent in the replacement
4. Update imports and function calls
5. Remove deprecated package from package.json
6. `pnpm install && pnpm typecheck && pnpm test <relevant-tests>`

## Strategy E: Accept Risk (Exploitability Analysis)

**All 4 conditions must be met** for a vulnerability to be exploitable:

1. **Attacker-controlled input** — can an external attacker control data flowing into the vulnerable code path?
2. **Dangerous sink reached** — does the codebase actually invoke the vulnerable function? Common false positives: types-only deps (`@types/*`), unused functions, client-only usage
3. **Defenses bypassed** — no input validation, sandboxing, or network restrictions in the way?
4. **Realistic impact** — would exploitation cause meaningful damage?

Report each as:

```markdown
**<CVE-ID>: <vulnerability name>**
- Package: <package>@<version> via <root-parent>
- ① Attacker input: YES/NO — <explanation>
- ② Dangerous sink: YES/NO — <explanation>
- ③ Defenses bypassed: YES/NO — <explanation>
- ④ Realistic impact: YES/NO — <explanation>
- **Verdict: EXPLOITABLE / NOT EXPLOITABLE**
```

## Subagent Completion Checklist

Before reporting back:

1. `pnpm audit --audit-level <level> --prod` — confirm CVE no longer appears
2. `pnpm typecheck` — no type errors introduced
3. `pnpm test <relevant-tests>` — existing tests pass
4. Report: what was done, which strategy succeeded, any remaining issues
