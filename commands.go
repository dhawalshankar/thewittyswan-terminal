package main

import (
    "crypto/tls"
    "fmt"
    "net"
    "net/http"
    "strings"
    "time"
)

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
        return "available commands: help, whoami, ls, cat, grep, refresh, headers, tls, dig, ping"

    case "whoami":
        return "thewittyswan terminal — network security & literature, self-hosted on a free GCP VM"

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
    return fmt.Sprintf("blog synced — %d posts loaded", len(getPosts()))
}

func handleHeaders(args []string) string {
    if len(args) == 0 {
        return "usage: headers <domain>"
    }
    domain := args[0]
    url := "https://" + strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://")

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
    domain := strings.TrimPrefix(strings.TrimPrefix(args[0], "https://"), "http://")
    domain = strings.TrimSuffix(domain, "/")

    conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 8 * time.Second}, "tcp", domain+":443", &tls.Config{ServerName: domain})
    if err != nil {
        return fmt.Sprintf("could not establish TLS connection to %s: %v", domain, err)
    }
    defer conn.Close()

    cert := conn.ConnectionState().PeerCertificates[0]
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
        versionNames[conn.ConnectionState().Version],
    )
}

func handleDig(args []string) string {
    if len(args) == 0 {
        return "usage: dig <domain>"
    }
    domain := args[0]

    var lines []string

    ips, err := net.LookupHost(domain)
    if err == nil && len(ips) > 0 {
        lines = append(lines, fmt.Sprintf("A/AAAA: %s", strings.Join(ips, ", ")))
    }

    cname, err := net.LookupCNAME(domain)
    if err == nil && cname != domain+"." {
        lines = append(lines, fmt.Sprintf("CNAME:  %s", cname))
    }

    mxRecords, err := net.LookupMX(domain)
    if err == nil && len(mxRecords) > 0 {
        var mx []string
        for _, r := range mxRecords {
            mx = append(mx, fmt.Sprintf("%s (priority %d)", r.Host, r.Pref))
        }
        lines = append(lines, fmt.Sprintf("MX:     %s", strings.Join(mx, ", ")))
    }

    txtRecords, err := net.LookupTXT(domain)
    if err == nil && len(txtRecords) > 0 {
        lines = append(lines, fmt.Sprintf("TXT:    %s", strings.Join(txtRecords, " | ")))
    }

    if len(lines) == 0 {
        return fmt.Sprintf("no records found for %s (or lookup failed)", domain)
    }
    return strings.Join(lines, "\n")
}

func handlePing(args []string) string {
    if len(args) == 0 {
        return "usage: ping <host>"
    }
    host := args[0]

    const attempts = 4
    var results []string
    reachable := 0

    for i := 0; i < attempts; i++ {
        start := time.Now()
        conn, err := net.DialTimeout("tcp", host+":443", 4*time.Second)
        elapsed := time.Since(start)

        if err != nil {
            results = append(results, fmt.Sprintf("attempt %d: unreachable", i+1))
            continue
        }
        conn.Close()
        reachable++
        results = append(results, fmt.Sprintf("attempt %d: connected in %v", i+1, elapsed.Round(time.Millisecond)))
    }

    header := fmt.Sprintf("TCP reachability check to %s:443 (%d/%d reached) — note: not a true ICMP ping\n", host, reachable, attempts)
    return header + strings.Join(results, "\n")
}