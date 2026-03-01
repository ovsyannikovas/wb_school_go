package main

import (
	"net/url"
	"sync"
	"time"
)

type Config struct {
	BaseURL   string
	MaxDepth  int
	Parallel  int
	OutputDir string
	Timeout   time.Duration
	Visited   map[string]bool
	Domain    string
	mu        sync.RWMutex
}

func (c *Config) IsVisited(url string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Visited[url]
}

func (c *Config) MarkVisited(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Visited[url] = true
}

func (c *Config) IsSameDomain(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Hostname() == c.Domain || parsed.Hostname() == ""
}
