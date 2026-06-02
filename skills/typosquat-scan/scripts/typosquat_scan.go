// typosquat_scan.go — resolve pending and stale candidates from a JSON memory
// file via DNS (A + MX) in parallel and write results back to the same file.
// For every row that resolves we additionally probe both https:// and http://
// (separately, following up to 5 redirects each) so the report layer can use
// HTTP status code + final URL as a triage signal alongside DNS state.
//
// The candidate set itself is produced by the LLM that drives the skill and
// is written into the memory file (new rows with an empty `last_checked`);
// this helper just picks up everything that needs (re-)checking and resolves
// it. Selection rule: any row with empty `last_checked` is brand-new and gets
// checked; any row whose `last_checked` is older than 7d is auto-rechecked,
// so the file doubles as a brand-protection feed and an
// unregistered → resolves transition shows up automatically on subsequent runs.
//
// Scope: this binary only handles DNS resolution, HTTP probing, and memory-file
// persistence. Report rendering (markdown + HTML) is the skill's responsibility —
// the LLM driving the workflow reads the memory file after this script exits and
// writes the report files itself (see SKILL.md and references/report_template.*).
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	staleAfter     = 7 * 24 * time.Hour
	defaultMaxCand = 200
	workers        = 50
	dnsTimeout     = 2 * time.Second
	httpTimeout    = 6 * time.Second
	maxRedirects   = 5
	probeUserAgent = "Mozilla/5.0 (compatible; typosquat-scan/1.0; +brand-protection)"
)

type httpProbe struct {
	Status   int    `json:"status,omitempty"`    // 0 = no response (see Error)
	FinalURL string `json:"final_url,omitempty"` // URL after following up to maxRedirects
	Error    string `json:"error,omitempty"`     // short connect/TLS/timeout reason; empty on success
}

type candidateRecord struct {
	Candidate   string     `json:"candidate"`
	Technique   string     `json:"technique"`
	Status      string     `json:"status"`
	IP          string     `json:"ip,omitempty"`
	MX          string     `json:"mx,omitempty"`
	HTTPS       *httpProbe `json:"https,omitempty"`
	HTTP        *httpProbe `json:"http,omitempty"`
	HTTPChecked string     `json:"http_checked,omitempty"`
	FirstSeen   string     `json:"first_seen,omitempty"`
	LastChecked string     `json:"last_checked,omitempty"`
	PrevStatus  string     `json:"prev_status,omitempty"`
	PrevChecked string     `json:"prev_checked,omitempty"`
}

// Status values stored on each candidate row:
//   resolves     — A record exists (domain is registered and pointing somewhere)
//   unregistered — authoritative DNS says the name does not exist (NXDOMAIN);
//                  in practice this almost always means the domain is
//                  available to register, though technically a registered
//                  domain can return NXDOMAIN if misconfigured
//   error        — DNS lookup failed for some other reason (SERVFAIL, timeout,
//                  network issue); status is unknown and will be retried
type stats struct {
	Total        int `json:"total"`
	Resolving    int `json:"resolving"`
	Unregistered int `json:"unregistered"`
	Errors       int `json:"errors"`
}

type memory struct {
	Target     string             `json:"target"`
	FirstSeen  string             `json:"first_seen"`
	LastRun    string             `json:"last_run"`
	RunCount   int                `json:"run_count"`
	Stats      stats              `json:"stats"`
	Candidates []*candidateRecord `json:"candidates"`
}

type lookupResult struct {
	candidate      string
	status, ip, mx string
	https, http    *httpProbe // nil unless status == "resolves"
}

func resolveOne(ctx context.Context, r *net.Resolver, fqdn string) (status, ip, mx string) {
	ips, err := r.LookupHost(ctx, fqdn)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return "unregistered", "", ""
		}
		return "error", "", ""
	}
	if len(ips) == 0 {
		return "unregistered", "", ""
	}
	mxs, _ := r.LookupMX(ctx, fqdn)
	names := make([]string, 0, len(mxs))
	for _, m := range mxs {
		names = append(names, strings.TrimSuffix(m.Host, "."))
	}
	return "resolves", ips[0], strings.Join(names, ";")
}

// shortError trims a Go network error down to something useful for triage.
// Long stack-trace-style addresses ("dial tcp 1.2.3.4:443:") are stripped so
// the memory file stays diffable.
func shortError(err error) string {
	if err == nil {
		return ""
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "timeout"
	}
	msg := err.Error()
	// Drop everything before the last ": " — Go wraps network errors with
	// host:port prefixes that change between runs and bloat diffs.
	if i := strings.LastIndex(msg, ": "); i != -1 && i < len(msg)-2 {
		msg = msg[i+2:]
	}
	// Cap length so a stray verbose error can't blow up the memory file.
	if len(msg) > 120 {
		msg = msg[:120]
	}
	return msg
}

// httpProbeOne issues a single HEAD-then-GET probe against the given URL,
// follows up to maxRedirects, and returns the final status code + URL. TLS
// verification is intentionally disabled — we want a status code from squat
// infrastructure even when it has a self-signed or expired cert. We do NOT
// trust any response body, only the status line and final URL.
func httpProbeOne(client *http.Client, url string) *httpProbe {
	makeReq := func(method string) (*http.Response, error) {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", probeUserAgent)
		req.Header.Set("Accept", "*/*")
		return client.Do(req)
	}
	// HEAD first — cheaper and won't pull HTML payloads from parking pages.
	// Many sites reject HEAD with 405 / serve different status; on anything
	// that smells off, retry with GET. We discard the body either way.
	resp, err := makeReq("HEAD")
	if err == nil && (resp.StatusCode == 405 || resp.StatusCode == 501) {
		resp.Body.Close()
		resp, err = makeReq("GET")
	}
	if err != nil {
		return &httpProbe{Error: shortError(err)}
	}
	defer resp.Body.Close()
	final := resp.Request.URL.String()
	return &httpProbe{Status: resp.StatusCode, FinalURL: final}
}

// newProbeClient returns an http.Client with a redirect cap, lax TLS, and a
// per-request timeout. One client per worker is fine — http.Client is safe
// for concurrent use, but pooling per worker keeps connection reuse local.
func newProbeClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		DialContext:           (&net.Dialer{Timeout: httpTimeout}).DialContext,
		TLSHandshakeTimeout:   httpTimeout,
		ResponseHeaderTimeout: httpTimeout,
		DisableKeepAlives:     true, // each probe hits a different host
	}
	return &http.Client{
		Transport: tr,
		Timeout:   httpTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func memoryPath(domain string) string {
	safe := strings.ReplaceAll(domain, ".", "_")
	name := safe + "-typosquat-memory.json"
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, name)
	}
	return name
}

// loadMemory reads the JSON memory file. On load it dedupes candidates by
// name (keeping the row with a non-empty last_checked when there's a clash),
// so the LLM can append-merge new candidates without worrying about exact
// duplicate handling.
func loadMemory(path, domain string) (*memory, map[string]*candidateRecord, bool, error) {
	m := &memory{Target: domain, Candidates: []*candidateRecord{}}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, map[string]*candidateRecord{}, false, nil
		}
		return nil, nil, false, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(m); err != nil {
		return nil, nil, false, fmt.Errorf("parse %s: %w", path, err)
	}
	migrated := false
	index := map[string]*candidateRecord{}
	for _, r := range m.Candidates {
		r.Candidate = strings.ToLower(strings.TrimSpace(r.Candidate))
		if r.Candidate == "" {
			continue
		}
		// Migrate legacy status values from earlier versions of the skill.
		if r.Status == "nxdomain" {
			r.Status = "unregistered"
			migrated = true
		}
		if r.PrevStatus == "nxdomain" {
			r.PrevStatus = "unregistered"
			migrated = true
		}
		prev, seen := index[r.Candidate]
		if !seen {
			index[r.Candidate] = r
			continue
		}
		// Dedup: prefer the row that has been checked, or merge technique.
		if prev.LastChecked == "" && r.LastChecked != "" {
			index[r.Candidate] = r
		}
		if index[r.Candidate].Technique == "" && r.Technique != "" {
			index[r.Candidate].Technique = r.Technique
		}
	}
	deduped := make([]*candidateRecord, 0, len(index))
	for _, r := range index {
		deduped = append(deduped, r)
	}
	sort.Slice(deduped, func(i, j int) bool { return deduped[i].Candidate < deduped[j].Candidate })
	m.Candidates = deduped
	return m, index, migrated, nil
}

// writeMemory rewrites the file via temp-file rename. Safe against crashes
// mid-write; readers see either the old or the new full file.
func writeMemory(path string, m *memory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func recomputeStats(m *memory) {
	s := stats{Total: len(m.Candidates)}
	for _, r := range m.Candidates {
		switch r.Status {
		case "resolves":
			s.Resolving++
		case "unregistered":
			s.Unregistered++
		case "error":
			s.Errors++
		}
	}
	m.Stats = s
}

func ipCluster(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	}
	return ip
}

func mxNote(mx string) string {
	if mx == "" {
		return ""
	}
	return "  [MX]"
}

// httpNote renders the per-row HTTPS/HTTP triage hint for terminal output.
// Format: " https=200→example.com http=conn-refused". Empty when both probes
// are missing (i.e. row hasn't been HTTP-checked yet on an older memory file).
func httpNote(r *candidateRecord) string {
	if r.HTTPS == nil && r.HTTP == nil {
		return ""
	}
	one := func(p *httpProbe) string {
		if p == nil {
			return "n/a"
		}
		if p.Status != 0 {
			// Trim the final URL to bare host so the terminal line stays narrow.
			host := finalHost(p.FinalURL)
			if host != "" {
				return fmt.Sprintf("%d→%s", p.Status, host)
			}
			return fmt.Sprintf("%d", p.Status)
		}
		if p.Error != "" {
			return p.Error
		}
		return "?"
	}
	return fmt.Sprintf("  https=%s http=%s", one(r.HTTPS), one(r.HTTP))
}

// finalHost returns just the host component of a URL, or "" on parse failure.
// Used to keep terminal lines from wrapping when redirects land on long paths.
func finalHost(u string) string {
	if u == "" {
		return ""
	}
	// Crude but stdlib-only and good enough for terminal output.
	s := u
	if i := strings.Index(s, "://"); i != -1 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i != -1 {
		s = s[:i]
	}
	return s
}

// printResolving groups resolving candidates by /24 so parking-provider
// clusters surface as one block instead of many individual lines.
func printResolving(records []*candidateRecord) {
	clusters := map[string][]*candidateRecord{}
	var order []string
	for _, r := range records {
		k := ipCluster(r.IP)
		if _, ok := clusters[k]; !ok {
			order = append(order, k)
		}
		clusters[k] = append(clusters[k], r)
	}
	sort.Slice(order, func(i, j int) bool {
		if len(clusters[order[i]]) != len(clusters[order[j]]) {
			return len(clusters[order[i]]) > len(clusters[order[j]])
		}
		return order[i] < order[j]
	})
	for _, k := range order {
		rs := clusters[k]
		sort.Slice(rs, func(i, j int) bool { return rs[i].Candidate < rs[j].Candidate })
		if len(rs) == 1 {
			r := rs[0]
			fmt.Printf("  %-40s %-18s (%s)%s%s\n", r.Candidate, r.IP, r.Technique, mxNote(r.MX), httpNote(r))
		} else {
			fmt.Printf("  cluster %s (%d names on shared infrastructure):\n", k, len(rs))
			for _, r := range rs {
				fmt.Printf("    %-38s %-18s (%s)%s%s\n", r.Candidate, r.IP, r.Technique, mxNote(r.MX), httpNote(r))
			}
		}
	}
}

func printTransitions(transitions []*candidateRecord) {
	if len(transitions) == 0 {
		return
	}
	sort.Slice(transitions, func(i, j int) bool { return transitions[i].Candidate < transitions[j].Candidate })
	fmt.Printf("\nStatus transitions this run (%d):\n", len(transitions))
	for _, r := range transitions {
		prev := r.PrevStatus
		if prev == "" {
			prev = "new"
		}
		fmt.Printf("  %-40s %s → %s%s\n", r.Candidate, prev, r.Status, mxNote(r.MX))
	}
}

func main() {
	maxCand := flag.Int("max-candidates", defaultMaxCand,
		"max candidates to check this run (pending + stale combined). 0 = no cap.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: typosquat_scan <domain> [-max-candidates N]\n\n")
		fmt.Fprintf(os.Stderr, "Reads ./<domain>-typosquat-memory.json, resolves every row whose\n")
		fmt.Fprintf(os.Stderr, "last_checked is empty (new candidate, not yet resolved) or older than %s,\n", staleAfter)
		fmt.Fprintf(os.Stderr, "then writes results back to the same file. For each row that resolves,\n")
		fmt.Fprintf(os.Stderr, "also probes https:// and http:// (up to %d redirects, %s timeout) and\n", maxRedirects, httpTimeout)
		fmt.Fprintf(os.Stderr, "records status code + final URL per scheme.\n\n")
		fmt.Fprintf(os.Stderr, "New candidates are added to the memory file out-of-band before running\n")
		fmt.Fprintf(os.Stderr, "this script (see SKILL.md).\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	domain := strings.ToLower(flag.Arg(0))
	path := memoryPath(domain)

	mem, index, migrated, err := loadMemory(path, domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if mem.Target == "" {
		mem.Target = domain
	}
	if len(mem.Candidates) == 0 {
		fmt.Fprintf(os.Stderr,
			"Memory file %s is empty or missing.\nGenerate candidates per references/typosquatting_patterns.md and add them to the memory file first.\n",
			path)
		os.Exit(1)
	}

	// Pending = last_checked empty (brand-new candidate added by the LLM).
	// Stale = last_checked older than staleAfter or unparseable.
	// Both qualify; pending goes first so first-time resolution beats
	// rechecking already-known rows under the cap.
	cutoff := time.Now().UTC().Add(-staleAfter)
	var pending, stale []*candidateRecord
	for _, r := range mem.Candidates {
		if r.LastChecked == "" {
			pending = append(pending, r)
			continue
		}
		t, perr := time.Parse(time.RFC3339, r.LastChecked)
		if perr != nil || t.Before(cutoff) {
			stale = append(stale, r)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Candidate < pending[j].Candidate })
	sort.Slice(stale, func(i, j int) bool {
		// Oldest first so the longest-unobserved rows get refreshed first.
		return stale[i].LastChecked < stale[j].LastChecked
	})

	toCheck := append(pending, stale...)
	if *maxCand > 0 && *maxCand < len(toCheck) {
		toCheck = toCheck[:*maxCand]
	}
	fmt.Fprintf(os.Stderr,
		"Memory %s has %d rows for %s (%d pending, %d stale >%s); checking %d this run\n",
		path, len(mem.Candidates), domain, len(pending), len(stale), staleAfter, len(toCheck))

	if len(toCheck) == 0 {
		fmt.Fprintln(os.Stderr, "Nothing to check.")
		if migrated {
			recomputeStats(mem)
			if werr := writeMemory(path, mem); werr != nil {
				fmt.Fprintln(os.Stderr, "warning: could not persist legacy-status migration:", werr)
			} else {
				fmt.Fprintln(os.Stderr, "Migrated legacy 'nxdomain' status values to 'unregistered' in memory file.")
			}
		}
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	resolver := &net.Resolver{}
	jobs := make(chan string)
	out := make(chan lookupResult, len(toCheck))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// One probe client per worker — reuses transport state inside
			// the worker without contending with peers.
			client := newProbeClient()
			for fqdn := range jobs {
				lookupCtx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
				status, ip, mx := resolveOne(lookupCtx, resolver, fqdn)
				cancel()
				res := lookupResult{candidate: fqdn, status: status, ip: ip, mx: mx}
				// HTTP probe only when the name actually resolves; otherwise
				// there's nothing to connect to and we'd just burn timeouts.
				if status == "resolves" {
					// Probe https and http in parallel — independent connections,
					// both signals matter (squats often have only one).
					var pwg sync.WaitGroup
					pwg.Add(2)
					go func() {
						defer pwg.Done()
						res.https = httpProbeOne(client, "https://"+fqdn)
					}()
					go func() {
						defer pwg.Done()
						res.http = httpProbeOne(client, "http://"+fqdn)
					}()
					pwg.Wait()
				}
				out <- res
			}
		}()
	}
	go func() {
		for _, r := range toCheck {
			jobs <- r.Candidate
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()

	var transitions, touched []*candidateRecord
	done := 0
	for r := range out {
		existing := index[r.candidate]
		statusChanged := existing.Status != "" && existing.Status != r.status
		if statusChanged {
			existing.PrevStatus = existing.Status
			existing.PrevChecked = existing.LastChecked
			transitions = append(transitions, existing)
		} else if existing.Status == "" && r.status == "resolves" {
			// First-time observation of a resolving candidate; surface it
			// in the transitions block too — same signal as nxdomain→resolves.
			transitions = append(transitions, existing)
		}
		existing.Status = r.status
		existing.IP = r.ip
		existing.MX = r.mx
		existing.LastChecked = now
		if existing.FirstSeen == "" {
			existing.FirstSeen = now
		}
		// Replace HTTP probe data wholesale on every check — old codes are
		// stale once DNS may have repointed. If the row no longer resolves,
		// clear any prior probe so the report doesn't show stale 200s on
		// what is now NXDOMAIN.
		if r.status == "resolves" {
			existing.HTTPS = r.https
			existing.HTTP = r.http
			existing.HTTPChecked = now
		} else {
			existing.HTTPS = nil
			existing.HTTP = nil
			existing.HTTPChecked = ""
		}
		touched = append(touched, existing)
		done++
		if done%100 == 0 {
			fmt.Fprintf(os.Stderr, "  resolved %d/%d\n", done, len(toCheck))
		}
	}

	sort.Slice(mem.Candidates, func(i, j int) bool { return mem.Candidates[i].Candidate < mem.Candidates[j].Candidate })
	if mem.FirstSeen == "" {
		mem.FirstSeen = now
	}
	mem.LastRun = now
	mem.RunCount++
	recomputeStats(mem)

	if err := writeMemory(path, mem); err != nil {
		fmt.Fprintln(os.Stderr, "error writing memory:", err)
		os.Exit(1)
	}

	var resolving []*candidateRecord
	for _, r := range touched {
		if r.Status == "resolves" {
			resolving = append(resolving, r)
		}
	}
	fmt.Printf("\n%d resolving candidate(s) among %d checked for %s:\n", len(resolving), len(touched), domain)
	printResolving(resolving)
	printTransitions(transitions)
	fmt.Fprintf(os.Stderr,
		"\nWrote %d total rows to %s  (resolving=%d unregistered=%d errors=%d, run #%d)\n",
		mem.Stats.Total, path, mem.Stats.Resolving, mem.Stats.Unregistered, mem.Stats.Errors, mem.RunCount)
}

