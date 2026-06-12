package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

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

type observationCollector interface {
	GetAllObservations(indicators []model.Indicator, countries []string) ([]model.Observation, error)
}

var newObservationCollector = func(opts ...collector.Option) observationCollector {
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

	if err := saveIndicators("indicators.csv", allIndicators); err != nil {
		return fmt.Errorf("failed to save indicators: %w", err)
	}

	requestedIndicators, err := selectRequestedIndicators(indicators, allIndicators)
	if err != nil {
		return err
	}

	observationCollector := newObservationCollector(
		collector.WithTimeout(60),
		collector.WithMaxRequestsPerSecond(10),
	)

	result, err := observationCollector.GetAllObservations(requestedIndicators, countries)
	if err != nil {
		return fmt.Errorf("failed to fetch observations: %w", err)
	}

	if err := saveObservationsToCSV("observations.csv", result); err != nil {
		return fmt.Errorf("failed to save observations: %w", err)
	}

	return nil

}

func selectRequestedIndicators(requestedIDs []string, allIndicators []model.Indicator) ([]model.Indicator, error) {
	if containsAllIndicator(requestedIDs) {
		return allIndicators, nil
	}

	indicatorsByID := make(map[string]model.Indicator, len(allIndicators))
	for _, indicator := range allIndicators {
		indicatorsByID[indicator.ID] = indicator
	}

	requestedIndicators := make([]model.Indicator, 0, len(requestedIDs))
	missingIndicatorIDs := make([]string, 0)

	for _, requestedID := range requestedIDs {
		indicator, found := indicatorsByID[requestedID]
		if !found {
			missingIndicatorIDs = append(missingIndicatorIDs, requestedID)
			continue
		}
		requestedIndicators = append(requestedIndicators, indicator)
	}

	if len(missingIndicatorIDs) > 0 {
		return nil, fmt.Errorf("requested indicators not found: %s", strings.Join(missingIndicatorIDs, ", "))
	}

	return requestedIndicators, nil
}

func containsAllIndicator(requestedIDs []string) bool {
	for _, requestedID := range requestedIDs {
		if requestedID == "all" {
			return true
		}
	}

	return false
}

func saveObservationsToCSV(filename string, observations []model.Observation) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"indicator_id",
		"indicator",
		"country_id",
		"country",
		"country_iso3_code",
		"date",
		"value",
		"unit",
		"obs_status",
		"decimal",
	}); err != nil {
		return err
	}

	for _, observation := range observations {
		value := ""
		if observation.Value != nil {
			value = strconv.FormatFloat(*observation.Value, 'f', -1, 64)
		}

		if err := writer.Write([]string{
			observation.Indicator.ID,
			observation.Indicator.Value,
			observation.Country.ID,
			observation.Country.Value,
			observation.CountryISO3,
			observation.Date,
			value,
			observation.Unit,
			observation.ObsStatus,
			strconv.Itoa(observation.Decimal),
		}); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	fmt.Printf("Successfully saved %d observations to %s\n", len(observations), filename)
	return nil
}
