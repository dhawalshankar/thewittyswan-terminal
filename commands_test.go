// commands_test.go
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/time/rate"
)

func TestRunCommandCatMatchesMultiwordTitle(t *testing.T) {
	postsMutex.Lock()
	originalPosts := postsCache
	postsCache = []Post{{
		Slug:    "use-ai-only-where-it-belongs",
		Title:   "Use AI only where it belongs!",
		Content: "article body",
	}}
	postsMutex.Unlock()
	t.Cleanup(func() {
		postsMutex.Lock()
		postsCache = originalPosts
		postsMutex.Unlock()
	})

	for _, command := range []string{
		`cat Use AI only where it belongs`,
		`cat "Use AI only where it belongs"`,
	} {
		got := runCommand(command)
		want := "Use AI only where it belongs!\n\narticle body"
		if got != want {
			t.Errorf("runCommand(%q) = %q, want %q", command, got, want)
		}
	}
}

func TestClientIPTrustsForwardedHeaderOnlyFromLoopback(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{
			name:       "direct client cannot spoof forwarded address",
			remoteAddr: "198.51.100.10:1234",
			forwarded:  "203.0.113.7",
			want:       "198.51.100.10",
		},
		{
			name:       "local proxy supplies client address",
			remoteAddr: "127.0.0.1:1234",
			forwarded:  "198.51.100.10, 203.0.113.7",
			want:       "203.0.113.7",
		},
		{
			name:       "invalid forwarded address falls back to proxy",
			remoteAddr: "[::1]:1234",
			forwarded:  "not-an-ip",
			want:       "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			r.Header.Set("X-Forwarded-For", tt.forwarded)
			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	handler := chain(healthHandler, securityHeaders)

	get := httptest.NewRecorder()
	handler(get, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if get.Code != http.StatusNoContent {
		t.Fatalf("GET /healthz status = %d, want %d", get.Code, http.StatusNoContent)
	}
	if got := get.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}

	post := httptest.NewRecorder()
	handler(post, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if post.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /healthz status = %d, want %d", post.Code, http.StatusMethodNotAllowed)
	}
	if got := post.Header().Get("Allow"); got != http.MethodGet {
		t.Errorf("Allow = %q, want %q", got, http.MethodGet)
	}
}

func TestRunHandlerRejectsOversizedBody(t *testing.T) {
	limiter := &rateLimiter{
		visitors: make(map[string]*visitor),
		rps:      rate.Inf,
		burst:    1,
	}
	handler := runHandler(limiter)
	body := `{"command":"` + strings.Repeat("x", maxCommandRequestBytes) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
	response := httptest.NewRecorder()

	handler(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("POST /run status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestSanitizeHostRejectsPrivateAndInvalidAddresses(t *testing.T) {
	for _, input := range []string{
		"127.0.0.1",
		"10.0.0.1",
		"169.254.169.254",
		"::1",
		"not a host",
	} {
		if got, err := sanitizeHost(input); err == nil {
			t.Errorf("sanitizeHost(%q) = %q, want error", input, got)
		}
	}

	if got, err := sanitizeHost("8.8.8.8"); err != nil || got != "8.8.8.8" {
		t.Errorf("sanitizeHost(public IP) = %q, %v; want 8.8.8.8, nil", got, err)
	}
}
