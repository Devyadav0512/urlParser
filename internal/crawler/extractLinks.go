package crawler

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// internal/crawler/extractLinks.go
func (c *Crawler) extractLinks(baseURL string, content string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		c.logger.Debug("Failed to parse HTML", "url", baseURL, "error", err)
		return nil
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		c.logger.Debug("Failed to parse base URL", "url", baseURL, "error", err)
		return nil
	}

	var links []string
	seen := make(map[string]bool)

	// Extract links from <a> tags with rate limiting
	doc.Find("a[href]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		select {
		case <-c.ctx.Done():
			return false // Stop processing if context cancelled
		default:
		}

		href, exists := s.Attr("href")
		if !exists {
			return true
		}

		link, err := c.processLink(base, href)
		if err != nil {
			// Skip logging for expected cases
			if !errors.Is(err, ErrExternalDomain) && 
			   !errors.Is(err, ErrInvalidLink) &&
			   !errors.Is(err, ErrNonHTMLResource) {
				c.logger.Debug("Skipping link", "url", href, "error", err)
			}
			return true
		}

		normalized := c.normalizeURL(link)
		if !seen[normalized] {
			seen[normalized] = true
			links = append(links, normalized)
		}
		return true
	})

	return links
}

// processLink converts a relative link to absolute and validates it
func (c *Crawler) processLink(base *url.URL, href string) (string, error) {
    href = strings.TrimSpace(href)
    
    // Skip empty and special links
    if href == "" || href == "#" || strings.HasPrefix(href, "javascript:") {
        return "", ErrInvalidLink
    }

    // Skip non-HTTP(S) links
    if strings.HasPrefix(href, "mailto:") || 
       strings.HasPrefix(href, "tel:") || 
       strings.HasPrefix(href, "data:") || 
       strings.HasPrefix(href, "chrome-extension:") {
        return "", ErrInvalidScheme
    }

    // Parse the href
    linkURL, err := url.Parse(href)
    if err != nil {
        return "", err
    }

    // Resolve relative URLs
    absoluteURL := base.ResolveReference(linkURL)

    // Remove fragments
    absoluteURL.Fragment = ""

    // Skip non-HTTP(S) links
    if absoluteURL.Scheme != "http" && absoluteURL.Scheme != "https" {
        return "", ErrInvalidScheme
    }

    // Skip links to different domains
    if absoluteURL.Host != base.Host {
        c.logger.Debug("Skipping external domain", "url", absoluteURL.String(), "base", base.Host)
        return "", ErrExternalDomain
    }

    // Skip common non-HTML resources
    ext := strings.ToLower(filepath.Ext(absoluteURL.Path))
    if isNonHTMLResource(ext) {
        return "", ErrNonHTMLResource
    }

    return absoluteURL.String(), nil
}

// isNonHTMLResource checks if the extension indicates a non-HTML resource
func isNonHTMLResource(ext string) bool {
	nonHTMLExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".zip": true, ".tar": true, ".gz": true, ".mp3": true, ".mp4": true,
		".avi": true, ".mov": true, ".css": true, ".js": true, ".svg": true,
	}
	return nonHTMLExtensions[ext]
}

// normalizeURL standardizes URLs for comparison
func (c *Crawler) normalizeURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return strings.ToLower(urlStr)
	}

	// Standardize scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Remove fragments and empty queries
	u.Fragment = ""
	if u.RawQuery == "" {
		u.ForceQuery = false
	}

	// Clean path (remove duplicate slashes, etc.)
	u.Path = strings.TrimSuffix(filepath.Clean(u.Path), "/")
	u.Path = strings.ReplaceAll(u.Path, "\\", "/")

	// Sort query parameters for consistent comparison
	if u.RawQuery != "" {
		query := u.Query()
		u.RawQuery = query.Encode()
	}

	return u.String()
}