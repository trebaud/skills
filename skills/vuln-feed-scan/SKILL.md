---
name: vuln-feed-scan
description: Scan a whitelist of security feeds for recent high-signal vulnerabilities that pose immediate threat to our stack (nodejs, typescript, mongo, npm, aws, crypto), then cross-check each surviving finding against repositories in the current working directory and report concrete lockfile/package hits. Use when the user asks to check security news, scan vuln feeds, look for recent CVEs/zero-days, or run an OSINT sweep for security incidents. Keywords include "check vulns", "security feed scan", "any new CVEs", "zero-day check", "scan for threats".
---

# Vuln Feed Scan

Scan the whitelist sources below and surface only vulnerabilities that directly threaten our stack. Keep it terse.

## Our stack (scoring context)

Flag any finding that touches: **Node.js**, **TypeScript**, **MongoDB**, **npm** (registry or any popular package), **AWS** (any service), **crypto** (wallets, custody, exchanges, Bitcoin/Lightning, web3 libs).

Also flag infra that typically sits in front of or alongside the above: TLS libs, HTTP parsers, container runtimes, CI/CD (GitHub Actions), widely-used Node frameworks (Express, Fastify, Next.js, NestJS), ORMs (Mongoose, Prisma), auth libs (jsonwebtoken, passport).

## Whitelist

Check these first. Do not read sources not on this list unless the user expands the list for this run. Sources are tiered; the tier affects scoring in step 3.

**Primary / authoritative:** original advisories and canonical vuln data; signal stands on its own.

- https://github.com/advisories?query=type%3Areviewed+ecosystem%3Anpm (npm advisories)
- https://socket.dev/blog
- https://www.aikido.dev/category/vulnerabilities-threats
- https://snyk.io/blog/
- https://www.stepsecurity.io/blog
- https://www.wiz.io/blog

**Secondary / journalism:** original reporting, often derivative of a vendor advisory but can break stories.

- https://www.bleepingcomputer.com/
- https://thehackernews.com/

**Community-curated:** link aggregation; front-page position is a strong community validation signal.

- https://news.ycombinator.com/ — filter to security/CVE/exploit stories and record the item's approximate page position (top 10 / rest of page 1 / later). High placement = real signal even without other corroboration.

If the user provides additional URLs in their prompt, add them to this run only; do not persist.

## Workflow

Run these in order.

### 1. Plan the run

Emit a one-line plan stating: scan start timestamp (UTC, format `YYYY-MM-DD HH:MM:SS UTC`), cutoff window (default: last 24h), stack filter, and whitelist size. Let the user interrupt if they want to change scope (e.g. widen to 3d after a quiet weekend).

### 2. Fetch whitelist in parallel

Use `WebFetch` on every whitelist URL in a single message with parallel tool calls. Do not fetch serially. Ask each fetch to extract: item title, publish date, affected product/version, CVE id if present, one-line summary.

For aggregator pages (HN, Bleeping, Krebs), the prompt should request only items matching security/vuln/CVE/exploit/breach keywords, to avoid pulling general tech news.

### 3. Build candidate list

**3a. Age gate.** Drop anything older than the cutoff window × 1.5 (default window 3 days → hard drop at 5 days). Exception: items labeled actively exploited, CISA KEV, or supply-chain may extend to 10 days. Anything older than 10 days is always dropped. Deduplicate across sources (same CVE or same incident).

**3b. Stack pre-filter (hard drop, not scoring).** An item must satisfy at least one of the following before it is scored. If none match, drop it — do not carry it into scoring on "interesting" grounds:

- Directly affects a stack item: Node.js runtime, TypeScript compiler, MongoDB (server or official drivers), npm registry itself, a popular npm package (≥100k weekly downloads OR known transitive dep of Express/Fastify/Next.js/NestJS/Mongoose/Prisma), any AWS service, a crypto/web3 library or wallet/custody/exchange system we could plausibly touch.
- Affects adjacent infra we actually run: TLS libs (OpenSSL, BoringSSL, rustls), HTTP parsers used by Node, container runtimes (Docker/containerd), GitHub Actions or the hosting provider (Vercel, Netlify, Cloudflare) for a Node/TS app, widely-used Node framework or auth lib listed above.
- Supply-chain incident affecting any developer tool we plausibly use: npm/GitHub/CI provider breach, maintainer account takeover of a popular package, malicious version published to a registry, leaked CI/provider tokens.

Items that affect only Windows desktops, Java/JVM-only stacks (ActiveMQ, Tomcat, Log4j-adjacent), PHP/WordPress, Microsoft Office, ICS/OT, or niche npm packages (<100k weekly downloads and no popular parent) are **out of scope** — drop them even if they look spicy.

**3c. Score survivors.**

**Severity / impact:**

| Signal | Weight |
|---|---|
| Published ≤ 24h ago | +4 |
| Published ≤ 48h ago | +2 |
| Published 2–3d ago | −2 |
| Published 4–7d ago | −4 (only via the KEV/exploited/supply-chain age exception) |
| Published 8–10d ago | −6 (KEV/exploited/supply-chain only; must be genuinely active) |
| Labeled zero-day / actively exploited / in CISA KEV | +5 |
| Directly affects a stack item | +4 |
| Affects adjacent infra we run | +2 |
| Pre-auth RCE | +3 |
| Credential theft / secret exposure affecting our stack | +3 |
| Requires unusual local access or social engineering only | −3 |
| No CVE, no vendor advisory, no PoC | −2 |

**Supply-chain signals** (apply on top of the severity table; one compromised dep poisons every repo that pulls it, so weight these heavily):

| Signal | Weight |
|---|---|
| Popular npm package compromised: malicious version published, or maintainer account takeover | +6 |
| npm registry itself, GitHub Actions, or a major CI/CD provider breached | +5 |
| Hosting/deploy provider we could use breached with dev tokens leaked (Vercel, Netlify, Cloudflare, AWS-adjacent) | +5 |
| Typosquat / dependency-confusion campaign targeting npm with confirmed downloads | +3 |
| Build-tool or lockfile-manipulation vulnerability (npm, pnpm, yarn, bun) | +3 |

**Source signal** (apply by source tier, not raw count; multiple mentions of the same underlying article do not stack):

| Signal | Weight |
|---|---|
| Reported by ≥1 primary source | +3 |
| Reported by ≥2 primary sources | +4 (instead of +3) |
| Reported by a secondary (journalism) source | +2 |
| Same incident in both primary and secondary | +1 (on top of the above) |
| Appears on HN front page (top ~10) | +2 |
| Appears on HN page 1 below top 10, or later pages | +1 |

Keep items with score ≥ 7 before step 4. Cap at top 8 candidates. If nothing scores ≥ 7, say so and stop. Do not pad. Supply-chain items that clear the bar must be ranked above same-score non-supply-chain items in the final report.

### 4. Corroborate with a generic search

For each surviving candidate, run one `WebSearch` using the CVE id or incident name plus a confirmation term (`"exploited" OR "PoC" OR "patch"`). Do these searches in parallel.

Adjust score based on what the search returns. Do not drop items for lack of corroboration; a freshly-disclosed vuln may legitimately have only one source at first.

| Search result | Weight |
|---|---|
| Vendor advisory, KEV entry, or CVE record found independently | +3 |
| Second outlet or public PoC repo found | +2 |
| Nothing found | −2 |

Re-apply the score ≥ 7 cutoff after this adjustment. A single primary-source item (e.g. a fresh KEV entry) still clears the bar on its own; a single secondary-source story with no corroboration usually will not.

### 5. Cross-check against local repos

For every finding that survived step 4, check whether the affected package/component appears in any repository sitting in the current working directory. This is the step that turns a feed scan into an actionable patch list.

**5a. Enumerate repos and attack surfaces.** Run `ls -la` (or `find . -maxdepth 2 -name package.json -not -path "*/node_modules/*"`) to identify candidate repo directories. Treat each immediate subdirectory as a separate repo. Skip `node_modules`, build output dirs, and anything obviously not a project root. Record the list — it goes in the report header.

For **every** repo in the list, enumerate the attack surfaces below before per-finding lookup. Do not skip a surface because the previous repo didn't have it; check each repo independently. Missing files are fine, but the *check* must run per repo per surface.

- **Package manifests and lockfiles** — `package.json`, `pnpm-lock.yaml` (**always check this — pnpm is the primary package manager**), `pnpm-workspace.yaml`, `package-lock.json`, `yarn.lock`, `npm-shrinkwrap.json`, `bun.lockb`, and any nested workspace manifests (e.g. `packages/*/package.json`, `apps/*/package.json`). These are the primary attack surface; never skip them.
- **GitHub Actions** — every file under `.github/workflows/` (`*.yml`, `*.yaml`) plus `.github/actions/*/action.yml` for composite actions. Treat this surface as mandatory for every repo, even if the finding looks unrelated — a finding can hit via a reusable action or a third-party `uses:` reference you wouldn't otherwise expect.
- **Other manifests** — `Dockerfile`, `docker-compose*.yml`, IaC (`*.tf`, `cdk.*`, `serverless.yml`, `cloudformation*.{yml,json}`), and runtime configs (`.nvmrc`, `.node-version`, `tsconfig.json`) when the finding plausibly touches them.

**5b. Per-finding lookup.** For each surviving candidate, run the cheapest applicable check across every repo in parallel. **Every finding must be cross-checked against every repo** — do not stop at the first hit, and do not assume one repo's result applies to the others. When in doubt, run the grep; the cost is negligible compared to missing a hit.

- **npm package compromise / CVE in a JS/TS library** — for **every** repo, grep all package manifests (`package.json` at root and in any workspace path) for the package name, then grep **every** lockfile present (`pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`, `npm-shrinkwrap.json`, `bun.lockb`) to capture the actually-resolved version. Both direct and transitive matter — a clean `package.json` with a poisoned lockfile entry is still a hit.
- **npm registry / supply-chain worm with a known publish-window** — also check each lockfile's `mtime` against the malicious-publish window (`ls -la <repo>/pnpm-lock.yaml` etc., for every lockfile in every repo). A lockfile written inside or seconds after the window is a critical signal even when versions look clean.
- **GitHub Actions / CI vulnerability** — for **every** repo, grep all files under `.github/workflows/` (`*.yml` and `*.yaml`) and any `.github/actions/*/action.yml` composite actions for the affected action, `uses:` reference, trigger, or pattern. Always run this check even when the finding's primary vector looks unrelated — a vulnerable third-party action can land via a workflow that pulls it transitively.
- **AWS service issue** — grep IaC files (`*.tf`, `cdk.*`, `serverless.yml`, `cloudformation*.{yml,json}`) and SDK imports (`@aws-sdk/`, `aws-sdk`) for the affected service.
- **Container / runtime CVE** — grep `Dockerfile`, `docker-compose*.yml`, base image tags.
- **Crypto / web3 library** — grep package names plus any direct imports.

For each finding, capture: the repo(s) that hit, the file(s) and line(s), the actually-resolved version where applicable, and whether it's direct or transitive. Note the lockfile `mtime` only when the finding has a time-bounded compromise window. Keep this evidence — it goes into the detail block.

**5c. Run read-only checks and capture verbatim output.** Use cheap, read-only tools: `grep -rn`, `ls -la`, `cat`, `find`. Run them via `Bash` and **keep the exact output** — trimmed to the relevant lines, but verbatim (file paths, line numbers, resolved versions). This output is embedded into the report in step 6b as a fenced `Local evidence` block, so the user can see for themselves what we found. Do not paraphrase. Do not run anything that mutates lockfiles, installs packages, hits the network, or modifies CI state — those are *remediation commands* the user runs themselves (step 6c). If a repo's lockfile is in an unfamiliar format, note it (`{repo}: lockfile format unrecognized`) and move on.

**5d. No-hit findings still ship.** A finding can clear the score bar without matching any local repo — it stays in the report, but its detail block's `In our repos` line says `No direct or transitive match found in <N> repos checked.` so the user sees the check was actually run.

**5e. Hard-drop findings the local check rules out.** If a finding's only path to relevance is "we might use X" and grep proves we don't (e.g. Jenkins plugin compromise + no Jenkins anywhere), move the finding from the main report to *Also worth a glance* with the reason `no usage in local repos`. Don't silently delete it.

### 6. Report

Output the findings to the user. The report is the only thing they read — format it for scanning, most-urgent first, with whitespace between findings. The exact templates (ranked table columns, detail-block structure, remediation-commands block, "Also worth a glance" section, empty-report sentinel) live in **`references/report-format.md`** under "On-screen report". Read that file before rendering the report.

### 7. Save report files

After outputting the report, write a markdown file and a self-contained HTML file in parallel with `Write`. Use `date -u +%Y-%m-%d_%H-%M-%S` to obtain the timestamp at write time.

- Markdown path: `./YYYY-MM-DD_HH-MM-SS_vuln-scan.md`
- HTML path: `./YYYY-MM-DD_HH-MM-SS_vuln-scan.html`

Content and styling specs for both files are in **`references/report-format.md`** under "Saved markdown file" and "Saved HTML file" — header format, metadata line, severity badge colors, layout, etc.

After writing the HTML file, open it:

```bash
open ./YYYY-MM-DD_HH-MM-SS_vuln-scan.html
```

Confirm with a single line: `Reports saved to ./<filename>.md and <filename>.html`.

## Rules

- Always run step 4. A single primary source (KEV, GH Advisory, NVD, vendor advisory) is sufficient signal on its own. A single secondary source, or an HN story with an unfamiliar underlying link, must be corroborated by step 4 before it qualifies.
- Never invent CVE ids or dates. If a date isn't in the fetched content, mark it `date?` and down-rank.
- Do not include general security best-practice advice. Report findings only, not hygiene.
- Do not summarize items the user almost certainly already knows (e.g. year-old log4j). Respect the cutoff window.
- If a fetch fails, note it in a single line at the end (`Fetch failed: <url>`) and continue with the rest. One broken source should not block the whole run.
