package cli

import (
	"fmt"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/collector"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/config"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/io"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/utils"
)

type indicatorCollector interface {
	GetAllIndicators() ([]model.Indicator, error)
}

var newIndicatorCollector = func(opts ...collector.Option) indicatorCollector {
	return collector.NewCollector(opts...)
}

var saveIndicators = io.SaveIndicatorsToCSV

func Run(cfg *config.Config) error {
	countries := utils.ParseCSVStringIntoSlice(cfg.Countries)
	indicators := utils.ParseCSVStringIntoSlice(cfg.Indicators)

	if len(countries) == 0 || len(indicators) == 0 {
		return fmt.Errorf("must provide at least 1 country and 1 indicator")
	}

	indicatorCollector := newIndicatorCollector(
		collector.WithTimeout(30),
		collector.WithMaxRequestsPerSecond(10),
	)

	allIndicators, err := indicatorCollector.GetAllIndicators()
	if err != nil {
		return fmt.Errorf("Failed to collect all indicators with error: %v\n", err)
	}

	return saveIndicators("indicators.csv", allIndicators)
}
