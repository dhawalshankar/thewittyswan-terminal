// portfolio_commands.go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

func line() string {
	return strings.Repeat("-", 62)
}

// cmdEntry is one row in the help output: the command syntax on the
// left, a short description on the right.
type cmdEntry struct {
	syntax string
	desc   string
}

// section renders a titled block of aligned command entries. Column
// width is computed from the longest syntax string IN THAT SECTION,
// not a fixed global width, so short sections (like SHELL) stay
// compact and long sections (like NETWORK LAB) don't wrap or misalign.
func section(title string, entries []cmdEntry) string {
	maxLen := 0
	for _, e := range entries {
		if len(e.syntax) > maxLen {
			maxLen = len(e.syntax)
		}
	}
	width := maxLen + 3 // padding between syntax and description

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "  %-*s%s\n", width, e.syntax, e.desc)
	}
	return b.String()
}

func handleWhoAmi() string {
	return "Dhawal Shukla — Backend Engineering • Network Security • Cloud"
}

func handleGithub(_ []string) string {
	p := getPortfolio()

	if p.Profiles.GitHub.URL == "" {
		return "github profile not configured"
	}

	return fmt.Sprintf(
		"GitHub\n\nUsername: %s\nProfile:  %s",
		p.Profiles.GitHub.Username,
		p.Profiles.GitHub.URL,
	)
}

func handleLeetcode(_ []string) string {
	p := getPortfolio()

	if p.Profiles.LeetCode.URL == "" {
		return "leetcode profile not configured"
	}

	return fmt.Sprintf(
		"LeetCode\n\nUsername: %s\nProfile:  %s",
		p.Profiles.LeetCode.Username,
		p.Profiles.LeetCode.URL,
	)
}

func handleAbout() string {
	p := getPortfolio()
	if p.Profile.Name == "" {
		return "portfolio data not loaded yet — try again in a moment"
	}
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		p.Profile.Name,
		strings.Join(p.Profile.Tagline, "\n"),
		p.Profile.Education,
		p.Profile.Summary,
	)
}

func handleResume(args []string) string {
	p := getPortfolio()

	if len(args) > 0 {
		switch args[0] {
		case "pdf", "open", "download":
			if p.Resume.PDFURL == "" {
				return "resume PDF link not configured"
			}
			return fmt.Sprintf("opening resume: %s", p.Resume.PDFURL)
		}
	}

	var b strings.Builder
	b.WriteString(p.Resume.Summary)
	b.WriteString("\n\ntype: resume pdf   — to get the download link\n")
	b.WriteString("type: education / experience / projects / skills — for details on each section")
	return b.String()
}

func handleProjects(args []string) string {
	p := getPortfolio()
	if len(p.Projects) == 0 {
		return "no projects found (portfolio sync may not have run yet)"
	}

	if len(args) > 0 {
		idx, err := strconv.Atoi(args[0])
		if err != nil || idx < 1 || idx > len(p.Projects) {
			return fmt.Sprintf("invalid project number. type 'projects' to see the list (1-%d)", len(p.Projects))
		}
		proj := p.Projects[idx-1]
		var b strings.Builder
		b.WriteString(proj.Name)
		b.WriteString("\n\n")
		b.WriteString(proj.Description)
		b.WriteString("\n\n")
		if len(proj.Stack) > 0 {
			b.WriteString("stack:  ")
			b.WriteString(strings.Join(proj.Stack, ", "))
			b.WriteString("\n")
		}
		if proj.Link != "" {
			b.WriteString("link:   ")
			b.WriteString(proj.Link)
			b.WriteString("\n")
		}
		if proj.Github != "" {
			b.WriteString("github: ")
			b.WriteString(proj.Github)
		}
		return b.String()
	}

	var lines []string
	for i, proj := range p.Projects {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, proj.Name))
	}
	lines = append(lines, "\ntype: projects <n>   — for details on a specific project")
	return strings.Join(lines, "\n")
}

func handleSkills() string {
	p := getPortfolio()
	if len(p.Skills) == 0 {
		return "skills data not loaded yet"
	}
	return strings.Join(p.Skills, "\n")
}

func handleContact() string {
	p := getPortfolio()

	var lines []string

	if p.Contact.Email != "" {
		lines = append(lines, "email:     "+p.Contact.Email)
	}

	if p.Contact.Phone != "" {
		lines = append(lines, "phone:     "+p.Contact.Phone)
	}

	if p.Contact.Location != "" {
		lines = append(lines, "location:  "+p.Contact.Location)
	}

	if p.Profiles.LinkedIn.URL != "" {
		lines = append(lines, "linkedin:  "+p.Profiles.LinkedIn.URL)
	}

	if p.Profiles.GitHub.URL != "" {
		lines = append(lines, "github:    "+p.Profiles.GitHub.URL)
	}

	if p.Profiles.LeetCode.URL != "" {
		lines = append(lines, "leetcode:  "+p.Profiles.LeetCode.URL)
	}

	if len(lines) == 0 {
		return "contact data not loaded yet"
	}

	return strings.Join(lines, "\n")
}

func handleAchievements() string {
	p := getPortfolio()

	if len(p.Achievements) == 0 {
		return "achievements data not loaded yet"
	}

	var lines []string
	for _, a := range p.Achievements {
		lines = append(lines,
			fmt.Sprintf("%s (%s)\n%s\n", a.Title, a.Year, a.Description),
		)
	}

	return strings.Join(lines, "\n")
}

func handleEducation() string {
	p := getPortfolio()
	if len(p.Education) == 0 {
		return "education data not loaded yet"
	}
	var lines []string
	for _, e := range p.Education {
		lines = append(lines, fmt.Sprintf("%s\n%s — %s\n%s\n", e.Degree, e.Institute, e.Duration, e.Details))
	}
	return strings.Join(lines, "\n")
}

func handleExperience() string {
	p := getPortfolio()
	if len(p.Experience) == 0 {
		return "experience data not loaded yet"
	}
	var lines []string
	for _, e := range p.Experience {
		lines = append(lines, fmt.Sprintf("%s @ %s (%s)\n%s\n", e.Role, e.Company, e.Duration, e.Description))
	}
	return strings.Join(lines, "\n")
}

func handleTimeline() string {
	p := getPortfolio()
	if len(p.Timeline) == 0 {
		return "timeline data not loaded yet"
	}
	var lines []string
	for _, t := range p.Timeline {
		lines = append(lines, fmt.Sprintf("%s — %s", t.Date, t.Event))
	}
	return strings.Join(lines, "\n")
}

// handleHelp renders every command the terminal understands, grouped
// by category, each section aligned to its own longest entry so
// nothing wraps or drifts out of column regardless of how long an
// individual command's syntax string is.
//
// Note: mkdir / touch / cd / rm / ls (no args) / pwd / echo redirects /
// edit / clear / history are all handled client-side in terminal.js
// (the virtual filesystem), never hitting runCommand here — they're
// still listed below so `help` stays a complete reference either way.
func handleHelp() string {
	var out []string
	out = append(out, line(), "")

	out = append(out, section("PROFILE", []cmdEntry{
		{"about", "who I am, short bio"},
		{"education", "degrees and institutions"},
		{"experience", "work history"},
		{"achievements", "awards, notable results"},
	}))

	out = append(out, section("CAREER", []cmdEntry{
		{"resume", "resume summary"},
		{"resume pdf", "get the resume PDF link"},
		{"skills", "tech stack / skills list"},
		{"contact", "email, socials, links"},
		{"timeline", "career/education timeline"},
	}))

	out = append(out, section("DEVELOPMENT", []cmdEntry{
		{"projects", "list all projects"},
		{"projects <n>", "details on project n"},
		{"github", "GitHub profile"},
		{"leetcode", "LeetCode profile"},
	}))

	out = append(out, section("NETSEC BLOGS", []cmdEntry{
		{"ls blog", "list all blog posts"},
		{"ls <section>", "list posts in a section, e.g. ls netsec"},
		{"cat <slug>", "read a blog post"},
		{"grep <term>", "search blog posts by keyword"},
	}))

	out = append(out, section("NETWORK LAB", []cmdEntry{
		{"headers <host>", "check security headers + cookie flags"},
		{"tls <host>", "inspect TLS cert, protocol, cipher, expiry"},
		{"dig <host>", "DNS records: A/AAAA, CNAME, MX, TXT"},
		{"ping <host>", "TCP reachability check (not true ICMP)"},
		{"curl <url>", "fetch a URL, show status + body preview"},
		{"ports <host>", "scan common ports for open/closed"},
		{"pkgcheck <eco> <pkg>", "check known CVEs for a package (npm/pip/go/cargo)"},
	}))

	out = append(out, section("FILESYSTEM (virtual, saved in your browser)", []cmdEntry{
		{"ls", "list files/dirs in current directory"},
		{"cd <dir>", "change directory (.. to go up, / for root)"},
		{"pwd", "print working directory"},
		{"mkdir <dir>", "create a directory"},
		{"touch <file>", "create an empty file"},
		{"cat <file>", "print a file's contents"},
		{"edit <file>", "open the line editor (. or Ctrl+D to save)"},
		{"echo <text> > f", "write text to a file (overwrite)"},
		{"echo <text> >> f", "append text to a file"},
		{"rm <name>", "delete a file or directory"},
	}))

	out = append(out, section("SHELL", []cmdEntry{
		{"echo <text>", "print text back"},
		{"date", "current date and time"},
		{"uname", "system info"},
		{"uptime", "server uptime"},
	}))

	out = append(out, section("SYSTEM", []cmdEntry{
		{"help", "show this menu"},
		{"clear", "clear the screen"},
		{"history", "show command history"},
		{"refresh", "resync blog posts + portfolio data"},
	}))

	out = append(out, line())
	return strings.Join(out, "\n")
}
