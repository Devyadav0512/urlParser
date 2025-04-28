package test

import (
	"context"
	"testing"

	"ecommerce-crawler/internal/crawler"
	"ecommerce-crawler/internal/utils"
)

func TestURLPatternMatch(t *testing.T) {
	logger := utils.NewLogger()
	c := crawler.NewCrawler(context.Background(), []string{}, 1, 3, 1, "", "", logger)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "product URL pattern",
			url:      "https://example.com/product/123",
			expected: true,
		},
		{
			name:     "item URL pattern",
			url:      "https://example.com/item/456",
			expected: true,
		},
		{
			name:     "p short URL pattern",
			url:      "https://example.com/p/789",
			expected: true,
		},
		{
			name:     "non-product URL",
			url:      "https://example.com/about",
			expected: false,
		},
		{
			name:     "URL with product in query",
			url:      "https://example.com/page?product_id=123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.URLPatternMatch(tt.url)
			if result != tt.expected {
				t.Errorf("URLPatternMatch(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestIsProductPage(t *testing.T) {
	logger := utils.NewLogger()
	c := crawler.NewCrawler(context.Background(), []string{}, 1, 3, 1, "", "", logger)

	tests := []struct {
		name     string
		url      string
		content  string
		expected bool
	}{
		{
			name:    "product page with og:type",
			url:     "https://example.com/product/123",
			content: `<html><head><meta property="og:type" content="product"></head><body></body></html>`,
			expected: true,
		},
		{
			name:    "product page with schema.org",
			url:     "https://example.com/item/456",
			content: `<html><script type="application/ld+json">{"@type":"Product"}</script></html>`,
			expected: true,
		},
		{
			name:    "non-product page",
			url:     "https://example.com/about",
			content: `<html><body>About Us</body></html>`,
			expected: false,
		},
		{
			name:    "product page with breadcrumbs",
			url:     "https://example.com/p/789",
			content: `<html><body><div class="breadcrumb">Home > Products > Product Name</div></body></html>`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.IsProductPage(tt.url, tt.content)
			if result != tt.expected {
				t.Errorf("IsProductPage(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}