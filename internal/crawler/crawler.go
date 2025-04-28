package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ecommerce-crawler/internal/utils"
	"ecommerce-crawler/pkg/workerpool"
)

type Crawler struct {
    ctx         context.Context     // Add this line
    domains     []string
    workerPool  *workerpool.WorkerPool
    visitedURLs *sync.Map
    productURLs *DomainURLMap
    httpClient  *HTTPClient
    userAgent   string
    maxDepth    int
    crawlDelay  time.Duration
    outputFile  string
    logger      *utils.Logger
}

type DomainURLMap struct {
	sync.Map
}

func (m *DomainURLMap) Add(domain, url string) {
	urls, _ := m.LoadOrStore(domain, &sync.Map{})
	urls.(*sync.Map).Store(url, true)
}

func (m *DomainURLMap) ToJSON() map[string][]string {
	result := make(map[string][]string)
	m.Range(func(key, value interface{}) bool {
		domain := key.(string)
		urls := value.(*sync.Map)
		var urlList []string
		urls.Range(func(url, _ interface{}) bool {
			urlList = append(urlList, url.(string))
			return true
		})
		result[domain] = urlList
		return true
	})
	return result
}

func NewCrawler(
	ctx context.Context,
	domains []string,
	maxWorkers, maxDepth int,
	crawlDelay time.Duration,
	userAgent, outputFile string,
	logger *utils.Logger,
) *Crawler {
	return &Crawler{
		ctx:         ctx,
		domains:     domains,
		workerPool:  workerpool.NewWorkerPool(maxWorkers, 30*time.Second), // 30s timeout per task
		visitedURLs: &sync.Map{},
		productURLs: &DomainURLMap{},
		httpClient:  NewHTTPClient(logger),
		userAgent:   userAgent,
		maxDepth:    maxDepth,
		crawlDelay:  crawlDelay,
		outputFile:  outputFile,
		logger:      logger,
	}
}

// internal/crawler/crawler.go
func (c *Crawler) Start(ctx context.Context) error {
	// Initialize queue
	for _, domain := range c.domains {
		parsedURL, err := url.Parse(domain)
		if err != nil {
			continue
		}
		c.workerPool.AddTask(&workerpool.Task{
			URL:    domain,
			Depth:  0,
			Domain: parsedURL.Host,
		})
	}

	// Start heartbeat monitor
	go c.monitor(ctx)

	// Start processing
	go c.workerPool.Run(ctx, c.processTask, c.logger)

	// Wait for completion
	<-ctx.Done()
	return c.generateOutput()
}

func (c *Crawler) monitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	lastCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentCount := c.visitedCount()
			if currentCount == lastCount {
				c.logger.Warn("Crawler stalled - no progress in last 30 seconds",
					"visitedCount", currentCount,
					"productCount", c.productCount(),
				)
				// Optional: Add recovery logic here if needed
			}
			lastCount = currentCount
		}
	}
}

func (c *Crawler) productCount() int {
	count := 0
	c.productURLs.Range(func(_, urls interface{}) bool {
		urls.(*sync.Map).Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		return true
	})
	return count
}

func (c *Crawler) visitedCount() int {
	count := 0
	c.visitedURLs.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (c *Crawler) processTask(task *workerpool.Task) error {
	// Check context cancellation
    if err := c.ctx.Err(); err != nil {
        return err
    }

	// Normalize URL first
	normalizedURL := c.normalizeURL(task.URL)

	// Check if we've already visited this URL
	if _, loaded := c.visitedURLs.LoadOrStore(normalizedURL, true); loaded {
		return nil
	}

	c.logger.Debug("Processing URL", "url", normalizedURL, "depth", task.Depth)

	// Check robots.txt first
	robotsAllowed, crawlDelay, err := c.checkRobotsTxt(task)
	if err != nil {
		c.logger.Error("Robots.txt check failed", "url", normalizedURL, "error", err)
		return err
	}
	if !robotsAllowed {
		c.logger.Debug("URL disallowed by robots.txt", "url", normalizedURL)
		return nil
	}

	// Respect crawl delay
	time.Sleep(crawlDelay)

	// Check sitemap first if we're at the root
	if task.Depth == 0 {
		sitemapURLs, err := c.checkSitemap(task.URL)
		if err == nil && len(sitemapURLs) > 0 {
			for _, u := range sitemapURLs {
				// Don't follow sitemap links deeper than max depth
				if task.Depth+1 <= c.maxDepth {
					c.workerPool.AddTask(&workerpool.Task{
						URL:    u,
						Depth:  task.Depth + 1,
						Domain: task.Domain,
					})
				}
			}
		}
	}

	// Fetch the page with timeout
	_, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	content, err := c.httpClient.FetchWithContext(c.ctx, task.URL)
    if err != nil {
        if errors.Is(err, ErrTimeout) {
            c.logger.Warn("Timeout while fetching URL",
                "url", task.URL,
                "depth", task.Depth)
            return nil // Skip this URL but continue processing
        }
        return err
    }

	// Detect if this is a product page
	if c.IsProductPage(normalizedURL, content) {
		c.productURLs.Add(task.Domain, normalizedURL)
		c.logger.Info("Found product page", "url", normalizedURL)
		// Don't crawl further from product pages
		return nil
	}

	// Extract links and add to queue if we haven't reached max depth
	if task.Depth < c.maxDepth {
		links := c.extractLinks(normalizedURL, content)
		for _, link := range links {
			c.workerPool.AddTask(&workerpool.Task{
				URL:    link,
				Depth:  task.Depth + 1,
				Domain: task.Domain,
			})
		}
	}

	return nil
}

func (c *Crawler) generateOutput() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(c.outputFile), 0755); err != nil {
		return err
	}

	// Convert product URLs to JSON structure
	outputData := c.productURLs.ToJSON()

	// Write to file
	file, err := os.Create(c.outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(outputData); err != nil {
		return err
	}

	c.logger.Info("Output written successfully", "file", c.outputFile)
	return nil
}

// GetProductURLs returns a map of domains to their product URLs
func (c *Crawler) GetProductURLs() map[string][]string {
    return c.productURLs.ToJSON()
}

// GetVisitedURLs returns all visited URLs
func (c *Crawler) GetVisitedURLs() []string {
    var urls []string
    c.visitedURLs.Range(func(key, value interface{}) bool {
        urls = append(urls, key.(string))
        return true
    })
    return urls
}