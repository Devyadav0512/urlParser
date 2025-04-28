# E-commerce Product URL Crawler

## 1. Project Objectives

This crawler is designed to:
- Discover product pages across multiple e-commerce websites
- Handle large-scale crawling (millions of URLs) efficiently
- Respect website policies (robots.txt, crawl-delay)
- Identify product pages using multiple detection techniques
- Provide structured output per domain
- Maintain high performance with configurable limits

## 2. How to Run

### Prerequisites
- Go 1.20+
- Redis (for distributed crawling - optional)

### Installation

git clone https://github.com/yourusername/ecommerce-crawler.git
cd ecommerce-crawler
go mod download

### Configuration

Edit configs/crawler.yaml:

domains:
  - https://www.example1.com/
  - https://www.example2.com/
max_workers: 20              # Concurrent workers
max_depth: 3                 # Maximum link depth to follow
max_urls_per_domain: 1000    # Max product URLs per domain
crawl_delay: 1s              # Delay between requests
request_timeout: 30s         # Timeout per request

### Execution

go run cmd/crawler/main.go

### Output

Results are saved in JSON format at outputs/products.json:
{
  "www.example1.com": [
    "https://www.example1.com/product/123",
    "https://www.example1.com/item/456"
  ]
}

## 3. Tech Stack & Architecture

Libraries: 

* goquery: HTML parsing and DOM traversal
* sync/atomic: Thread-safe counters
* sync.Map: Concurrent URL tracking
* workerpool: Concurrent task processing
* yaml: Configuration parsing

Design Patterns:

* Worker Pool Pattern: For concurrent crawling
* Producer-Consumer: URL discovery vs processing
* Strategy Pattern: Multiple product detection techniques
* Observer Pattern: Progress monitoring

Optimization Techniques:

* Concurrent-safe data structures (sync.Map, atomic counters)
* Context-based cancellation
* Exponential backoff for failed requests
* Connection pooling
* Memory-efficient URL storage
* Domain-based rate limiting

Coding Style:

* Clean Go idioms
* Interface-driven design
* Structured logging
* Comprehensive error handling
* Unit testable components

## 4. File Structure

ecommerce-crawler/
├── cmd/
│   └── crawler/
│       └── main.go          # Application entry point
├── configs/
│   └── crawler.yaml         # Configuration template
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration loader
│   ├── crawler/
│   │   ├── crawler.go       # Main crawler logic
│   │   ├── detector.go      # Product detection
│   │   ├── fetcher.go       # HTTP client
│   │   ├── queue.go         # URL queue
│   │   ├── robots.go        # robots.txt parser
│   │   └── sitemap.go       # sitemap.xml parser
│   ├── models/
│   │   └── models.go        # Data structures
│   └── utils/
│       └── logger.go        # Structured logging
├── pkg/
│   └── workerpool/          # Worker pool impl
│       ├── workerpool.go
│       └── task.go
├── test/
│   ├── crawler_test.go      # Integration tests
│   ├── detector_test.go     # Unit tests
│   └── mocks/              # Test mocks
├── outputs/                # Result storage
├── go.mod
└── go.sum

## 5. Workflow Diagram

graph TD
    A[Start Crawler] --> B[Load Config]
    B --> C[Initialize Worker Pool]
    C --> D[Seed Domain URLs]
    D --> E[Worker: Fetch URL]
    E --> F{Check robots.txt?}
    F -->|Yes| G[Parse robots.txt]
    F -->|No| H[Fetch Page]
    G --> H
    H --> I[Detect Product Page]
    I -->|Yes| J[Store Product URL]
    I -->|No| K[Extract Links]
    K --> L[Add to Queue]
    J --> M[Check Limits]
    L --> M
    M --> N{Done?}
    N -->|No| E
    N -->|Yes| O[Generate Output]

Step Explanations:

1. Initialization: Load config and setup worker pool
2. Seed URLs: Start with domain homepages
3. URL Processing:
    * Check robots.txt restrictions
    * Fetch page content
    * Detect product pages using multiple techniques
4. Link Discovery: Extract and queue new links
5. Result Storage: Save product URLs per domain
6. Termination: Stop when limits reached or queue empty

## 6. Monitoring & Testing

Monitoring

View real-time logs:

INFO: Crawling started workers=20 domains=4
DEBUG: Processing URL url=https://www.example.com/ depth=0
WARN: Slow response url=... elapsed=2.1s

Key metrics logged:

* URLs crawled per second
* Product URLs found per domain
* Error rates
* Memory usage

Testing

Run unit tests:

go test -v ./...

Test cases include:

* URL pattern matching
* Product detection
* robots.txt parsing
* Concurrent queue operations
* Error handling

## 7. Outcomes & Metrics

Expected Output

* Structured JSON with product URLs per domain
* Log file with crawl statistics

Key Metrics

  |-----------------------------------------------------------------------------|
  |  Metric	                     |        Target	   |       Measurement      |
  |------------------------------|---------------------|------------------------|
  |  URLs processed/sec	         |         100+	       |  Benchmark test        |
  |  Product detection accuracy	 |         95%+	       |  Manual verification   |
  |  Memory usage	             |    <2GB per 1M URLs |  Profiling             |
  |  Error rate	                 |         <1%	       |  Error logs            |
  |  Domain coverage	         |         100%	       |  Output analysis       |
  |-----------------------------------------------------------------------------|

Performance Optimization

* 50-100 concurrent requests (configurable)
* Domain-specific rate limiting
* Connection reuse
* Efficient duplicate URL detection

Limitations

* JavaScript-rendered content not supported
* May miss some dynamic product URLs
* Rate-limited by target sites

Future Enhancements

* Distributed crawling with Redis
* Headless browser support
* Machine learning for product detection
* Automatic pagination handling