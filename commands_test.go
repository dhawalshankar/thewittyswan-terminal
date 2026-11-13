// commands_test.go
package main

import "testing"

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
