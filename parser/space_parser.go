package parser

import (
	"fmt"
	"strings"
)

// innerDelimiter is a delimiter that is used to replace spaces in the values of variables in order to parse log lines
// correctly according to the format. After parsing, this delimiter is used to return spaces in values. For example
// space in time [19/Sep/2018:19:52:01 +0400] will look like [19/Sep/2018:19:52:01___+0400].
const innerDelimiter = "___"

// ParseSpacedFormat parses log line according format using space as delimiter
func ParseSpacedFormat(format, content string) (map[string]string, error) {
	variables := strings.Split(format, " ")
	values := strings.Split(replaceInnerSpaces(content), " ")

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

		// replace inner delimiter with spaces
		value = strings.Replace(value, innerDelimiter, " ", -1)

		data[variable] = value
	}

	return data, nil
}

// replaceInnerSpaces replaces inner spaces(spaces in value of variables), like [19/Sep/2018:19:52:01 +0400] with
// delimiter innerDelimiter
func replaceInnerSpaces(content string) string {
	frame := 0
	modifiedContent := ""
	for _, v := range content {
		char := string(v)
		if char == "\"" {
			if frame != 0 {
				frame--
			} else {
				frame++
			}
		}

		if char == "[" || char == "(" {
			frame++
		}

		if char == "]" || char == ")" {
			frame--
		}

		if char == " " && frame > 0 {
			char = innerDelimiter
		}

		modifiedContent += char
	}

	return modifiedContent
}
