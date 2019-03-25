package parser

// LogLineParser is an interface for service that parses log lines
type LogLineParser func(format, content string) (map[string]string, error)
