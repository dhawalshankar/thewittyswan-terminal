package main

import (
    "fmt"
    "strings"
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
        return "available commands: help, whoami, ls, cat, grep, refresh"

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