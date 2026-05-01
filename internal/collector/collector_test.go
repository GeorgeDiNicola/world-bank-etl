package collector

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

// happy path
func TestGetAllIndicatorsReturnsAllPages(t *testing.T) {
	// mock a successful API call for 3 pages of indicators
	stubFetchIndicators(t, func(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
		return []model.Indicator{
			{ID: fmt.Sprintf("indicator-%d", page)},
		}, model.PageMetadata{Page: page, Pages: 3}, nil
	})

	collector := NewCollector(WithMaxRequestsPerSecond(100))

	indicators, err := collector.GetAllIndicators()
	if err != nil {
		t.Fatalf("GetAllIndicators() returned error: %v", err)
	}

	gotIDs := indicatorIDs(indicators)
	wantIDs := []string{"indicator-1", "indicator-2", "indicator-3"}

	// ensure same order since goroutines are non-deterministic
	slices.Sort(gotIDs)
	slices.Sort(wantIDs)

	// ensure all of the indicator IDs were collected
	if !slices.Equal(gotIDs, wantIDs) {
		t.Fatalf("GetAllIndicators() IDs = %v, want %v", gotIDs, wantIDs)
	}
}

func TestGetAllIndicatorsReturnsInitialFetchError(t *testing.T) {
	expectedErr := errors.New("first page failed")

	// when the collector asks for page 1, throw an error
	stubFetchIndicators(t, func(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
		if page == 1 {
			return nil, model.PageMetadata{}, expectedErr
		}

		t.Fatalf("unexpected page request: %d", page)
		return nil, model.PageMetadata{}, nil
	})

	collector := NewCollector()

	indicators, err := collector.GetAllIndicators()
	if err == nil {
		t.Fatal("GetAllIndicators() error = nil, want an error")
	}

	if indicators != nil {
		t.Fatalf("GetAllIndicators() indicators = %v, want nil", indicators)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetAllIndicators() error = %v, want wrapped %v", err, expectedErr)
	}
}

// Ensure proper handling of 1 of the requests failing
func TestGetAllIndicatorsReturnsWorkerError(t *testing.T) {
	expectedErr := errors.New("page fetch failed")

	// when the collector asks for page 2, throw an error. the other 2 pages are successful
	stubFetchIndicators(t, func(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
		switch page {
		case 1:
			return []model.Indicator{{ID: "indicator-1"}}, model.PageMetadata{Page: 1, Pages: 3}, nil
		case 2:
			return nil, model.PageMetadata{}, expectedErr
		case 3:
			return []model.Indicator{{ID: "indicator-3"}}, model.PageMetadata{Page: 3, Pages: 3}, nil
		default:
			t.Fatalf("unexpected page request: %d", page)
			return nil, model.PageMetadata{}, nil
		}
	})

	collector := NewCollector(WithMaxRequestsPerSecond(100))

	// ensure the retrieval is all or nothing. No indicators should be returned if any fail
	indicators, err := collector.GetAllIndicators()
	if err == nil {
		t.Fatal("GetAllIndicators() error = nil, want an error")
	}

	if indicators != nil {
		t.Fatalf("GetAllIndicators() indicators = %v, want nil", indicators)
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetAllIndicators() error = %v, want wrapped %v", err, expectedErr)
	}
}

func TestGetAllIndicatorsRateLimitsWorkerRequests(t *testing.T) {
	// needed to prevent a data race
	var mu sync.Mutex
	workerCallTimes := make([]time.Time, 0, 2)

	stubFetchIndicators(t, func(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
		// let the first page pass successfully so the collector can begin
		if page == 1 {
			return []model.Indicator{{ID: "indicator-1"}}, model.PageMetadata{Page: 1, Pages: 3}, nil
		}

		// record the time @ which every worker calls the API
		mu.Lock()
		workerCallTimes = append(workerCallTimes, time.Now())
		mu.Unlock()

		return []model.Indicator{{ID: fmt.Sprintf("indicator-%d", page)}}, model.PageMetadata{Page: page, Pages: 3}, nil
	})

	// set the speed to 2 requests per second to see if the workers respect it
	collector := NewCollector(WithMaxRequestsPerSecond(2))

	_, err := collector.GetAllIndicators()
	if err != nil {
		t.Fatalf("GetAllIndicators() returned error: %v", err)
	}

	if len(workerCallTimes) != 2 {
		t.Fatalf("worker requests = %d, want 2", len(workerCallTimes))
	}

	/*
	  get the gap of time between the worker calls to the API
	  since 2 reqs/sec, should be 500ms between, but ensure a buffer of 50ms
	  to avoid flaky tests when there is a few ms difference
	*/
	gotGap := workerCallTimes[1].Sub(workerCallTimes[0])
	minGap := 450 * time.Millisecond

	if gotGap < minGap {
		t.Fatalf("worker request gap = %v, want at least %v", gotGap, minGap)
	}
}

func stubFetchIndicators(
	t *testing.T,
	stub func(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error),
) {
	t.Helper()

	// swap the func variable with the stub
	originalFetchIndicators := fetchIndicators
	fetchIndicators = stub

	t.Cleanup(func() {
		fetchIndicators = originalFetchIndicators
	})
}

func indicatorIDs(indicators []model.Indicator) []string {
	ids := make([]string, 0, len(indicators))
	for _, indicator := range indicators {
		ids = append(ids, indicator.ID)
	}

	return ids
}
