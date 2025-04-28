package crawler

import (
	"ecommerce-crawler/pkg/workerpool"
	"net/http"
	"net/url"
	"time"

	"github.com/temoto/robotstxt"
)

func (c *Crawler) checkRobotsTxt(task *workerpool.Task) (bool, time.Duration, error) {
	// Get robots.txt URL
	robotsURL, err := url.Parse(task.URL)
	if err != nil {
		return false, c.crawlDelay, err
	}
	robotsURL.Path = "/robots.txt"

	// Fetch robots.txt
	resp, err := http.Get(robotsURL.String())
	if err != nil {
		// If robots.txt doesn't exist, assume all paths are allowed
		return true, c.crawlDelay, nil
	}
	defer resp.Body.Close()

	// Parse robots.txt
	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		return true, c.crawlDelay, nil
	}

	// Check if our user agent is allowed to access this URL
	group := data.FindGroup(c.userAgent)
	if group == nil {
		// No specific rules for our user agent, use default
		return true, c.crawlDelay, nil
	}

	// Check if the path is allowed
	allowed := group.Test(task.URL)
	if !allowed {
		return false, c.crawlDelay, nil
	}

	// Get crawl delay if specified
	crawlDelay := c.crawlDelay
	if delay := group.CrawlDelay; delay > 0 {
		crawlDelay = time.Duration(delay * time.Second)
	}

	return true, crawlDelay, nil
}