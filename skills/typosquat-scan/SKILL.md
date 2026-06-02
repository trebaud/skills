---
name: typosquat-scan
description: Generate typosquat candidates for a domain (omission, transposition, single + compound homoglyphs, keyboard typos, bitsquatting, hyphenation, vowel swaps, TLD swaps, doppelganger prefixes, bounded combosquats) and check which ones resolve via DNS, capturing A and MX records, then probe HTTP/HTTPS on every resolver to capture response code and final URL. Candidates are produced by applying the patterns reference; a Go helper resolves and probes them in parallel and persists results to a JSON memory file so successive runs surface status transitions. Use when the user wants to find lookalike domains, defensive registrations, phishing impersonators, or brand-protection candidates for a domain. Trigger on phrasings like "find typosquats for X", "domain squatters", "lookalike domains", "check phishing variants", "brand protection scan".
---

# typosquat-scan

Generates lookalike domain candidates for a target by applying the patterns in [`references/typosquatting_patterns.md`](./references/typosquatting_patterns.md), writes them into a per-target JSON memory file, then runs a small Go helper that resolves any pending or stale row (A + MX) and writes results back. Every row that resolves is additionally probed over `https://` and `http://` (independently, following up to 5 redirects each) so the HTTP status code + final URL land in the memory file alongside DNS state — this is what drives the triage priority in the report. Rows last checked >7 days ago are auto-rechecked on every run, so the file doubles as a brand-protection feed — the highest-signal event it captures is an `unregistered → resolves` transition, which surfaces fresh squatter registrations. After the helper exits, the skill renders two report files (`<domain>-typosquat-report.md` and `<domain>-typosquat-report.html`) from the memory file using the templates in [`references/`](./references/).

The split is deliberate. The Go helper has one job: resolve DNS + probe HTTP in parallel and persist results. Everything else — candidate generation, report rendering, triage — lives in the skill layer (i.e., here, driven by the LLM):
- **Candidate generation** is a linguistic transformation: apply the patterns doc to the target label.
- **DNS resolution + HTTP probe** is parallel I/O: the helper does this.
- **Report rendering** is formatting: read the memory file and write markdown + HTML using the templates.

## When to use

User asks any of:
- "find typosquats for `<domain>`"
- "what lookalike domains exist for `<domain>`"
- "scan for phishing domains targeting `<domain>`"
- "defensive registration candidates for `<domain>`"
- "check `<domain>` typosquats" / "brand-protection scan"

## Workflow

Run from the directory where the memory file should live (typically a brand-protection or research project directory). Memory file path: `./<domain>-typosquat-memory.json` (dots in the domain become `_`).

1. **Read [`references/typosquatting_patterns.md`](./references/typosquatting_patterns.md).** Each section defines one technique and what its output looks like for a sample label.
2. **Read the memory file** if it exists, so you know which candidates are already tracked.
3. **Generate candidates.** Apply every applicable technique from the patterns doc to the target's label. Skip ones already in memory.
4. **Write the memory file.** Emit a JSON file with the merged candidate set — every existing row kept verbatim, every new candidate added as `{"candidate": "<fqdn>", "technique": "<name>"}` (no `status`, no `last_checked`). Empty `last_checked` is the marker that tells the helper this row is pending DNS resolution.
5. **Run the helper.** It picks up every row whose `last_checked` is empty (new) or older than 7 days (auto-recheck), caps the batch at `-max-candidates`, resolves A + MX in parallel, and for every row that resolves additionally probes `https://` and `http://` (independent requests, up to 5 redirects each, ~6s timeout, TLS verification disabled so self-signed certs still surface a status code). Writes everything back to the same file. Terminal output: resolving candidates clustered by `/24` with HTTP status inline, plus any status transitions.
6. **Render the reports.** Read the updated memory file, then follow [`references/report_template.md`](./references/report_template.md) (which references [`references/report_template.html`](./references/report_template.html)) to write `./<domain>-typosquat-report.md` and `./<domain>-typosquat-report.html` next to the memory file. Same domain-to-underscore rule as the memory file path. Each report opens with a **Recommended next steps** block — up to 5 highest-priority resolving candidates ranked by the HTTP-aware rubric in the template, each with a one-line reason and a strip of quick-test links (visit, crt.sh, VT, urlscan, whois) so reviewers can click straight into triage. The markdown is for diffable history; the HTML is a standalone artifact for non-engineering stakeholders.
7. **Triage the resolving set.** The HTTP probe already differentiates "live site" from "DNS-only registration" from "redirect to brand"; for anything you flag in the priority block, follow up out-of-band: WHOIS, TLS cert SAN, reverse DNS, CT-log history. See "Triage and monitoring beyond DNS" in the patterns doc.

If the memory file doesn't exist, just create it: write the JSON skeleton with `target`, an empty `stats` object, and your generated `candidates` array. The helper fills in the rest on first run.

## How to run

Single self-contained Go file, stdlib only. The only flag is `-max-candidates` (default 200, `0` = no cap):

```bash
go run ~/.claude/skills/typosquat-scan/scripts/typosquat_scan.go amazon.com
go run ~/.claude/skills/typosquat-scan/scripts/typosquat_scan.go -max-candidates 500 amazon.com
```

For repeated use, build once:

```bash
go build -o ~/.claude/skills/typosquat-scan/scripts/typosquat_scan \
  ~/.claude/skills/typosquat-scan/scripts/typosquat_scan.go
```

## Fallback: no Go installed

If `go` isn't available, install it first (`brew install go` on macOS, `apt install golang` on Debian/Ubuntu). The helper is stdlib-only and builds in under a second.

If installing Go is genuinely off the table, fall back to bash + `dig` (preinstalled on macOS; on Linux ensure `dnsutils`/`bind-utils` is present) to resolve the candidate list directly out of the memory JSON. Persistence updates and stale-recheck logic won't run in that path — you'd have to write status back into the memory file by hand.

## JSON memory

Default location: `./<domain>-typosquat-memory.json` in the **current working directory** (dots in the domain are replaced with `_`, so `example.com` → `example_com-typosquat-memory.json`).

Schema:

```json
{
  "target": "example.com",
  "first_seen": "2026-05-14T21:54:01Z",
  "last_run": "2026-05-14T21:54:10Z",
  "run_count": 2,
  "stats": { "total": 8, "resolving": 5, "unregistered": 3, "errors": 0 },
  "candidates": [
    {
      "candidate": "exarnple.com",
      "technique": "compound_homoglyph",
      "status": "resolves",
      "ip": "149.120.153.85",
      "mx": "mx1.example-mail.net",
      "https": { "status": 200, "final_url": "https://exarnple.com/login" },
      "http": { "status": 301, "final_url": "https://exarnple.com/login" },
      "http_checked": "2026-05-14T21:54:01Z",
      "first_seen": "2026-05-14T21:54:01Z",
      "last_checked": "2026-05-14T21:54:01Z",
      "prev_status": "unregistered",
      "prev_checked": "2026-05-07T08:00:00Z"
    }
  ]
}
```

One entry per candidate. When a re-check flips status, `status` becomes the new value while the previous state is preserved in `prev_status` / `prev_checked`. The `https` and `http` objects are refreshed wholesale on each check (so a domain that goes NXDOMAIN drops its stale probe data) and only written for rows that resolve. The file is rewritten in full on every run via temp-file rename, so partial-write corruption isn't a risk.

**HTTP probe fields** (per scheme, `https` and `http`):

| Field | Meaning |
|---|---|
| `status` | HTTP status code from the final response (after up to 5 redirects). `0` / omitted means the probe didn't get a response — see `error`. |
| `final_url` | URL after redirects. If different host from the candidate, the squat is redirecting somewhere — often a brand site (defensive registration) or unrelated infrastructure (potential redirect-based abuse). |
| `error` | Short reason the probe failed: `timeout`, `connection refused`, TLS handshake failures, etc. Empty on success. |

A row with `https.status = 200` and `final_url` pointing back at itself is a live site — highest triage priority. `https.status = 200` with a `final_url` on the target brand's own domain is usually a defensive registration. Both probes erroring (`connection refused`/`timeout`) plus a non-empty `mx` means DNS+MX-only registration: HTTP isn't the attack surface, mail interception is.

**Status values:**

| Value | Meaning |
|---|---|
| `resolves` | A record exists — the domain is registered and DNS is pointing somewhere (parking page, real site, CDN). |
| `unregistered` | Authoritative DNS returned NXDOMAIN — the name does not exist. In practice this almost always means the domain is available to register; a `unregistered → resolves` transition is the strongest fresh-squatter signal. |
| `error` | Lookup failed for some other reason (SERVFAIL, timeout, network issue). Status is unknown; the row is retried on the next run. |


## Operational caveats

The script uses the system resolver. Behind a corporate split-horizon DNS, results may differ from an external view. For an attacker's-eye view, re-run on a machine that only sees public resolvers.

`resolves` does not mean "active squatter." Most resolving lookalikes are registrar parking pages or CDN catch-alls. Clustering output by `/24` helps surface these — twenty candidates on the same parking provider IP collapse into one cluster line.

`[MX]` on a resolving candidate means the name can receive email. Combined with a non-registrar A record, this is a stronger phishing/mail-interception signal than HTTP-only.

ASCII labels only. IDN and Punycode homograph attacks are out of scope; see "Out of scope" in [`references/typosquatting_patterns.md`](./references/typosquatting_patterns.md).
