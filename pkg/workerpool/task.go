package workerpool

// Task represents a unit of work for the crawler
type Task struct {
	URL    string `json:"url"`    // URL to crawl
	Depth  int    `json:"depth"`  // Current depth of crawling
	Domain string `json:"domain"` // Domain being crawled
}

// NewTask creates a new crawling task
func NewTask(url string, depth int, domain string) *Task {
	return &Task{
		URL:    url,
		Depth:  depth,
		Domain: domain,
	}
}