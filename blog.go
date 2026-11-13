// blog.go
package main

import (
	"encoding/json"
	"log"
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

var feedHTTPClient = &http.Client{Timeout: 8 * time.Second}

func fetchPosts() {
	resp, err := feedHTTPClient.Get(postsURL)
	if err != nil {
		log.Printf("fetch blog feed: %v", err)
		return
	}
	defer closeAndLog("blog feed response body", resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf("fetch blog feed: HTTP %d", resp.StatusCode)
		return
	}

	var posts []Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		log.Printf("decode blog feed: %v", err)
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
