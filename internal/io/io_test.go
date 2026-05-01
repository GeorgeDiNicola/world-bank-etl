package io

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

func TestWriteIndicatorsCSVReturnsWriterError(t *testing.T) {
	writer := csv.NewWriter(failingWriter{err: errors.New("write failed")})

	err := writeIndicatorsCSV(writer, []model.Indicator{{ID: "indicator-1", Name: "Indicator 1"}})
	if err == nil {
		t.Fatal("writeIndicatorsCSV() error = nil, want an error")
	}
}

func TestSaveIndicatorsToCSVWritesExpectedContent(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "indicators.csv")
	indicators := []model.Indicator{
		{
			ID:   "indicator-1",
			Name: "Indicator 1",
			Topics: []model.Topic{
				{ID: "1", Value: "Topic 1"},
				{ID: "2", Value: "Topic 2"},
			},
		},
	}

	err := SaveIndicatorsToCSV(filename, indicators)
	if err != nil {
		t.Fatalf("SaveIndicatorsToCSV() returned error: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("os.ReadFile() returned error: %v", err)
	}

	got := string(content)
	if !strings.Contains(got, "id,name,topics") {
		t.Fatalf("saved CSV = %q, want header row", got)
	}

	if !strings.Contains(got, "indicator-1,Indicator 1,\"Topic 1, Topic 2\"") {
		t.Fatalf("saved CSV = %q, want indicator row", got)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(p []byte) (int, error) {
	return 0, w.err
}
