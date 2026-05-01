package io

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/GeorgeDiNicola/world-bank-etl/internal/model"
)

func SaveIndicatorsToCSV(filename string, indicators []model.Indicator) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writeIndicatorsCSV(writer, indicators); err != nil {
		return err
	}

	fmt.Printf("Successfully saved %d indicators to %s\n", len(indicators), filename)
	return nil
}

func writeIndicatorsCSV(writer *csv.Writer, indicators []model.Indicator) error {
	if err := writer.Write([]string{"id", "name", "topics"}); err != nil {
		return err
	}

	for _, ind := range indicators {
		var topicNames []string
		for _, t := range ind.Topics {
			if t.Value != "" {
				topicNames = append(topicNames, t.Value)
			}
		}

		if err := writer.Write([]string{
			ind.ID,
			ind.Name,
			strings.Join(topicNames, ", "),
		}); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
