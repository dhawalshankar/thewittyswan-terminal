// commands.go
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var validHostPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

// isBlockedIP rejects loopback, private, link-local (this also covers
// the 169.254.169.254 cloud metadata endpoint on GCP/AWS/Azure),
// unspecified, and multicast ranges — anything that isn't a normal
// public internet host.
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}

// sanitizeHost strips scheme/trailing slash, rejects anything that
// isn't a plausible hostname, then resolves it and rejects any host
// that points at loopback/private/link-local/metadata addresses.
// This is the single choke point all network-lab commands go through
// before touching net/tls/http — it's what stands between a user
// typing a hostname and this server making a real outbound request
// on their behalf (i.e. it's the SSRF guard).
func sanitizeHost(input string) (string, error) {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimSuffix(input, "/")
	if input == "" || !validHostPattern.MatchString(input) {
		return "", fmt.Errorf("invalid host: %q", input)
	}

	// If it's a literal IP, check it directly.
	if ip := net.ParseIP(input); ip != nil {
		if isBlockedIP(ip) {
			return "", fmt.Errorf("host not allowed: %q", input)
		}
		return input, nil
	}

	// Otherwise resolve the hostname and check every returned address.
	// Rejecting if ANY resolved IP is internal prevents DNS-rebinding
	// tricks where a hostname resolves to a mix of public and private IPs.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, input)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("could not resolve host: %q", input)
	}
	for _, addr := range ips {
		if isBlockedIP(addr.IP) {
			return "", fmt.Errorf("host not allowed: %q", input)
		}
	}

	return input, nil
}

func runCommand(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	parts := strings.Fields(input)
	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help":
		return handleHelp()

	case "whoami":
		return handleWhoAmi()

	case "about":
		return handleAbout()

	case "resume":
		return handleResume(args)

	case "projects":
		return handleProjects(args)

	case "github":
		return handleGithub(args)

	case "leetcode":
		return handleLeetcode(args)

	case "skills", "stack":
		return handleSkills()

	case "contact":
		return handleContact()

	case "education":
		return handleEducation()

	case "experience":
		return handleExperience()

	case "timeline":
		return handleTimeline()

	case "achievements":
		return handleAchievements()

	case "ls":
		return handleLs(args)

	case "cat":
		return handleCat(args)

	case "grep":
		return handleGrep(args)

	case "refresh":
		return handleRefresh()

	case "headers":
		return handleHeaders(args)

	case "tls":
		return handleTLS(args)

	case "dig":
		return handleDig(args)

	case "ping":
		return handlePing(args)

	case "curl":
		return handleCurl(args)

	case "ports":
		return handlePorts(args)

	case "echo":
		return handleEcho(args)

	case "date":
		return handleDate()

	case "uname":
		return handleUname()

	case "pwd":
		return handlePwd()

	case "uptime":
		return handleUptime()

	default:
		return fmt.Sprintf("command not found: %s", cmd)
	}
}

func handleLs(args []string) string {
	posts := getPosts()
	if len(posts) == 0 {
		return "no posts found (blog sync may not have run yet)"
	}

	section := ""
	if len(args) > 0 {
		target := strings.TrimSuffix(strings.TrimPrefix(args[0], "blog/"), "/")
		if target != "blog" && target != "" {
			section = target
		}
	}

	var lines []string
	for _, p := range posts {
		if section != "" && p.Section != section {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s  [%s]  %s", p.Date, p.Section, p.Title))
	}

	if len(lines) == 0 {
		return "no posts found"
	}
	return strings.Join(lines, "\n")
}

func handleCat(args []string) string {
	if len(args) == 0 {
		return "usage: cat <slug>"
	}
	target := args[0]

	for _, p := range getPosts() {
		if strings.Contains(p.Slug, target) || strings.EqualFold(p.Title, target) {
			return fmt.Sprintf("%s\n\n%s", p.Title, p.Content)
		}
	}
	return fmt.Sprintf("no post found matching: %s", target)
}

func handleGrep(args []string) string {
	if len(args) == 0 {
		return "usage: grep <term>"
	}
	term := strings.ToLower(strings.Join(args, " "))

	var matches []string
	for _, p := range getPosts() {
		if strings.Contains(strings.ToLower(p.Content), term) || strings.Contains(strings.ToLower(p.Title), term) {
			matches = append(matches, fmt.Sprintf("%s  [%s]  %s", p.Date, p.Section, p.Title))
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("no matches for: %s", term)
	}
	return strings.Join(matches, "\n")
}

func handleRefresh() string {
	fetchPosts()
	fetchPortfolio()

	return fmt.Sprintf(
		"synced successfully\nblogs: %d\nportfolio: loaded",
		len(getPosts()),
	)
}

func handleHeaders(args []string) string {
	if len(args) == 0 {
		return "usage: headers <domain>"
	}
	domain, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}
	url := "https://" + domain

	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("could not reach %s: %v", domain, err)
	}
	defer resp.Body.Close()

	securityHeaders := []string{
		"Strict-Transport-Security",
		"Content-Security-Policy",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"Referrer-Policy",
		"Permissions-Policy",
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("HTTP %d from %s\n", resp.StatusCode, url))
	for _, h := range securityHeaders {
		v := resp.Header.Get(h)
		if v == "" {
			lines = append(lines, fmt.Sprintf("%-28s MISSING", h))
		} else {
			lines = append(lines, fmt.Sprintf("%-28s %s", h, v))
		}
	}
	return strings.Join(lines, "\n")
}

func handleTLS(args []string) string {
	if len(args) == 0 {
		return "usage: tls <domain>"
	}
	domain, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type dialResult struct {
		conn *tls.Conn
		err  error
	}
	resultCh := make(chan dialResult, 1)

	go func() {
		var d net.Dialer
		rawConn, err := d.DialContext(ctx, "tcp", domain+":443")
		if err != nil {
			resultCh <- dialResult{nil, err}
			return
		}
		tlsConn := tls.Client(rawConn, &tls.Config{ServerName: domain})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			rawConn.Close()
			resultCh <- dialResult{nil, err}
			return
		}
		resultCh <- dialResult{tlsConn, nil}
	}()

	var res dialResult
	select {
	case res = <-resultCh:
	case <-ctx.Done():
		return fmt.Sprintf("could not establish TLS connection to %s: timeout (3s)", domain)
	}

	if res.err != nil {
		return fmt.Sprintf("could not establish TLS connection to %s: %v", domain, res.err)
	}
	defer res.conn.Close()

	cert := res.conn.ConnectionState().PeerCertificates[0]
	versionNames := map[uint16]string{
		tls.VersionTLS10: "TLS 1.0",
		tls.VersionTLS11: "TLS 1.1",
		tls.VersionTLS12: "TLS 1.2",
		tls.VersionTLS13: "TLS 1.3",
	}

	return fmt.Sprintf(
		"domain:      %s\nissuer:      %s\nsubject:     %s\nvalid from:  %s\nvalid until: %s\nprotocol:    %s",
		domain,
		cert.Issuer.CommonName,
		cert.Subject.CommonName,
		cert.NotBefore.Format("2006-01-02"),
		cert.NotAfter.Format("2006-01-02"),
		versionNames[res.conn.ConnectionState().Version],
	)
}

func handleDig(args []string) string {
	if len(args) == 0 {
		return "usage: dig <domain>"
	}
	domain, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := net.Resolver{}
	var lines []string

	ips, err := resolver.LookupHost(ctx, domain)
	if err == nil && len(ips) > 0 {
		lines = append(lines, fmt.Sprintf("A/AAAA: %s", strings.Join(ips, ", ")))
	}

	cname, err := resolver.LookupCNAME(ctx, domain)
	if err == nil && cname != domain+"." {
		lines = append(lines, fmt.Sprintf("CNAME:  %s", cname))
	}

	mxRecords, err := resolver.LookupMX(ctx, domain)
	if err == nil && len(mxRecords) > 0 {
		var mx []string
		for _, r := range mxRecords {
			mx = append(mx, fmt.Sprintf("%s (priority %d)", r.Host, r.Pref))
		}
		lines = append(lines, fmt.Sprintf("MX:     %s", strings.Join(mx, ", ")))
	}

	txtRecords, err := resolver.LookupTXT(ctx, domain)
	if err == nil && len(txtRecords) > 0 {
		lines = append(lines, fmt.Sprintf("TXT:    %s", strings.Join(txtRecords, " | ")))
	}

	if len(lines) == 0 {
		return fmt.Sprintf("no records found for %s (or lookup timed out)", domain)
	}
	return strings.Join(lines, "\n")
}

func handlePing(args []string) string {
	if len(args) == 0 {
		return "usage: ping <host>"
	}
	host, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}

	const attempts = 4
	var results []string
	reachable := 0

	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		type dialResult struct {
			elapsed time.Duration
			err     error
		}
		resultCh := make(chan dialResult, 1)

		start := time.Now()
		go func() {
			var d net.Dialer
			conn, err := d.DialContext(ctx, "tcp", host+":443")
			if err == nil {
				conn.Close()
			}
			resultCh <- dialResult{elapsed: time.Since(start), err: err}
		}()

		select {
		case res := <-resultCh:
			if res.err != nil {
				results = append(results, fmt.Sprintf("attempt %d: unreachable", i+1))
			} else {
				reachable++
				results = append(results, fmt.Sprintf("attempt %d: connected in %v", i+1, res.elapsed.Round(time.Millisecond)))
			}
		case <-ctx.Done():
			results = append(results, fmt.Sprintf("attempt %d: timeout (3s)", i+1))
		}
		cancel()
	}

	header := fmt.Sprintf("TCP reachability check to %s:443 (%d/%d reached) — note: not a true ICMP ping\n", host, reachable, attempts)
	return header + strings.Join(results, "\n")
}

// handleCurl fetches a URL and shows status + a truncated body preview.
// Bounded: 8s timeout, response body capped at 2000 bytes to avoid
// someone using this to pull huge files through your server.
func handleCurl(args []string) string {
	if len(args) == 0 {
		return "usage: curl <url>"
	}
	domain, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}
	url := "https://" + domain

	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("curl: could not reach %s: %v", domain, err)
	}
	defer resp.Body.Close()

	const maxBytes = 2000
	buf := make([]byte, maxBytes)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	truncatedNote := ""
	if n == maxBytes {
		truncatedNote = "\n...[truncated]"
	}

	return fmt.Sprintf("HTTP %d %s\nContent-Type: %s\n\n%s%s",
		resp.StatusCode, http.StatusText(resp.StatusCode),
		resp.Header.Get("Content-Type"), body, truncatedNote)
}

// handlePorts checks a small fixed set of common ports rather than a
// full nmap-style sweep, so this stays a safe connectivity check and
// can't be used to port-scan arbitrary third-party hosts at scale.
func handlePorts(args []string) string {
	if len(args) == 0 {
		return "usage: ports <host>"
	}
	host, err := sanitizeHost(args[0])
	if err != nil {
		return err.Error()
	}

	type portInfo struct {
		port    int
		service string
	}
	commonPorts := []portInfo{
		{21, "FTP"},
		{22, "SSH"},
		{25, "SMTP"},
		{53, "DNS"},
		{80, "HTTP"},
		{110, "POP3"},
		{143, "IMAP"},
		{443, "HTTPS"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{6379, "Redis"},
		{8080, "HTTP-Alt"},
	}

	var lines []string
	for _, p := range commonPorts {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, p.port))
		cancel()

		status := "closed"
		if err == nil {
			status = "open"
			conn.Close()
		}
		lines = append(lines, fmt.Sprintf("%-6d %-12s %s", p.port, p.service, status))
	}

	header := fmt.Sprintf("Port scan (common ports only) — %s\n\n%-6s %-12s %s\n", host, "PORT", "SERVICE", "STATUS")
	return header + strings.Join(lines, "\n")
}

// --- zsh-flavored cosmetic commands ---

func handleEcho(args []string) string {
	return strings.Join(args, " ")
}

func handleDate() string {
	return time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
}

func handleUname() string {
	return "thewittyswan-terminal (portfolio-shell) x86_64"
}

func handlePwd() string {
	return "/home/swan"
}

func handleUptime() string {
	return fmt.Sprintf("up since server start — %s", time.Now().Format("15:04:05 MST"))
}
