package exporter

import (
	"context"
	"testing"
	"time"

	"github.com/vlamug/accesslog-exporter/config"
	"github.com/vlamug/accesslog-exporter/exposer"

	. "gopkg.in/check.v1"
)

const (
	dummyLogLine   = `4.000 test_localhost "GET /test" 200 124 "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36" 6`
	dummyUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
)

func TestExporter(t *testing.T) { TestingT(t) }

type ExporterSuite struct{}

var _ = Suite(&ExporterSuite{})

func (s ExporterSuite) TestRun(c *C) {
	// define config
	cfg := &config.Config{
		Global: config.Global{
			InternalSubnets:               []string{},
			UserAgentCacheSize:            1000,
			ExportWorkers:                 1000,
			UserAgentReplacementSettings:  nil,
			RequestURIReplacementSettings: nil,
		},
		Sources: []config.Source{{
			Host:      "localhost",
			LogFormat: `$request_time "$host" $request $status $body_bytes_sent "$http_user_agent" $connection_requests`,
		}},
	}

	dummyParseFunc := NewDummyParseFunc(map[string]string{
		"$request_time":        "4.000",
		"$host":                "test_localhost",
		"$request":             "GET /test",
		"$status":              "200",
		"$body_bytes_sent":     "124",
		"$http_user_agent":     dummyUserAgent,
		"$connection_requests": "6",
	})

	// create new exporter and run
	srv, err := NewExporter(
		cfg,
		NewDummySyslogInput(dummyLogLine),
		dummyParseFunc,
		NewDummyUserAgentParser("ustest", "devicetest", "ostest"),
		&DummyCache{},
		newDummyExposeFunc(c),
	)
	c.Assert(err, IsNil)

	ctx, cancel := context.WithCancel(context.Background())
	srv.Run(ctx)

	// @todo The dirty hack. It is needed to extract the exposing metrics from worker, because worker is run in goroutine
	time.Sleep(time.Second)

	cancel()
}

// newDummyExposeFunc creates new dummy metric exposer
func newDummyExposeFunc(c *C) func(name string, labels []string, value float64) {
	return func(name string, labels []string, value float64) {
		switch name {
		case exposer.UserAgentCachedTotal:
			c.Assert(value, Equals, float64(0))
			c.Assert(labels, DeepEquals, []string{"localhost"})
		case exposer.UserAgentCurrentCachedTotal:
			c.Assert(value, Equals, float64(0))
			c.Assert(labels, DeepEquals, []string{"localhost"})
		case exposer.HostResponseTimeSecondsMetricName:
			c.Assert(value, Equals, float64(4))
			c.Assert(labels, DeepEquals, []string{"test_localhost", "200"})
		case exposer.UserAgentResponseTimeSecondsMetricName:
			c.Assert(value, Equals, float64(4))
			c.Assert(labels, DeepEquals, []string{"test_localhost", "ustest", "200"})
		case exposer.URIResponseTimeSecondsMetricName:
			c.Assert(value, Equals, float64(4))
			c.Assert(labels, DeepEquals, []string{"test_localhost", "", "200"})
		case exposer.UserAgentRequestsTotalMetricName:
			c.Assert(value, Equals, float64(0))
			c.Assert(labels, DeepEquals, []string{"test_localhost", "ustest", "200"})
		case exposer.OsDeviceTypeRequestsTotalMetricName:
			c.Assert(value, Equals, float64(0))
			c.Assert(labels, DeepEquals, []string{"test_localhost", "Ostest", "desktop"})
		case exposer.NginxRequestsTotal:
			c.Assert(value, Equals, float64(0))
			c.Assert(labels, DeepEquals, []string{"localhost"})
		}
	}
}
