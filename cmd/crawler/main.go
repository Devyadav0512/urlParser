package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-crawler/internal/crawler"
	"ecommerce-crawler/internal/utils"
)

func main() {
	// Set up root context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

	// Initialize logger
	logger := utils.NewLogger()

	// Configuration
	domains := []string{
		"https://www.virgio.com/",
		"https://www.tatacliq.com/",
		"https://nykaafashion.com/",
		"https://www.westside.com/",
	}
	maxWorkers := 10
	maxDepth := 3
	crawlDelay := 1 * time.Second
	userAgent := "EcommerceCrawler/1.0 (+https://github.com/yourusername/ecommerce-crawler)"
	outputFile := "outputs/product_urls.json"

	// Create crawler instance
	crawler := crawler.NewCrawler(
		ctx, 
		domains,
		maxWorkers,
		maxDepth,
		crawlDelay,
		userAgent,
		outputFile,
		logger,
	)

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start crawler in a separate goroutine
	go func() {
		logger.Info("Starting crawler...")
		if err := crawler.Start(ctx); err != nil {
			logger.Error("Crawler error", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Received shutdown signal, stopping crawler...")
	cancel()

	// Give some time for cleanup
	time.Sleep(2 * time.Second)
	logger.Info("Crawler stopped successfully")
}