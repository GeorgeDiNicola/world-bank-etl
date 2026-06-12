package api

import (
	"io"
	"net/http"
	"net/url"
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
			[{"id":"indicator-1","name":"Indicator 1","source":{"id":"37","value":"LAC Equity Lab"},"sourceOrganization":"LAC Equity Lab tabulations of SEDLAC (CEDLAS and the World Bank).","topics":[{"id":"1","value":"Topic 1"}]}]
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

	if indicators[0].Source.ID != "37" {
		t.Fatalf("indicators[0].Source.ID = %q, want %q", indicators[0].Source.ID, "37")
	}

	if indicators[0].Source.Value != "LAC Equity Lab" {
		t.Fatalf("indicators[0].Source.Value = %q, want %q", indicators[0].Source.Value, "LAC Equity Lab")
	}

	if indicators[0].SourceOrganization != "LAC Equity Lab tabulations of SEDLAC (CEDLAS and the World Bank)." {
		t.Fatalf("indicators[0].SourceOrganization = %q, want source organization to be parsed", indicators[0].SourceOrganization)
	}
}

func TestFetchIndicatorObservationsIncludesSourceID(t *testing.T) {
	var requestedURL *url.URL

	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requestedURL = req.URL
			return newResponse(http.StatusOK, `[
				{"page":1,"pages":1,"total":0},
				[]
			]`), nil
		}),
	}

	_, _, err := fetchObservations(client, "https://example.com/v2/country/all/indicator/test?page=1&format=json&per_page=1000&source=37")
	if err != nil {
		t.Fatalf("fetchObservations() returned error: %v", err)
	}

	if requestedURL == nil {
		t.Fatal("fetchObservations() did not issue a request")
	}

	if requestedURL.Query().Get("source") != "37" {
		t.Fatalf("source query param = %q, want %q", requestedURL.Query().Get("source"), "37")
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
