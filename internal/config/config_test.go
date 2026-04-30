package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoadUsesEnvironmentDefaults(t *testing.T) {
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	})

	t.Setenv("COUNTRIES", "USA,CAN")
	t.Setenv("INDICATORS", "NY.GDP.PCAP.CD,SP.POP.TOTL")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}

	cfg := Load()

	if cfg.Countries != "USA,CAN" {
		t.Fatalf("Load().Countries = %q, expected %q", cfg.Countries, "USA,CAN")
	}

	if cfg.Indicators != "NY.GDP.PCAP.CD,SP.POP.TOTL" {
		t.Fatalf("Load().Indicators = %q, expected %q", cfg.Indicators, "NY.GDP.PCAP.CD,SP.POP.TOTL")
	}
}

func TestLoadAllowsFlagsToOverrideEnvironment(t *testing.T) {
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine

	// restore the global state after test finishes
	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	})

	t.Setenv("COUNTRIES", "USA,CAN")
	t.Setenv("INDICATORS", "NY.GDP.PCAP.CD")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{
		"cmd",
		"-countries=USA,CAN",
		"-indicators=SP.POP.TOTL",
	}

	cfg := Load()

	if cfg.Countries != "USA,CAN" {
		t.Fatalf("Load().Countries = %q, expected %q", cfg.Countries, "USA,CAN")
	}

	if cfg.Indicators != "SP.POP.TOTL" {
		t.Fatalf("Load().Indicators = %q, expected %q", cfg.Indicators, "SP.POP.TOTL")
	}
}
