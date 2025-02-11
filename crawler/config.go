package crawler

import (
	"net/http"
	"sync"
	"time"
)

type CrawlerConfig struct {
	ProductURLPatterns []string `json:"productUrlPatterns"`

	MaxDepth int `json:"maxDepth"`

	RateLimit time.Duration `json:"rateLimit"`

	MaxConcurrentWorkers int `json:"maxConcurrentWorkers"`

	RequestTimeout time.Duration `json:"requestTimeout"`
}

// DefaultCrawlerConfig provides a sensible default configuration
func DefaultCrawlerConfig() CrawlerConfig {
	return CrawlerConfig{
		ProductURLPatterns:   []string{"/product/", "/products/", "/item/", "/p/"},
		MaxDepth:             1,
		RateLimit:            500 * time.Millisecond,
		MaxConcurrentWorkers: 10,
		RequestTimeout:       30 * time.Second,
	}
}

type Crawler struct {
	config     CrawlerConfig
	httpClient *http.Client
	mu         sync.Mutex
}

func NewCrawler() *Crawler {
	return &Crawler{
		config: DefaultCrawlerConfig(),
		httpClient: &http.Client{
			Timeout: DefaultCrawlerConfig().RequestTimeout,
		},
	}
}
