---
name: vuln-feed-scan
description: Scan a whitelist of security feeds for recent high-signal vulnerabilities that pose immediate threat to our stack (nodejs, typescript, mongo, npm, aws, crypto). Use when the user asks to check security news, scan vuln feeds, look for recent CVEs/zero-days, or run an OSINT sweep for security incidents. Keywords include "check vulns", "security feed scan", "any new CVEs", "zero-day check", "scan for threats".
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

### 5. Report

The report is the only thing the user reads. Format it for scanning, with plenty of whitespace between findings.

**5a. Ranked table.** Most urgent first. One row per finding:

```
| # | Severity | CVE / Name | Affects | Published | Action |
```

- `#` is a 1-based index matching the detail blocks below, so the user can jump between them.
- Severity: Critical / High only. Anything lower does not belong in the report.
- Affects: the specific stack item it hits (e.g. "npm: express@<4.20.0", "AWS IAM", "MongoDB driver").
- Action: concrete next step in ≤12 words (e.g. "Pin express ≥4.20.0 in backend/package.json", "Rotate AWS access keys issued before YYYY-MM-DD", "Audit `npm ls <pkg>` across repos").

**5b. Detail blocks.** After the table, one block per finding in the same order. Separate blocks with a horizontal rule (`---`) so each finding is visually distinct. Use this structure, keeping each bullet on its own line:

```
### {index}. {CVE / Name} — {Severity}

- **Affects:** {specific stack item(s)}
- **Published:** {date, or `date?`}
- **Source:** [{primary source name}]({url}) · [{secondary, if any}]({url})

**What it is.** One sentence describing the vulnerability in plain terms.

**Why it matters to us.** One sentence tying it to our stack — which service, which dep, which blast radius.

**Attacker capability.** One sentence on what exploitation gets them (RCE, data exfil, token theft, etc.).

**Action.** The same concrete next step from the table, expanded to one sentence if useful.
```

Rules for the detail blocks:

- Never merge the four labeled paragraphs into one run-on line. Each gets its own line with a blank line before and after.
- No padding sentences, no "it is important to note that", no best-practice lectures. If a section would be empty or obvious, drop it.
- Every CVE/incident name in both the table and the detail heading is a markdown hyperlink to its primary source.

**5c. Suggested next commands.** A single fenced `bash` block listing exact commands the user can run (e.g. `npm audit`, `npm ls <pkg>`, `grep -r <indicator>`). One command per line, with a `# {finding index}: {short intent}` comment above each so the user knows which finding it maps to.

**5d. Also worth a glance.** A final section titled `## Also worth a glance` listing items that did not clear the score ≥ 7 bar but could still be interesting: tangential to our stack, partially corroborated, or a slow-burn supply-chain story that may matter later. One line per item, up to 8 items total:

```
- [{CVE / Name}]({url}) — {≤15 words on what it is and why it's borderline} _(score {n}, {reason it was down-ranked})_
```

Pull candidates from items dropped in step 3c / 4 for score reasons only. Do not include items hard-dropped by the 3b stack pre-filter (Windows-only, PHP/WordPress, ICS, etc.); those are noise. If nothing qualifies, omit the section entirely; do not emit an empty header.

**5e. Empty report.** If nothing survives step 4 and nothing qualifies for 5d, say exactly: `No high-signal threats in window. Next scan recommended in 12h.` with no table or detail blocks. If the main report is empty but 5d has items, keep 5d and prepend a single line: `No items cleared the high-signal bar, but these may be worth a glance:`.

### 6. Save report files

After outputting the report to the user, write both a markdown and an HTML file using the `Write` tool. Use `date -u +%Y-%m-%d_%H-%M-%S` to obtain the timestamp at write time. Write both files in parallel.

**6a. Markdown file** — `./YYYY-MM-DD_HH-MM-SS_vuln-scan.md`

- Content: the report as shown to the user (ranked table, source links, suggested commands)
- Every CVE / incident name in the table must be a markdown hyperlink to its primary source. Do not omit links.
- Add a header line: `# Vuln Scan — YYYY-MM-DD HH:MM:SS UTC` (same timestamp, spaces/colons for readability) and a `**Window:** Xd | **Stack:** Node/TS/MongoDB/npm/AWS/crypto` metadata line before the table. Example header: `# Vuln Scan — 2026-04-20 14:07:33 UTC`.

**6b. HTML file** — `~./YYYY-MM-DD_HH-MM-SS_vuln-scan.html`

Write a single self-contained HTML file with:
- Inline CSS only (no external dependencies)
- Dark background (`#0d1117`), GitHub-like aesthetic
- Clean sans-serif font (system-ui stack)
- A sticky header with scan title, timestamp, and window/stack metadata
- The ranked summary table rendered as an HTML `<table>` with hover rows and severity badges
- Severity badge colors:
  - Critical: red (`#da3633` bg, `#ffc1be` text)
  - High: orange (`#9e4a01` bg, `#ffa657` text)
- Detail blocks as styled `<article>` elements, each separated visually, containing:
  - `<h3>` heading with CVE/name linking to primary source
  - Labeled paragraphs (**What it is**, **Why it matters to us**, **Attacker capability**, **Action**) each in its own `<p>`
  - Source links as `<a>` tags
- Suggested commands in a styled `<pre><code>` block with a dark inset background
- "Also worth a glance" as a final `<section>` with a muted header and list items, if any findings qualify; omit entirely if none
- Smooth hover transitions on table rows and article cards
- Responsive layout (max-width 960px, centered)
- The HTML file must be fully self-contained (no external CSS/JS/fonts)

After writing the HTML file, open it:
```bash
open ./YYYY-MM-DD_HH-MM-SS_vuln-scan.html
```

Confirm to the user with a single line listing both paths: `Reports saved to ./<filename>.md and <filename>.html`.

## Rules

- Always run step 4. A single primary source (KEV, GH Advisory, NVD, vendor advisory) is sufficient signal on its own. A single secondary source, or an HN story with an unfamiliar underlying link, must be corroborated by step 4 before it qualifies.
- Never invent CVE ids or dates. If a date isn't in the fetched content, mark it `date?` and down-rank.
- Do not include general security best-practice advice. Report findings only, not hygiene.
- Do not summarize items the user almost certainly already knows (e.g. year-old log4j). Respect the cutoff window.
- If a fetch fails, note it in a single line at the end (`Fetch failed: <url>`) and continue with the rest. One broken source should not block the whole run.
