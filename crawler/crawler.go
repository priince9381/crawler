package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Product struct {
	URL string `json:"url"`
}

type CrawResult struct {
	Domain   string
	Products []Product
	Error    string `json:"error,omitempty"`
}

func (c *Crawler) CrawlerLoad(ctx context.Context, baseURL string) *CrawResult {
	seenURLs := make(map[string]bool)
	var products []Product

	result := &CrawResult{
		Domain: baseURL,
	}

	// Use a channel to control concurrency and rate limiting
	sem := make(chan struct{}, 10) // 10 concurrent goroutines
	var wg sync.WaitGroup

	var crawlRecursive func(string, int) error
	crawlRecursive = func(currentURL string, depth int) error {
		// Check depth and context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if depth > c.config.MaxDepth {
				return nil
			}
		}

		// Rate limiting
		time.Sleep(c.config.RateLimit)

		// Prevent duplicate crawling
		c.mu.Lock()
		if seenURLs[currentURL] {
			c.mu.Unlock()
			return nil
		}
		seenURLs[currentURL] = true
		c.mu.Unlock()

		// Prepare request
		req, err := http.NewRequestWithContext(ctx, "GET", currentURL, nil)
		if err != nil {
			log.Printf("Error creating request for %s: %v", currentURL, err)
			return err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ProductCrawler/2.0)")

		// Acquire semaphore
		sem <- struct{}{}
		defer func() { <-sem }()
		fmt.Printf("Current Url", currentURL)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("Error fetching %s: %v", currentURL, err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Non-200 status for %s: %d", currentURL, resp.StatusCode)
			return fmt.Errorf("HTTP status %d", resp.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("Error parsing HTML for %s: %v", currentURL, err)
			return err
		}

		if isProductURL(currentURL, c.config.ProductURLPatterns) {
			c.mu.Lock()
			products = append(products, Product{URL: currentURL})
			c.mu.Unlock()
		}

		if depth < c.config.MaxDepth {
			doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if exists {
					fullURL := resolveURL(currentURL, href)

					wg.Add(1)
					go func(url string) {
						defer wg.Done()
						crawlRecursive(url, depth+1)
					}(fullURL)
				}
			})
		}

		return nil
	}

	err := crawlRecursive(baseURL, 0)
	if err != nil {
		result.Error = err.Error()
	}

	wg.Wait()
	result.Products = products
	return result
}

func isProductURL(href string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(href, pattern) {
			return true
		}
	}
	return false
}

func resolveURL(base, href string) string {
	parsedBase, err := url.Parse(base)
	if err != nil {
		return href // Return as is if parsing fails
	}

	parsedHref, err := url.Parse(href)
	if err != nil {
		return href
	}

	return parsedBase.ResolveReference(parsedHref).String()
}
