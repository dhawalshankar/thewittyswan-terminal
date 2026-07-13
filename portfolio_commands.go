// portfolio_commands.go
package main

import (
	"fmt"
	"strconv"
	"strings"
)

func row(cols ...string) string {
	const width = 18

	var b strings.Builder
	for _, c := range cols {
		fmt.Fprintf(&b, "%-*s", width, c)
	}
	return b.String()
}

func line() string {
	return strings.Repeat("-", 60)
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

func handleHelp() string {
	return strings.Join([]string{
		line(),
		"",
		"PROFILE",
		row("about", "education", "experience", "achievements"),
		"",
		"CAREER",
		row("resume", "skills", "contact", "timeline"),
		"",
		"DEVELOPMENT",
		row("projects", "github", "leetcode"),
		"",
		"NETSEC BLOGS",
		row("ls netsec", "cat <slug>", "grep <term>", "refresh"),
		"",
		"NETWORK LAB",
		row("headers <host>", "tls <host>", "dig <host>", "ping <host>"),
		"",
		"SYSTEM",
		row("help", "clear", "history"),
		"",
		line(),
	}, "\n")
}
