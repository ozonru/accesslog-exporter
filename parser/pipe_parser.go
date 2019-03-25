package parser

import (
	"fmt"
	"strings"
)

// ParsePipedFormat parses log line according format using pipe as delimiter
func ParsePipedFormat(format, content string) (map[string]string, error) {
	variables := strings.Split(format, "|")
	values := strings.Split(content, "|")

	if len(variables) != len(values) {
		return nil, fmt.Errorf("format and content are inconsistent")
	}

	data := make(map[string]string)
	for k, variable := range variables {
		// clear variables
		variable = strings.TrimSpace(variable)
		variable = strings.Trim(variable, "\"")
		variable = strings.Trim(variable, "[")
		variable = strings.Trim(variable, "]")
		variable = strings.Trim(variable, "(")
		variable = strings.Trim(variable, ")")

		// clear values
		value := strings.TrimSpace(values[k])
		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "[")
		value = strings.Trim(value, "]")
		value = strings.Trim(value, "(")
		value = strings.Trim(value, ")")

		data[variable] = value

	}

	return data, nil
}
