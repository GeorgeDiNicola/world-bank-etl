package main

import (
	"fmt"
	"os"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/cli"
	"github.com/GeorgeDiNicola/world-bank-etl/internal/config"
)

func main() {
	cfg := config.Load()

	if err := cli.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
