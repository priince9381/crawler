package main

import (
	"context"
	"crawler/crawler"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	outputFile := "crawler_result.json"
	domains := []string{"https://dentalstall.com/shop/"}

	crawlerWorker := crawler.NewCrawler()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	results := make(map[int]crawler.CrawResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for index, domain := range domains {
		wg.Add(1)
		go func(index int, domain string) {
			defer wg.Done()

			result := crawlerWorker.CrawlerLoad(ctx, domain)

			mu.Lock()
			results[index] = *result
			mu.Unlock()

			if result.Error != "" {
				log.Printf("Crawling error for %s: %s", domain, result.Error)
			}
		}(index, domain)
	}
	wg.Wait()

	// Write results to JSON file
	file, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshaling error: %v", err)
	}

	err = os.WriteFile(outputFile, file, 0644)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	totalProducts := 0
	for _, result := range results {
		totalProducts += len(result.Products)
	}

	fmt.Printf("Crawled %d products across %d domains. Results saved to %s\n",
		totalProducts, len(domains), outputFile)
}
