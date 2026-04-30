package utils

import "strings"

func ParseCSVStringIntoSlice(input string) []string {
	if input == "" {
		return []string{}
	}

	raw := strings.Split(input, ",")
	var cleaned []string

	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return cleaned
}
