package parser

import (
	"log"
	"testing"

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

// BenchmarkParseUaParser checks the parsing of user agent using ua-parser package
func BenchmarkParseUaParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		userAgentParser.Parse(userAgent)
	}
}

// BenchmarkParseSpaceParser parses log line using space parser
func BenchmarkParseSpaceParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseSpacedFormat(spacedLogFormat, spacedLogContent)
		if err != nil {
			log.Fatalf("could not run benchmark: %s", err)
		}
	}
}

// BenchmarkParsePipeParser parser log line using pipe parser
func BenchmarkParsePipeParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParsePipedFormat(pipedLogFormat, pipedLogContent)
		if err != nil {
			log.Fatalf("could not run benchmark: %s", err)
		}
	}
}
