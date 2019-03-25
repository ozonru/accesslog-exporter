package exporter

import (
	"context"

	"github.com/vlamug/accesslog-exporter/input"

	"github.com/ua-parser/uap-go/uaparser"
)

type DummyCache struct{}

func (c *DummyCache) Set(key string, value map[string]string) {}

func (c *DummyCache) Get(key string) (map[string]string, bool) {
	return nil, false
}

func (c *DummyCache) Len() int {
	return 0
}

type DummySyslogInput struct {
	dummyLogLine string
}

func NewDummySyslogInput(dummyLogLine string) *DummySyslogInput {
	return &DummySyslogInput{dummyLogLine: dummyLogLine}
}

func (p *DummySyslogInput) Run(lines chan<- *input.LogLine, ctx context.Context) {
	lines <- input.NewLogLine("localhost", p.dummyLogLine)
}

func NewDummyParseFunc(data map[string]string) func(format, content string) (map[string]string, error) {
	return func(format, content string) (strings map[string]string, e error) {
		return data, nil
	}
}

type DummyUserAgentParser struct {
	userAgent string
	device    string
	os        string
}

func NewDummyUserAgentParser(userAgent string, device string, os string) *DummyUserAgentParser {
	return &DummyUserAgentParser{userAgent: userAgent, device: device, os: os}
}

func (p *DummyUserAgentParser) Parse(line string) *uaparser.Client {
	client := &uaparser.Client{
		UserAgent: &uaparser.UserAgent{Family: p.userAgent},
		Device:    &uaparser.Device{Family: p.device},
		Os:        &uaparser.Os{Family: p.os},
	}

	return client
}
