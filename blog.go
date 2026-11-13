// blog.go
package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Post struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Section string `json:"section"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

var (
	postsCache []Post
	postsMutex sync.RWMutex
)

const postsURL = "https://thewittyswan.space/posts.json"

func fetchPosts() {
	resp, err := http.Get(postsURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return
	}

	postsMutex.Lock()
	postsCache = posts
	postsMutex.Unlock()
}

func getPosts() []Post {
	postsMutex.RLock()
	defer postsMutex.RUnlock()
	return postsCache
}

func startBlogSync() {
	fetchPosts()
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			fetchPosts()
		}
	}()
}
