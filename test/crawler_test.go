package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"ecommerce-crawler/internal/crawler"
	"ecommerce-crawler/internal/utils"
)

func TestCrawlerBasicFunctionality(t *testing.T) {
	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<html><body><a href="/product/1">Product 1</a></body></html>`))
		case "/product/1":
			w.Write([]byte(`<html><head><meta property="og:type" content="product"></head></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Parse test server URL to get host
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}

	logger := utils.NewLogger()
	c := crawler.NewCrawler(
		context.Background(),
		[]string{ts.URL},
		5, // workers
		3, // maxDepth
		time.Millisecond, // crawlDelay
		"test-crawler",
		"test_output.json",
		logger,
	)

	// Run crawler with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = c.Start(ctx)
	if err != nil {
		t.Fatalf("Crawler failed: %v", err)
	}

	// Verify results
	results := c.GetProductURLs()
	if len(results[tsURL.Host]) != 1 {
		t.Errorf("Expected 1 product URL, got %d", len(results[tsURL.Host]))
	}
}

func TestRobotsTxtHandling(t *testing.T) {
	// Setup test server with robots.txt
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			w.Write([]byte("User-agent: *\nDisallow: /disallowed/"))
		case "/allowed":
			w.Write([]byte(`<html><head><meta property="og:type" content="product"></head></html>`))
		case "/disallowed/product":
			w.Write([]byte(`<html><head><meta property="og:type" content="product"></head></html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Parse test server URL to get host
	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}

	logger := utils.NewLogger()
	c := crawler.NewCrawler(
		context.Background(),
		[]string{ts.URL + "/allowed", ts.URL + "/disallowed/product"},
		1, // workers
		1, // maxDepth
		time.Millisecond, // crawlDelay
		"test-crawler",
		"test_output.json",
		logger,
	)

	err = c.Start(context.Background())
	if err != nil {
		t.Fatalf("Crawler failed: %v", err)
	}

	// Verify results
	results := c.GetProductURLs()
	if len(results[tsURL.Host]) != 1 {
		t.Errorf("Expected 1 product URL (only allowed one), got %d", len(results[tsURL.Host]))
	}
}