package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"ecommerce-crawler/internal/utils"
)

type HTTPClient struct {
	client    *http.Client
	userAgent string
	logger    *utils.Logger
}

func NewHTTPClient(logger *utils.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 20,
			},
		},
		userAgent: "EcommerceCrawler/1.0",
		logger:    logger,
	}
}

func (h *HTTPClient) Fetch(urlStr string) (string, error) {
	// First try HEAD request to check if we should proceed
	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HEAD request: %w", err)
	}
	req.Header.Set("User-Agent", h.userAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HEAD request failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	// Now do the actual GET request
	req, err = http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %w", err)
	}
	req.Header.Set("User-Agent", h.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Exponential backoff for retries
	var body string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = h.client.Do(req)
		if err != nil {
			if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
				// Retry on timeout
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return "", fmt.Errorf("GET request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
		}

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		body = string(content)
		break
	}

	if body == "" {
		return "", fmt.Errorf("failed after %d retries", maxRetries)
	}

	return body, nil
}

func (h *HTTPClient) FetchWithContext(ctx context.Context, urlStr string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
    if err != nil {
        return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
    }

    resp, err := h.client.Do(req)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return "", fmt.Errorf("%w: %v", ErrTimeout, err)
        }
        return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
    }
    defer resp.Body.Close()


    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
    }

	req, err = http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", h.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err = h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}
