package cmd

import "strings"

// convert comma separated list into an array, trimming spaces and ignoring empty values
func commaSeparatedListToArray(value string) (result []string) {
	for _, val := range strings.Split(value, ",") {
		val = strings.TrimSpace(val)
		if val != "" {
			result = append(result, val)
		}
	}
	return result
}
