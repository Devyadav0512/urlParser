package models

// CrawlResult represents the final output structure
type CrawlResult struct {
	Domain      string   `json:"domain"`
	ProductURLs []string `json:"product_urls"`
}

// Task represents a crawling task
type Task struct {
	URL    string `json:"url"`
	Depth  int    `json:"depth"`
	Domain string `json:"domain"`
}