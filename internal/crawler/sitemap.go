package crawler

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type sitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

type urlset struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func (c *Crawler) checkSitemap(domain string) ([]string, error) {
	sitemapURLs := []string{
		domain + "/sitemap.xml",
		domain + "/sitemap_index.xml",
		domain + "/sitemap-index.xml",
	}

	var productURLs []string

	for _, sitemapURL := range sitemapURLs {
		resp, err := http.Get(sitemapURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			continue
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read sitemap: %w", err)
		}

		// Check if it's a sitemap index
		if strings.Contains(string(content), "<sitemapindex") {
			var index sitemapIndex
			if err := xml.Unmarshal(content, &index); err != nil {
				return nil, fmt.Errorf("failed to parse sitemap index: %w", err)
			}

			for _, sitemap := range index.Sitemaps {
				if c.isProductSitemap(sitemap.Loc) {
					urls, err := c.parseSitemapURLs(sitemap.Loc)
					if err != nil {
						c.logger.Error("Failed to parse sitemap", "url", sitemap.Loc, "error", err)
						continue
					}
					productURLs = append(productURLs, urls...)
				}
			}
		} else {
			// Regular sitemap
			if c.isProductSitemap(sitemapURL) {
				urls, err := c.parseSitemapURLs(sitemapURL)
				if err != nil {
					return nil, fmt.Errorf("failed to parse sitemap: %w", err)
				}
				productURLs = append(productURLs, urls...)
			}
		}
	}

	return productURLs, nil
}

func (c *Crawler) isProductSitemap(url string) bool {
	return strings.Contains(strings.ToLower(url), "product") ||
		strings.Contains(strings.ToLower(url), "item") ||
		strings.Contains(strings.ToLower(url), "prod")
}

func (c *Crawler) parseSitemapURLs(sitemapURL string) ([]string, error) {
	resp, err := http.Get(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap content: %w", err)
	}

	var set urlset
	if err := xml.Unmarshal(content, &set); err != nil {
		return nil, fmt.Errorf("failed to parse sitemap URLs: %w", err)
	}

	var urls []string
	for _, u := range set.URLs {
		urls = append(urls, u.Loc)
	}

	return urls, nil
}