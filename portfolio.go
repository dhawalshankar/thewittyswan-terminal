// portfolio.go
package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Profile struct {
	Name      string   `json:"name"`
	Tagline   []string `json:"tagline"`
	Education string   `json:"education_short"`
	Location  string   `json:"location"`
	Summary   string   `json:"summary"`
}

type Education struct {
	Degree    string `json:"degree"`
	Institute string `json:"institute"`
	Duration  string `json:"duration"`
	Details   string `json:"details"`
}

type Experience struct {
	Role        string `json:"role"`
	Company     string `json:"company"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
}

type Project struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Stack       []string `json:"stack"`
	Link        string   `json:"link"`
	Github      string   `json:"github"`
}

type Contact struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
}

type GitHubProfile struct {
	Username string `json:"username"`
	URL      string `json:"url"`
}

type LeetCodeProfile struct {
	Username string `json:"username"`
	URL      string `json:"url"`
}

type LinkedInProfile struct {
	URL string `json:"url"`
}

type Profiles struct {
	GitHub   GitHubProfile   `json:"github"`
	LeetCode LeetCodeProfile `json:"leetcode"`
	LinkedIn LinkedInProfile `json:"linkedin"`
}

type TimelineEntry struct {
	Date  string `json:"date"`
	Event string `json:"event"`
}

type Achievement struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Year        string `json:"year"`
}

type Resume struct {
	Summary string `json:"summary"`
	PDFURL  string `json:"pdf_url"`
}

type Portfolio struct {
	Profile      Profile         `json:"profile"`
	Education    []Education     `json:"education"`
	Experience   []Experience    `json:"experience"`
	Projects     []Project       `json:"projects"`
	Skills       []string        `json:"skills"`
	Contact      Contact         `json:"contact"`
	Profiles     Profiles        `json:"profiles"`
	Timeline     []TimelineEntry `json:"timeline"`
	Achievements []Achievement   `json:"achievements"`
	Resume       Resume          `json:"resume"`
}

var (
	portfolioCache Portfolio
	portfolioMutex sync.RWMutex
)

// Point this at wherever you host portfolio.json — a GitHub repo (raw URL)
// or your own site, exactly like postsURL in blog.go.
const portfolioURL = "https://thewittyswan.space/portfolio.json"

func fetchPortfolio() {
	resp, err := http.Get(portfolioURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var p Portfolio
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return
	}

	portfolioMutex.Lock()
	portfolioCache = p
	portfolioMutex.Unlock()
}

func getPortfolio() Portfolio {
	portfolioMutex.RLock()
	defer portfolioMutex.RUnlock()
	return portfolioCache
}

func startPortfolioSync() {
	fetchPortfolio()
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			fetchPortfolio()
		}
	}()
}
