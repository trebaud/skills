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

## Resolving candidates ({{STATS_RESOLVING}})

Grouped by `/24` so parking-provider clusters collapse into a single block. `[MX]` means the name can receive email.

<!-- For each cluster:
  - if singleton:
    - `<candidate>` → `<ip>` _(<technique>)_  [MX]   ← omit [MX] if no MX
  - if multi:
    - **Cluster <ip>/24** (<n> names on shared infrastructure):
      - `<candidate>` → `<ip>` _(<technique>)_  [MX]
      - ...
  If no resolving rows, write: _None._
-->

## Status transitions this run ({{N_TRANSITIONS}})

<!-- Omit this whole section if there are no transitions. Otherwise one bullet per transition:
  - `<candidate>`: <prev_status or "new"> → **<status>**  [MX]
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

Use the file [`report_template.html`](./report_template.html) verbatim — it contains the full standalone document with embedded CSS, light/dark via `color-scheme`, and the same sections as the markdown report. Substitute the same `{{...}}` placeholders. The HTML-specific placeholders (`{{RESOLVING_BLOCK}}`, `{{TRANSITIONS_BLOCK}}`, `{{TECHNIQUE_ROWS}}`) are HTML fragments to splice in; the comment inside each one shows the exact tag shape to emit per row/cluster.

Always HTML-escape user-derived strings (candidate, IP, technique, status values) when emitting HTML — none of these are trusted, even though they came from your own DNS lookups.
