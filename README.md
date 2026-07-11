# thewittyswan-terminal

The dev-facing half of thewittyswan. A real, interactive terminal in the
browser — networking and security tools, and a browsable feed of the
blog itself. Lives at `thewittyswan.space/terminal`.

Built by **Dhawal**.

## Why this exists

The main site (`thewittyswan`) is for anyone. This one is for the "I'm a
dev" crowd — a place to actually poke at something instead of just
reading about it. Type commands, get real answers, maybe learn something
about how the site itself is built along the way.

## Why Go + xterm.js, not a framework

Same philosophy as the blog repo: simple, boring, easy to reason about.
Go compiles to a single binary — no runtime dependencies, no build
pipeline to fight with, easy to deploy to a plain VM. xterm.js gives a
genuine terminal look (cursor, monospace grid, keyboard capture) without
needing React or any component framework wrapped around it. One Go
process serves both the frontend page and the API — nothing else
running, nothing else to manage.

## Stack

- **Go** — backend, command parsing, HTTP server
- **xterm.js** — terminal emulator, loaded via CDN, no build step
- **SQLite** — planned, for caching and any persistent data (not yet wired in)
- **Groq API** — planned, powers the "swan assistant" panel (not yet built)
- **GCP e2-micro VM** — hosting (Always Free tier, always-on, no cold starts)
- **Caddy** — reverse proxy + automatic HTTPS on the VM
- **systemd** — keeps the Go binary running, auto-restarts on crash/reboot

This is one of two repos. The other, `thewittyswan`, is the Jekyll blog
and portfolio, hosted separately on Vercel. The two are independent —
this terminal reads the blog's public `posts.json` feed rather than
sharing a codebase or database with it.

## Local setup

```powershell
go run main.go blog.go commands.go
```

Visit `http://localhost:8080`.

## Commands (current)

- `help` — list available commands
- `whoami` — basic info about the terminal
- `ls blog/`, `ls blog/netsec`, `ls blog/literature` — list posts from the live blog feed
- `cat <slug or title>` — print a post's content
- `grep <term>` — search post titles and content
- `refresh` — manually re-fetch `posts.json` (blog auto-syncs every 10 minutes otherwise)

## Commands (planned)

- `headers <domain>` — security headers report
- `tls <domain>` — certificate/TLS report
- `dig` / `whois <domain>` — DNS/WHOIS lookups
- `ping <host>` — basic reachability check
- Swan assistant — Groq-backed Q&A panel, triggered by a button, not a typed command

## How the blog connection works

`blog.go` fetches `https://thewittyswan.space/posts.json` on startup and
every 10 minutes after that, caching it in memory. The terminal never
touches the blog's GitHub repo or database directly — just this one
public JSON feed. If the blog is temporarily unreachable, the terminal
keeps serving its last known cache rather than breaking.

## Deployment (planned)

Not yet deployed. Plan: provision a free GCP e2-micro VM, copy the built
Go binary over, run it under systemd, put Caddy in front for HTTPS and
routing (`/terminal` → this app, everything else on the domain →
Vercel's Jekyll deployment).

## License

Personal project, all rights reserved.