# Report templates

After the helper script writes results back to the memory file, the skill produces two report artifacts next to it:

- `<domain>-typosquat-report.md` — diffable, commits cleanly into a brand-protection repo
- `<domain>-typosquat-report.html` — single-file standalone artifact for non-engineering stakeholders

Both reports are derived entirely from the memory JSON. Read the memory file after the helper exits, then write each report using the template below. Substitute the `{{...}}` placeholders. Sort and group exactly as described — keep terminal output, markdown, and HTML aligned so reviewers can correlate them.

## Sorting and grouping rules

- **Resolving candidates** are grouped by `/24` of their A record. Clusters with more than one name get a header line; singletons render inline. Order clusters by size (largest first), then alphabetically by `/24` key. Within a cluster, sort candidates alphabetically.
- **Status transitions** include rows whose `prev_status` differs from the new `status`, plus first-time observations that resolved. Sort alphabetically by candidate. Show `prev_status` (or `new` if empty) → `status`.
- **By-technique table** counts every row in the memory file by `technique`, broken down by status. Sort by `resolving` desc, then `total` desc, then technique name.
- **`[MX]` marker** appears next to any resolving candidate whose `mx` field is non-empty.
- **HTTP signal** — each resolving row carries `https` and `http` probe objects (status code + final URL). Render the HTTP state inline so reviewers can see at a glance whether the squat serves content, redirects, or refuses connections.

## HTTP state classification

Resolve each candidate to one of these HTTP states by examining its `https` and `http` probe objects (HTTPS dominates when both succeed, since that's the realistic phishing channel):

| State | Definition | Triage weight |
|---|---|---|
| `LIVE` | `https.status` in 200–299 (or HTTP if no HTTPS) AND `final_url` host matches the candidate. The squat serves content from its own name. | Highest — pull a screenshot before it ages. |
| `REDIRECT_OFFSITE` | 2xx response but `final_url` host differs from the candidate AND from the target brand. Squat is funneling traffic somewhere. | High — check destination; often a redirect-based attack or affiliate skim. |
| `REDIRECT_TO_TARGET` | 2xx response and `final_url` host is (or is a subdomain of) the target brand itself. Defensive registration or someone signaling intent. | Medium — verify ownership via WHOIS. |
| `BLOCKED` | 4xx (403, 451, etc.) on both schemes. Server up, content gated — common for active infrastructure not yet pointed at a landing page. | Medium — recheck on next run. |
| `SERVER_ERROR` | 5xx on both schemes. | Low — could be misconfigured squat or transient. |
| `NO_HTTP` | Both probes error with connection refused / timeout / TLS failure. DNS-only registration. | Low for HTTP; **High if `mx` is non-empty** (mail-interception setup). |
| `OTHER` | Mix not covered above (e.g. HTTP 200 but HTTPS handshake fail, exotic codes). | Inspect manually. |

## Priority scoring (drives the "Recommended next steps" block)

HTTP state is the primary signal; cluster membership is the tiebreaker. Score every resolving row, highest first:

1. **Fresh registration** — row appears in this run's status transitions (`prev_status` was `unregistered` or empty, now `resolves`). Tag with its HTTP state (`LIVE`, `REDIRECT_OFFSITE`, etc.). Strongest squatter signal in the dataset regardless of HTTP state.
2. **LIVE singleton** — `LIVE` state, not part of any multi-name `/24` cluster. The squat serves its own content. `[MX]` raises urgency further (full phishing kit).
3. **REDIRECT_OFFSITE** — funnels visitors somewhere unrelated. Document the destination domain in the reason line.
4. **NO_HTTP + MX** — no HTTP server, but mail-capable. Mail-interception or BEC setup; HTTP-only triage will miss it.
5. **BLOCKED singleton** — 4xx on both schemes, solo A record. Infrastructure exists but isn't serving content; could be staging.
6. **REDIRECT_TO_TARGET** — points at the brand. Usually defensive but worth confirming via WHOIS.
7. **Cluster member + LIVE / +MX** — sits in a `/24` cluster (often parking) but has live content or mail capability. Lower urgency, spot-check.
8. **Everything else** — typical registrar parking or CDN catch-all. Note in aggregate; don't enumerate.

Quote at most 5 candidates by name in the priority block. Everything else rolls up into the cluster counts further down.

## Quick-test link set

Every resolving candidate gets the same five links so a reviewer can triage in one click. Use the candidate FQDN verbatim — these are user-derived strings, so URL-encode in HTML (markdown autolinks tolerate dots and hyphens directly).

| Label | URL pattern | What it answers |
|---|---|---|
| `visit` | `https://{candidate}` | Is there a live site? Parking page or real content? |
| `crt.sh` | `https://crt.sh/?q={candidate}` | When was the first TLS cert issued? Any subdomains? |
| `VT` | `https://www.virustotal.com/gui/domain/{candidate}/detection` | Any AV vendor flagging it? Passive DNS history. |
| `urlscan` | `https://urlscan.io/domain/{candidate}` | Recent screenshots, redirects, requested resources. |
| `whois` | `https://who.is/whois/{candidate}` | Registrar, creation date, registrant (often privacy-masked). |

Render them in markdown as ` · `-separated bracketed links after the IP/technique line. Don't add links to unregistered or error rows.

## Markdown template

```markdown
# Typosquat scan: {{TARGET}}

_Generated {{LAST_RUN}} — run #{{RUN_COUNT}}_

## Summary

| Metric | Count |
|---|---:|
| Candidates tracked | {{STATS_TOTAL}} |
| Resolving (registered, DNS active) | {{STATS_RESOLVING}} |
| Unregistered (NXDOMAIN — available) | {{STATS_UNREGISTERED}} |
| Lookup errors (retry next run) | {{STATS_ERRORS}} |

## Recommended next steps

<!-- One short paragraph headline ("N high-priority items this run."), then a
bulleted list of the top items by the priority rubric above. For each, lead
with the candidate FQDN, give the HTTP state tag, the one-line reason, then
the quick-test links. Cap at 5 candidates. Example shapes:

  - **Fresh registration · LIVE:** `<candidate>` → `<ip>` _(<technique>)_ [MX]
    https=200 (self-served), new this run — pull WHOIS + a screenshot before it ages.
    [visit](https://<candidate>) · [crt.sh](https://crt.sh/?q=<candidate>) · [VT](https://www.virustotal.com/gui/domain/<candidate>/detection) · [urlscan](https://urlscan.io/domain/<candidate>) · [whois](https://who.is/whois/<candidate>)

  - **LIVE singleton + MX:** `<candidate>` → `<ip>` _(<technique>)_ [MX]
    https=200 serving own content, mail-capable — likely active phishing kit.
    [visit](...) · [crt.sh](...) · [VT](...) · [urlscan](...) · [whois](...)

  - **REDIRECT_OFFSITE:** `<candidate>` → `<ip>` _(<technique>)_
    https=302 → `<final_url_host>` — funneling traffic; review destination.
    [visit](...) · [crt.sh](...) · [VT](...) · [urlscan](...) · [whois](...)

  - **NO_HTTP + MX:** `<candidate>` → `<ip>` _(<technique>)_ [MX]
    both schemes refused, MX present — likely mail-interception setup, not web.
    [visit](...) · [crt.sh](...) · [VT](...) · [urlscan](...) · [whois](...)

Close the section with a one-line summary of the rest:
  _N other resolving candidates sit in parking-provider clusters (see below)._

If no resolving candidates: write _No resolving candidates this run — nothing to action._
-->

## Resolving candidates ({{STATS_RESOLVING}})

Grouped by `/24` so parking-provider clusters collapse into a single block. `[MX]` means the name can receive email. The `https=…` / `http=…` line carries the HTTP probe state (status code + final-URL host if redirected, or the connection error). Each row also carries quick-test links — click through to triage.

<!-- For each cluster:
  - if singleton:
    - `<candidate>` → `<ip>` _(<technique>)_  [MX]
      `https=200 (self)` · `http=301→<final_host>`
      [visit](https://<candidate>) · [crt.sh](https://crt.sh/?q=<candidate>) · [VT](https://www.virustotal.com/gui/domain/<candidate>/detection) · [urlscan](https://urlscan.io/domain/<candidate>) · [whois](https://who.is/whois/<candidate>)
  - if multi:
    - **Cluster <ip>/24** (<n> names on shared infrastructure):
      - `<candidate>` → `<ip>` _(<technique>)_  [MX]
        `https=…` · `http=…`
        [visit](...) · [crt.sh](...) · [VT](...) · [urlscan](...) · [whois](...)
      - ...
  If no resolving rows, write: _None._
  Format conventions for the http line:
    - successful response: `https=<code>` if final_url host matches candidate, `https=<code>→<final_host>` otherwise.
    - missing response: `https=<error>` (the short `error` field from the probe, e.g. `connection refused`, `timeout`).
    - omit the per-scheme entry entirely if the probe object is missing (older memory files).
-->

## Status transitions this run ({{N_TRANSITIONS}})

<!-- Omit this whole section if there are no transitions. Otherwise one bullet per transition:
  - `<candidate>`: <prev_status or "new"> → **<status>**  [MX]
    `https=…` · `http=…` (only when status == resolves)
    [visit](https://<candidate>) · [crt.sh](https://crt.sh/?q=<candidate>) · [VT](https://www.virustotal.com/gui/domain/<candidate>/detection) · [urlscan](https://urlscan.io/domain/<candidate>) · [whois](https://who.is/whois/<candidate>)
  (Only attach the http line and links when the new status is `resolves`.)
-->

## By technique

| Technique | Total | Resolving | Unregistered | Errors |
|---|---:|---:|---:|---:|
<!-- one row per technique, sorted resolving desc / total desc / name asc -->

## Notes

- `resolves` ≠ active squatter. Most resolving lookalikes are registrar parking pages or CDN catch-alls. Triage out-of-band (HTTP fetch, TLS cert SAN, WHOIS, reverse DNS, CT-log history).
- `[MX]` on a resolving candidate is a stronger phishing/mail-interception signal than HTTP-only.
- The system resolver was used. Behind split-horizon DNS, results may differ from an external view.
```

## HTML template

Use the file [`report_template.html`](./report_template.html) verbatim — it contains the full standalone document with embedded CSS, light/dark via `color-scheme`, and the same sections as the markdown report. Substitute the same `{{...}}` placeholders. The HTML-specific placeholders (`{{PRIORITY_BLOCK}}`, `{{RESOLVING_BLOCK}}`, `{{TRANSITIONS_BLOCK}}`, `{{TECHNIQUE_ROWS}}`) are HTML fragments to splice in; the comment inside each one shows the exact tag shape to emit per row/cluster, including the inline `<a>` quick-test link strip.

Always HTML-escape user-derived strings (candidate, IP, technique, status values) when emitting HTML, and URL-encode the candidate when building link `href`s — none of these are trusted, even though they came from your own DNS lookups.
