package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

const worldBankAPIBaseURL = "https://api.worldbank.org/v2"

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
	}

	return fmt.Sprintf("unexpected status code: %d: %s", e.StatusCode, e.Body)
}

// TODO: allow for batching of indicators when necessary
func FetchIndicatorObservations(page int, indicator, countries, sourceID string, perPage int, timeout time.Duration) ([]model.Observation, model.PageMetadata, error) {
	url := fmt.Sprintf("https://api.worldbank.org/v2/country/%s/indicator/%s?page=%d&format=json&per_page=%d", countries, indicator, page, perPage)
	if sourceID != "" {
		url = fmt.Sprintf("%s&source=%s", url, sourceID)
	}

	fmt.Println(url)
	client := &http.Client{Timeout: timeout * time.Second}
	return fetchObservations(client, url)
}

func fetchObservations(client *http.Client, url string) ([]model.Observation, model.PageMetadata, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, model.PageMetadata{}, fmt.Errorf("failed to fetch observations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, model.PageMetadata{}, newHTTPStatusError(resp)
	}

	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, model.PageMetadata{}, err
	}

	if len(raw) < 2 {
		return nil, model.PageMetadata{}, fmt.Errorf("unexpected response shape: expected metadata and indicators")
	}

	// Index 0 is Metadata
	var meta model.PageMetadata
	if err := json.Unmarshal(raw[0], &meta); err != nil {
		return nil, model.PageMetadata{}, err
	}

	// Index 1 is the slice of Indicators
	var observations []model.Observation
	if err := json.Unmarshal(raw[1], &observations); err != nil {
		return nil, model.PageMetadata{}, err
	}

	return observations, meta, nil

}

func FetchIndicators(page, perPage int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
	url := fmt.Sprintf("%s/indicator?format=json&per_page=%d&page=%d", worldBankAPIBaseURL, perPage, page)
	client := &http.Client{Timeout: timeout * time.Second}
	return fetchIndicators(client, url)
}

func fetchIndicators(client *http.Client, url string) ([]model.Indicator, model.PageMetadata, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, model.PageMetadata{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, model.PageMetadata{}, newHTTPStatusError(resp)
	}

	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, model.PageMetadata{}, err
	}

	if len(raw) < 2 {
		return nil, model.PageMetadata{}, fmt.Errorf("unexpected response shape: expected metadata and indicators")
	}

	// Index 0 is Metadata
	var meta model.PageMetadata
	if err := json.Unmarshal(raw[0], &meta); err != nil {
		return nil, model.PageMetadata{}, err
	}

	// Index 1 is the slice of Indicators
	var indicators []model.Indicator
	if err := json.Unmarshal(raw[1], &indicators); err != nil {
		return nil, model.PageMetadata{}, err
	}

	return indicators, meta, nil
}

func newHTTPStatusError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPStatusError{StatusCode: resp.StatusCode}
	}

	return &HTTPStatusError{
		StatusCode: resp.StatusCode,
		Body:       strings.Join(strings.Fields(string(body)), " "),
	}
}
