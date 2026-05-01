package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchIndicatorsReturnsStatusCodeError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(http.StatusTooManyRequests, "rate limited"), nil
		}),
	}

	_, _, err := fetchIndicators(client, "https://example.com")
	if err == nil {
		t.Fatal("fetchIndicators() error = nil, want an error")
	}
}

func TestFetchIndicatorsReturnsResponseShapeError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(http.StatusOK, `[{"page":1,"pages":1,"total":1}]`), nil
		}),
	}

	_, _, err := fetchIndicators(client, "https://example.com")
	if err == nil {
		t.Fatal("fetchIndicators() error = nil, want an error")
	}
}

func TestFetchIndicatorsReturnsIndicatorsAndMetadata(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return newResponse(http.StatusOK, `[
			{"page":1,"pages":2,"total":2},
			[{"id":"indicator-1","name":"Indicator 1","topics":[{"id":"1","value":"Topic 1"}]}]
		]`), nil
		}),
	}

	indicators, metadata, err := fetchIndicators(client, "https://example.com")
	if err != nil {
		t.Fatalf("fetchIndicators() returned error: %v", err)
	}

	if metadata.Pages != 2 {
		t.Fatalf("metadata.Pages = %d, want 2", metadata.Pages)
	}

	if len(indicators) != 1 || indicators[0].ID != "indicator-1" {
		t.Fatalf("indicators = %v, want one parsed indicator", indicators)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
