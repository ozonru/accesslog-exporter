package benchmark

import (
	"log"
	"testing"

	"github.com/ozonru/accesslog-exporter/parser"

	"github.com/ua-parser/uap-go/uaparser"
)

var (
	userAgent        = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
	spacedLogFormat  = `$remote_addr - $remote_user [$time_local] "$request" $request_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for" $connection_requests`
	spacedLogContent = `127.0.0.1 - - [20/Sep/2018:19:37:35 +0400] "GET / HTTP/1.1" 0.000 304 0 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36" "-" 14`
	pipedLogFormat   = `$remote_addr | - | $remote_user | [$time_local] | "$request" | $request_time | $status | $body_bytes_sent | "$http_referer" | "$http_user_agent" | "$http_x_forwarded_for" | $connection_requests`
	pipedLogContent  = `127.0.0.1 | - | - | [20/Sep/2018:19:37:35 +0400] | "GET / HTTP/1.1" | 0.000 | 304 | 0 | "-" | "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36" | "-" | 14`

	userAgentParser = uaparser.NewFromSaved()
)

// BenchmarkUaParserParse checks the parsing of user agent using ua-parser package
func BenchmarkUaParserParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		userAgentParser.Parse(userAgent)
	}
}

// BenchmarkSpaceParser parses log line using space parser
func BenchmarkSpaceParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseSpacedFormat(spacedLogFormat, spacedLogContent)
		if err != nil {
			log.Fatalf("could not run benchmark: %s", err)
		}
	}
}

// BenchmarkPipeParser parser log line using pipe parser
func BenchmarkPipeParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parser.ParsePipedFormat(pipedLogFormat, pipedLogContent)
		if err != nil {
			log.Fatalf("could not run benchmark: %s", err)
		}
	}
}
