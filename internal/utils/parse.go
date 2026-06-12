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

// Generic helper to split a slice into chunks of a given size
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
