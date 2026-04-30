package cli

import (
	"testing"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/config"
)

func TestRun(t *testing.T) {
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
