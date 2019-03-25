package input

// LogLine is struct that contains hostname, content of log line and type of input.
type LogLine struct {
	NginxHost string
	Content   string
}

// NewLogLine creates new struct, that contains nginx host, content of log line and type of input.
func NewLogLine(nginxName, content string) *LogLine {
	return &LogLine{NginxHost: nginxName, Content: content}
}
