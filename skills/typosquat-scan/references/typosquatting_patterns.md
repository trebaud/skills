# Typosquatting patterns

Playbook for generating typosquat candidates. The LLM produces the candidate list by applying every applicable technique below to the target label, then writes the merged result into `./<domain>-typosquat-memory.json` (new entries with `{"candidate": "<fqdn>", "technique": "<name>"}` and no `last_checked`); the helper script `scripts/typosquat_scan.go` then resolves every pending and stale row. Each section names a technique, says what it models, and shows what its output looks like for the example label `google` (TLD `com`). The technique names match the `technique` field on each candidate in the JSON memory file.

When generating, apply every technique to the target's label, plus the `tld_swap` set against a curated TLD list (`com, net, org, co, io, app, dev, cc, tv, me, info, biz, us, uk, eu, xyz, online, site, store, shop, club, page, ai`). Skip candidates already present in the memory file. The helper caps the batch at `-max-candidates`, so for very common-letter labels prefer a balanced subset across techniques rather than exhausting one type first.

## omission

Drop one character from the label. This is what happens when a finger slips or a key gets missed. Output: `oogle`, `ggle`, `goole`, `googe`, `googl`. Labels of two characters or fewer are skipped, since the result is too short to be meaningfully confusable.

## multi_omission

Drop two adjacent characters. Only runs for labels of eight characters or more, where the result is still distinctively close to the brand. Output for `cloudflare`: `oudflare`, `cudflare`, `cldflare`, `cloflare`, `cludflare`-style results — every adjacent pair removed in turn. Captures realistic two-finger fumbles on long brand names that the single-character omission misses.

## repetition

Double one character. Captures stuck keys and held-too-long fingers. Output: `ggoogle`, `gooogle`, `googgle`, `googlee`. A noticeable share of real squatters use this for short, recognizable brand names.

## transposition

Swap two adjacent characters, the classic hand-coordination slip. Output: `ogogle`, `gogole`, `goolge`, `googel`. Pairs of identical characters are skipped because the swap would be a no-op.

## replacement (keyboard)

Replace a character with one of its US-QWERTY neighbors, i.e. hitting the wrong key. Output for `g`: `f`, `t`, `y`, `h`, `b`, `v` substituted in place. Tuned for ASCII and US keyboards. AZERTY and Dvorak users mistype differently, but US QWERTY remains the dominant target.

## insertion (keyboard)

Insert a keyboard-neighbor character next to an existing one, the stray-finger case. Two insertion points per neighbor: before and after the source character. Output for `g`: `fgoogle`, `gfoogle`, `tgoogle`, `gtoogle`, and so on.

## homoglyph

Replace a character with a visually similar single character (ASCII only). The attacker is targeting the reader's eye, not the typist's fingers; the URL should look right at a glance. Substitutions used: `o`/`0`, `l`/`1`/`i`, `e`/`3`, `a`/`4`, `s`/`5`, `b`/`8`, `g`/`9`, `z`/`2`, `t`/`7`. Output: `g00gle`, `googie`, `goog1e`, `g0ogle`. Unicode and IDN homographs (Cyrillic `а` for Latin `a`, and so on) are not covered. See "Out of scope" below.

## compound_homoglyph (cognitive blindspot)

Replace a single character with a *multi-character cluster* whose glyphs combine into the same visual shape at a glance — `m` → `rn`, `w` → `vv`. Output: `arnazon.com` for `amazon`, `vvalmart.com` for `walmart`, `tvvitter.com` for `twitter`, `grnail.com` for `gmail`.

This is distinct from single-character homoglyph. There, the substitution exploits *character-level* visual similarity (`o` looks like `0` on its own). Here, the substitution exploits a perceptual blindspot in *word-level* reading: the brain chunks adjacent glyphs into the shape they collectively resemble and glides past without inspecting each letter, a phenomenon related to saccadic reading and gestalt grouping. The effect is robust across the proportional sans-serif fonts used in browsers, email clients, and chat apps — `rn` and `m` are nearly indistinguishable in Helvetica, Arial, and the default system UI fonts on macOS, Windows, and Android.

High-value targets are short, well-known brand labels containing `m` or `w`, where the cluster substitution still reads as a familiar word at a glance: `amazon`, `walmart`, `gmail`, `twitter`, `microsoft`, `meta`.

## hyphenation

Insert a single hyphen between two characters of the label. Mirrors the "company-name.com" pattern common in landing pages and phishing kits. Output: `g-oogle`, `go-ogle`, `goo-gle`, `goog-le`, `googl-e`.

## vowel_swap

Replace a vowel with another vowel. Captures phonetic mistakes and non-native-speaker spelling guesses. Output for `o`: `gaogle`, `gegogle`, `gigogle`, `gugogle` and so on, also applied to the second `o` and to `e`. False-positive rate is high, but generation is cheap and the matches that actually resolve are usually intentional squatters.

## bitsquat

Flip a single bit in the ASCII value of one character. Keep it only if the result is still a valid DNS label character (alphanumeric or hyphen). The threat model is a cosmic ray or hardware fault that flips a bit in a DNS query in flight, in DNS cache, or in a CPU register. Dinaburg's 2011 DEF CON research showed measurable traffic to bitsquat domains of high-volume brands. Defensive value is lower than for typos because the squatter cannot predict which bit will flip, but a handful of canonical bitsquat domains for big brands are still worth registering. Output for `g` (0x67): flipping each bit gives `f`, `e`, `c`, `o`, `w`, `7`, so candidates include `foogle`, `eoogle`, `coogle`, `ooogle`, `woogle`, `7oogle`.

## tld_swap

Keep the label, swap the TLD for a common alternative. The model squatter is one who couldn't get the `.com` and registered `.co`, `.io`, `.app`, `.dev`, `.xyz`, etc. instead. Built-in TLD list: `com, net, org, co, io, app, dev, cc, tv, me, info, biz, us, uk, eu, xyz, online, site, store, shop, club, page, ai`. For well-known brand names this is often the highest-yield technique; many `<brand>.io` and `<brand>.ai` lookalikes already resolve.

## doppelganger

Glue a subdomain-style prefix to the label without the separating dot, producing a valid DNS label that reads at a glance like the real subdomain. Built-in prefix list: `www, mail, login, secure, my, app, account, auth, id`. Output: `wwwgoogle.com`, `mailgoogle.com`, `logingoogle.com`, `securegoogle.com`. The Godai Group "Doppelganger Domains" study estimated that prefixes of this shape on Fortune 500 brands intercepted on the order of 20 GB of misdirected email over six months — making this a high-yield generator even when the registrant is passive.

## combosquat

Glue a phishing-kit keyword to either side of the label, with or without a separating hyphen. Built-in keyword list: `login, secure, pay, support, help, account, verify, app`. Each keyword produces four candidates per target: `<label><kw>`, `<kw><label>`, `<label>-<kw>`, `<kw>-<label>`. Output for `google`: `googlelogin.com`, `logingoogle.com`, `google-login.com`, `login-google.com`, and the analogous forms for the other keywords. Search space is bounded by design (keyword list × 4 per target), so the technique stays cheap while covering the shapes that phishing kits actually use.

## Out of scope

These attack classes exist but are deliberately not generated by the script.

*IDN and Punycode homographs* are visually identical Unicode substitutions, like Cyrillic `а` for Latin `a` or Greek `ο` for Latin `o`. Detecting them well needs a full confusables table from Unicode TR39 plus per-TLD policy (some registries refuse mixed-script labels). A separate generator would belong in its own pass.

*Sound-alike or phonetic squats* (`gewgle.com`, `gugel.com`) need a phoneme model.

*Subdomain confusion* (`google.com.evil.tld`) is a phishing-page concern, not a domain-registration one, and is not visible at the DNS-resolution layer.

*Path-based typosquatting* (`googl.ecom`) produces broken DNS labels that will not resolve.

If you need any of these, extend this playbook with a new section describing the rule, and start emitting the candidates with a fresh technique name — the helper script will accept and persist whatever technique label you choose.

## Resolution status meanings

What the `status` field on each candidate actually tells you:

- `resolves`: the system resolver returned at least one A record. This does *not* mean a squatter is active. The IP might be a registrar parking page, a CDN catch-all, or an unrelated legitimate site that happens to share the name. Always triage further with HTTP fetch, TLS cert, WHOIS, reverse DNS.
- `unregistered`: DNS authoritatively says the domain does not exist (NXDOMAIN). The candidate is currently available, or more often for valuable brands, registered but not delegated.
- `error`: timeout, SERVFAIL, or network issue. Re-run later; transient errors are common when scanning hundreds of names against a local resolver.

A status transition from `unregistered` to `resolves` between scans is the highest-signal event the memory file captures: someone registered and delegated the domain in the interval. `prev_status` and `prev_checked` are populated automatically whenever a re-check flips the status, and the run output prints a "Status transitions this run" block summarizing them. To surface fresh registrations, just run the script again — rows older than 7 days are auto-rechecked.

## The `mx` field

When a candidate resolves, the script also captures its MX records. An MX-bearing lookalike can receive email regardless of whether anyone serves an HTTP page from it, so this field is a distinct, high-value signal:

- An MX on a candidate plus an active A record is consistent with someone running a passive mail-harvester, capturing whatever misaddressed email arrives.
- An MX with a registrar's catch-all hostname (e.g. `mx*.parkingcrew.net`) is usually benign parking; an MX on the squatter's own infrastructure is not.
- Doppelganger and combosquat candidates show up with MX more often than typo candidates, matching the threat model — these are the shapes attackers register specifically for mail interception.

## Triage and monitoring beyond DNS

DNS resolution is a cheap first filter, not a verdict. A `resolves` row tells you a name exists; it does not tell you who is behind it or what they intend. To close the loop on a candidate, layer on:

- *Certificate Transparency logs.* Every public TLS cert issued for a domain is published to CT. Watch CT feeds (crt.sh, Google's Argon, Cloudflare's Nimbus) for newly issued certs whose SAN matches your brand's lookalike set — a fresh Let's Encrypt cert on a candidate that resolved last week is a strong signal of an imminent phishing page.
- *WHOIS and registration analytics.* Bulk registrations of brand variants across many TLDs by the same registrant, registrar, or nameserver cluster point to an organized squatter rather than a coincidence. Damerau-Levenshtein distance against your brand list is a reasonable similarity metric when sifting WHOIS bulk feeds.
- *HTTP / TLS fingerprint.* Fetch the candidate over HTTPS and look at the served cert, the body, and any redirect chain. Parked-domain templates, registrar holding pages, and phishing kits each have recognizable fingerprints.
- *Inbound traffic to your own infra.* If you operate authoritative DNS or mail servers, log query patterns for misspellings. Squatters sometimes set up catch-all MX records to harvest misdirected email — Godai Group's "Doppelganger Domains" study estimated that dotless-subdomain squats alone intercepted on the order of 20 GB of email over six months from Fortune 500 lookalikes. Misrouted internal mail is a defensive priority on its own, separate from phishing of customers.
