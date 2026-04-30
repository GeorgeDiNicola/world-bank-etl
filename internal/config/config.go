package config

import (
	"flag"
	"os"
)

type Config struct {
	Countries  string
	Indicators string
}

func Load() *Config {
	cfg := &Config{}

	// default values
	envCountries := os.Getenv("COUNTRIES")
	envIndicators := os.Getenv("INDICATORS")

	flag.StringVar(&cfg.Countries, "countries", envCountries, "Comma-separated country codes (env: COUNTRIES)")
	flag.StringVar(&cfg.Indicators, "indicators", envIndicators, "Comma-separated indicator codes (env: INDICATORS)")

	flag.Parse()

	return cfg
}
