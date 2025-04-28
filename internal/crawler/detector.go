package crawler

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Crawler) IsProductPage(urlStr string, content string) bool {
	score := 0

	// Technique a: Regex-based and heuristic-based filter
	if c.URLPatternMatch(urlStr) {
		score += 20
	}

	// Parse HTML content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		c.logger.Error("Failed to parse HTML", "url", urlStr, "error", err)
		return false
	}

	// Technique b: Meta tags and breadcrumb navigation
	if c.checkMetaTags(doc) {
		score += 15
	}
	if c.checkBreadcrumbs(doc) {
		score += 10
	}

	// Technique c: URL Query Parameters
	if c.checkQueryParams(urlStr) {
		score += 10
	}

	// Technique d: Anchor Text or Button Text
	if c.checkAnchorTexts(doc) {
		score += 10
	}

	// Technique e: Structured Data (Schema.org)
	if c.checkStructuredData(doc) {
		score += 20
	}

	// Technique f: Analyzing Canonical Tags
	if c.checkCanonicalTags(doc) {
		score += 10
	}

	// Technique h: Anchor Density
	if c.checkAnchorDensity(doc) {
		score += 5
	}

	c.logger.Debug("Product detection score", "url", urlStr, "score", score)

	// Threshold can be adjusted based on requirements
	return score >= 50
}

func (c *Crawler) URLPatternMatch(urlStr string) bool {
	patterns := []string{
		`/product/`,
		`/item/`,
		`/p/`,
		`/prod/`,
		`-prod\d+`,
		`/buy/`,
		`/shop/`,
		`/product\.html`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, strings.ToLower(urlStr))
		if matched {
			return true
		}
	}
	return false
}

func (c *Crawler) checkMetaTags(doc *goquery.Document) bool {
	// Check for og:type product
	ogType, exists := doc.Find("meta[property='og:type']").Attr("content")
	if exists && strings.ToLower(ogType) == "product" {
		return true
	}

	// Check for other commerce-related meta tags
	found := false
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); strings.Contains(strings.ToLower(name), "product") {
			found = true
		}
		if property, _ := s.Attr("property"); strings.Contains(strings.ToLower(property), "product") {
			found = true
		}
	})

	return found
}

func (c *Crawler) checkBreadcrumbs(doc *goquery.Document) bool {
	// Look for breadcrumb navigation containing "product"
	breadcrumbs := doc.Find(".breadcrumb, .breadcrumbs, .bc, .breadcrumb-trail")
	if breadcrumbs.Length() == 0 {
		return false
	}

	breadcrumbText := strings.ToLower(breadcrumbs.Text())
	return strings.Contains(breadcrumbText, "product") || 
		strings.Contains(breadcrumbText, "item") || 
		strings.Contains(breadcrumbText, "detail")
}

func (c *Crawler) checkQueryParams(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	queryParams := u.Query()
	for param := range queryParams {
		lowerParam := strings.ToLower(param)
		if strings.Contains(lowerParam, "product") || 
			strings.Contains(lowerParam, "item") || 
			strings.Contains(lowerParam, "prod") || 
			strings.Contains(lowerParam, "sku") {
			return true
		}
	}
	return false
}

func (c *Crawler) checkAnchorTexts(doc *goquery.Document) bool {
	// Look for product-related anchor texts
	productPhrases := []string{
		"buy now",
		"add to cart",
		"add to bag",
		"view product",
		"product details",
		"shop now",
	}

	found := false
	doc.Find("a, button").Each(func(i int, s *goquery.Selection) {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		for _, phrase := range productPhrases {
			if strings.Contains(text, phrase) {
				found = true
				return
			}
		}
	})

	return found
}

func (c *Crawler) checkStructuredData(doc *goquery.Document) bool {
	// Check for Schema.org Product markup
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		// In a real implementation, we would parse the JSON-LD and check for @type: Product
		// This is a simplified version
		content := strings.ToLower(s.Text())
		if strings.Contains(content, `"@type":"product"`) || 
			strings.Contains(content, `'@type':'product'`) {
			return
		}
	})
	return false
}

func (c *Crawler) checkCanonicalTags(doc *goquery.Document) bool {
	// Check if canonical URL matches product patterns
	canonical, exists := doc.Find("link[rel='canonical']").Attr("href")
	if !exists {
		return false
	}
	return c.URLPatternMatch(canonical)
}

func (c *Crawler) checkAnchorDensity(doc *goquery.Document) bool {
	// Count all elements and anchor elements
	totalElements := 0
	anchorElements := 0

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		totalElements++
		if goquery.NodeName(s) == "a" {
			anchorElements++
		}
	})

	if totalElements == 0 {
		return false
	}

	// Product pages tend to have higher anchor density
	density := float64(anchorElements) / float64(totalElements)
	return density > 0.3
}