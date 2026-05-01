package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/api"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

var fetchIndicators = api.FetchIndicators

type Option func(*Collector)

func WithTimeout(timeoutSeconds time.Duration) Option {
	return func(c *Collector) {
		c.timeout = timeoutSeconds
	}
}

func WithMaxRequestsPerSecond(requests int) Option {
	return func(c *Collector) {
		c.maxRequestsPerSecond = requests
	}
}

type Collector struct {
	// API limits
	maxRequestsPerSecond int
	timeout              time.Duration
	maxConnections       int

	// Extraction params
	resultsPerPage   int
	countryBatchSize int

	// Retry logic
	timeUntilRetry    int
	maxRetries        int
	backoffMultiplier int
}

func NewCollector(opts ...Option) *Collector {
	c := &Collector{
		maxRequestsPerSecond: 10,
		timeout:              30,
		maxConnections:       10,
		resultsPerPage:       1000,
		countryBatchSize:     25,
		timeUntilRetry:       5,
		maxRetries:           2,
		backoffMultiplier:    1,
	}

	// override defaults
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Fan-out, fan-in pattern
func (c *Collector) GetAllIndicators() ([]model.Indicator, error) {
	// Get 1st page to get the page count metadata
	firstPage, metadata, err := fetchIndicators(1, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial page: %w", err)
	}

	totalNumPages := metadata.Pages
	var allIndicators []model.Indicator
	allIndicators = append(allIndicators, firstPage...)

	resultsChannel := make(chan []model.Indicator, totalNumPages)
	errorChannel := make(chan error, totalNumPages)

	var wg sync.WaitGroup
	var rateLimiter <-chan time.Time

	if c.maxRequestsPerSecond > 0 {
		ticker := time.NewTicker(time.Second / time.Duration(c.maxRequestsPerSecond))
		defer ticker.Stop()
		rateLimiter = ticker.C
	}

	// Fan-out
	for page := 2; page <= totalNumPages; page++ {
		wg.Add(1)
		// send page in so each worker gets their own page #
		go func(page int) {
			defer wg.Done()
			if rateLimiter != nil {
				<-rateLimiter
			}

			data, _, err := fetchIndicators(page, c.timeout)
			if err != nil {
				errorChannel <- fmt.Errorf("error on page %d: %w", page, err)
				return
			}
			resultsChannel <- data
		}(page)
	}

	go func() {
		wg.Wait()
		close(resultsChannel)
		close(errorChannel)
	}()

	// Fan-in the results
	for data := range resultsChannel {
		allIndicators = append(allIndicators, data...)
		fmt.Printf("Collected batch. Current total: %d\n", len(allIndicators))
	}

	for workerErr := range errorChannel {
		if workerErr != nil {
			return nil, workerErr
		}
	}

	return allIndicators, nil
}
