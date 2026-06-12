package collector

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/api"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/utils"
)

var fetchIndicators = api.FetchIndicators
var fetchObservations = api.FetchIndicatorObservations

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

func WithResultsPerPage(resultsPerPage int) Option {
	return func(c *Collector) {
		c.resultsPerPage = resultsPerPage
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

const indicatorBatchSize = 40

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
func (c *Collector) GetAllObservations(indicators []model.Indicator, countries []string) ([]model.Observation, error) {
	var allObservations []model.Observation

	// prepare country batches
	var countryBatches [][]string
	if usesAllCountry(countries) {
		countryBatches = [][]string{{"all"}}
	} else {
		countryBatches = utils.ChunkSlice(countries, c.countryBatchSize)
	}

	indicatorBatches := batchIndicatorsBySource(indicators, indicatorBatchSize)

	var rateLimiter <-chan time.Time

	if c.maxRequestsPerSecond > 0 {
		ticker := time.NewTicker(time.Second / time.Duration(c.maxRequestsPerSecond))
		defer ticker.Stop()
		rateLimiter = ticker.C
	}

	// Iterate through batches
	for _, cBatch := range countryBatches {
		countriesString := strings.Join(cBatch, ";")

		for _, iBatch := range indicatorBatches {
			indicatorIDs := make([]string, 0, len(iBatch))
			for _, indicator := range iBatch {
				indicatorIDs = append(indicatorIDs, indicator.ID)
			}

			indicatorsString := strings.Join(indicatorIDs, ";")
			sourceID := sourceIDForBatch(iBatch)

			// Fetch pages for this specific batch combination
			batchResults, err := c.fetchFullDataset(indicatorsString, countriesString, sourceID, rateLimiter)
			if err != nil {
				return nil, err
			}
			allObservations = append(allObservations, batchResults...)
		}
	}

	return allObservations, nil
}

func usesAllCountry(countries []string) bool {
	for _, country := range countries {
		if country == "all" {
			return true
		}
	}

	return false
}

func batchIndicatorsBySource(indicators []model.Indicator, batchSize int) [][]model.Indicator {
	if len(indicators) == 0 {
		return nil
	}

	indicatorsBySourceID := make(map[string][]model.Indicator)
	sourceOrder := make([]string, 0)

	for _, indicator := range indicators {
		sourceID := indicator.Source.ID
		if _, seen := indicatorsBySourceID[sourceID]; !seen {
			sourceOrder = append(sourceOrder, sourceID)
		}
		indicatorsBySourceID[sourceID] = append(indicatorsBySourceID[sourceID], indicator)
	}

	batches := make([][]model.Indicator, 0)
	for _, sourceID := range sourceOrder {
		sourceIndicators := indicatorsBySourceID[sourceID]
		batches = append(batches, utils.ChunkSlice(sourceIndicators, batchSize)...)
	}

	return batches
}

// Helper to handle the pagination for a single batch combination
func (c *Collector) fetchFullDataset(indicatorString, countriesString, sourceID string, rateLimiter <-chan time.Time) ([]model.Observation, error) {
	firstPage, metadata, err := c.fetchObservationsPage(1, indicatorString, countriesString, sourceID, rateLimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch initial page: %w", err)
	}

	observations := firstPage
	for page := 2; page <= metadata.Pages; page++ {
		nextPage, _, err := c.fetchObservationsPage(page, indicatorString, countriesString, sourceID, rateLimiter)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		observations = append(observations, nextPage...)
	}
	return observations, nil
}

func sourceIDForBatch(indicators []model.Indicator) string {
	if len(indicators) == 0 {
		return ""
	}

	return indicators[0].Source.ID
}

func (c *Collector) fetchObservationsPage(page int, indicatorString, countriesString, sourceID string, rateLimiter <-chan time.Time) ([]model.Observation, model.PageMetadata, error) {
	retryDelay := time.Duration(c.timeUntilRetry) * time.Second
	attempts := c.maxRetries + 1
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		if rateLimiter != nil {
			<-rateLimiter
		}

		observations, metadata, err := fetchObservations(page, indicatorString, countriesString, sourceID, c.resultsPerPage, c.timeout)
		if err == nil {
			return observations, metadata, nil
		}

		lastErr = err
		if attempt == attempts || !shouldRetryObservationError(err) {
			break
		}

		time.Sleep(retryDelay)
		retryDelay *= time.Duration(c.backoffMultiplier + 1)
	}

	return nil, model.PageMetadata{}, lastErr
}

func shouldRetryObservationError(err error) bool {
	var statusErr *api.HTTPStatusError
	if errors.As(err, &statusErr) {
		if statusErr.StatusCode == http.StatusBadRequest ||
			statusErr.StatusCode == http.StatusTooManyRequests ||
			statusErr.StatusCode >= http.StatusInternalServerError {
			return true
		}
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

// Fan-out, fan-in pattern
func (c *Collector) GetAllIndicators() ([]model.Indicator, error) {
	// Get 1st page to get the page count metadata
	firstPage, metadata, err := fetchIndicators(1, c.resultsPerPage, c.timeout)
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

			data, _, err := fetchIndicators(page, c.resultsPerPage, c.timeout)
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
