package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

// TODO: pull out base_url, per_page
func FetchIndicators(page int, timeout time.Duration) ([]model.Indicator, model.PageMetadata, error) {
	url := fmt.Sprintf("https://api.worldbank.org/v2/indicator?format=json&per_page=1000&page=%d", page)
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
		return nil, model.PageMetadata{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
