package cli

import (
	"fmt"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/config"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/utils"
)

func Run(cfg *config.Config) error {
	countries := utils.ParseCSVStringIntoSlice(cfg.Countries)
	indicators := utils.ParseCSVStringIntoSlice(cfg.Indicators)

	if len(countries) == 0 || len(indicators) == 0 {
		return fmt.Errorf("must provide at least 1 country and 1 indicator")
	}

	fmt.Printf("Processing %d countries and %d indicators...\n", len(countries), len(indicators))

	for _, c := range countries {
		for _, i := range indicators {
			fmt.Printf("Requesting %s for %s\n", i, c)
		}
	}

	return nil
}
