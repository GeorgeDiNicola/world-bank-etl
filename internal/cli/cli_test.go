package cli

import (
	"testing"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/collector"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/config"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

type stubCollector struct {
	indicators []model.Indicator
	err        error
}

func (s stubCollector) GetAllIndicators() ([]model.Indicator, error) {
	return s.indicators, s.err
}

func TestRun(t *testing.T) {
	// store the real functions so they can be returned to their original state when tests are done
	originalNewIndicatorCollector := newIndicatorCollector
	originalSaveIndicators := saveIndicators

	// mock indicator collector
	newIndicatorCollector = func(opts ...collector.Option) indicatorCollector {
		return stubCollector{
			indicators: []model.Indicator{
				{ID: "indicator-1", Name: "Indicator 1"},
			},
		}
	}

	// a mock of a successful file save
	saveIndicators = func(filename string, indicators []model.Indicator) error {
		return nil
	}

	// restore original state
	t.Cleanup(func() {
		newIndicatorCollector = originalNewIndicatorCollector
		saveIndicators = originalSaveIndicators
	})

	tests := []struct {
		name          string
		cfg           *config.Config
		expectedError bool
	}{
		{
			name: "Valid input",
			cfg: &config.Config{
				Countries:  "USA, CAN",
				Indicators: "NY.GDP.PCAP.CD, SP.POP.TOTL",
			},
			expectedError: false,
		},
		{
			name: "Missing countries",
			cfg: &config.Config{
				Countries:  "",
				Indicators: "NY.GDP.PCAP.CD, SP.POP.TOTL",
			},
			expectedError: true,
		},
		{
			name: "Missing indicators",
			cfg: &config.Config{
				Countries:  "USA, CAN",
				Indicators: "",
			},
			expectedError: true,
		},
		{
			name: "Whitespace only input",
			cfg: &config.Config{
				Countries:  "  ",
				Indicators: " , ",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.cfg)

			if (err != nil) != tt.expectedError {
				t.Errorf("Run() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
