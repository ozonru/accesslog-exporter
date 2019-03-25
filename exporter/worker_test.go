package exporter

import (
	"context"
	"regexp"
	"testing"

	"github.com/vlamug/accesslog-exporter/config"

	. "gopkg.in/check.v1"
)

func TestWorker(t *testing.T) { TestingT(t) }

type WorkerSuite struct{}

var _ = Suite(&WorkerSuite{})

func (s WorkerSuite) TestTryDetectCustomUserAgentLabels(c *C) {
	// Match
	userAgentReplacementSettings := config.UserAgentReplacementSetting{
		MatchRe:      regexp.MustCompile(`^myapp_android\/([0-9\.]+)`),
		Replacements: config.UserAgentReplacement{UserAgent: "myapp_android_$1", Device: "mobile", Os: "android"},
	}
	replacements := []config.UserAgentReplacementSetting{userAgentReplacementSettings}

	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&replacements,
		nil,
		nil,
	)

	uaLbs := w.tryDetectCustomUserAgentLabels(map[string]string{"$http_user_agent": "myapp_android/9.1"})
	c.Assert(uaLbs, NotNil)
	c.Assert(uaLbs, DeepEquals, &uaLabels{userAgent: "myapp_android_9.1", device: "mobile", os: "android"})

	uaLbs = w.tryDetectCustomUserAgentLabels(map[string]string{})
	c.Assert(uaLbs, IsNil)

	uaLbs = w.tryDetectCustomUserAgentLabels(map[string]string{"$http_user_agent": "myapp_ios/9.0"})
	c.Assert(uaLbs, IsNil)

	// Match
	userAgentReplacementSettings = config.UserAgentReplacementSetting{
		Match:        "myapp_android",
		Replacements: config.UserAgentReplacement{UserAgent: "myapp_android", Device: "mobile", Os: "android"},
	}
	replacements = []config.UserAgentReplacementSetting{userAgentReplacementSettings}

	w = NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&replacements,
		nil,
		nil,
	)

	uaLbs = w.tryDetectCustomUserAgentLabels(map[string]string{"$http_user_agent": "myapp_android"})
	c.Assert(uaLbs, NotNil)
	c.Assert(uaLbs, DeepEquals, &uaLabels{userAgent: "myapp_android", device: "mobile", os: "android"})
}

func (s WorkerSuite) TestDetectFormat(c *C) {
	logFormat := `$request_time "$host" $request $status $body_bytes_sent "$http_user_agent" $connection_requests`

	sources := []config.Source{{
		Host:      "localhost",
		LogFormat: logFormat,
	}}

	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		&sources,
		nil,
		nil,
		nil,
	)

	format := w.detectFormat(context.Background(), "localhost")
	c.Assert(format, Equals, logFormat)

	format = w.detectFormat(context.Background(), "somehost")
	c.Assert(format, Equals, defaultLogFormat)
}

func (s WorkerSuite) TestDetectUserAgentLabels(c *C) {
	w := NewExportWorker(
		nil,
		NewDummyUserAgentParser("ustest", "devicetest", "ostest"),
		&DummyCache{},
		newDummyExposeFunc(c),
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	uaLbs := w.detectUserAgentLabels(map[string]string{}, "localhost")
	c.Assert(uaLbs, DeepEquals, &uaLabels{userAgent: "unknown", os: "unknown", device: "unknown"})

	data := map[string]string{
		"$http_user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
	}
	uaLbs = w.detectUserAgentLabels(data, "localhost")
	c.Assert(uaLbs, DeepEquals, &uaLabels{userAgent: "ustest", os: "Ostest", device: "devicetest"})
}

func (s WorkerSuite) TestNeedParseUserAgent(c *C) {
	internalSubnets := []string{"30.0.0.0/8", "31.0.0.0/8", "32.0.0.0/8"}

	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		&internalSubnets,
		nil,
		nil,
		nil,
		nil,
	)

	c.Assert(w.needParseUserAgent(map[string]string{"$remote_addr": "30.2.2.1"}, context.Background()), Equals, false)
	c.Assert(w.needParseUserAgent(map[string]string{}, context.Background()), Equals, true)
	c.Assert(w.needParseUserAgent(map[string]string{"$remote_addr": "33.1.1.1"}, context.Background()), Equals, true)
}

func (s WorkerSuite) TestDetectHttpCodeLabel(c *C) {
	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	httpCode, err := w.detectHttpCodeLabel(map[string]string{"$status": "200"})
	c.Assert(err, IsNil)
	c.Assert(httpCode, Equals, "200")

	httpCode, err = w.detectHttpCodeLabel(map[string]string{"$status": "invalid"})
	c.Assert(err.Error(), Equals, `strconv.Atoi: parsing "invalid": invalid syntax`)
	c.Assert(httpCode, Equals, "unknown")

	httpCode, err = w.detectHttpCodeLabel(map[string]string{})
	c.Assert(err, IsNil)
	c.Assert(httpCode, Equals, "unknown")

	httpCode, err = w.detectHttpCodeLabel(map[string]string{"$status": "0"})
	c.Assert(err, IsNil)
	c.Assert(httpCode, Equals, "unknown")
}

func (s WorkerSuite) TestDetectURILabel(c *C) {
	replacements := []config.RequestURIReplacementSetting{{
		Method:       "POST",
		Regexp:       regexp.MustCompile(`^/search/.*`),
		Replacements: config.RequestURIReplacement{RequestURI: "search"},
	}}

	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&replacements,
		nil,
	)

	uri := w.detectURILabel(map[string]string{"$request": "POST /search/items?name=mobile"})
	c.Assert(uri, Equals, "search")

	uri = w.detectURILabel(map[string]string{})
	c.Assert(uri, Equals, "")

	uri = w.detectURILabel(map[string]string{"$request": "POST"})
	c.Assert(uri, Equals, "")

	uri = w.detectURILabel(map[string]string{"$request": "GET /search/items?name=mobile"})
	c.Assert(uri, Equals, "")

	uri = w.detectURILabel(map[string]string{"$request": "POST /category/product/2422"})
	c.Assert(uri, Equals, "")
}

func (s WorkerSuite) TestDetectHostLabel(c *C) {
	hosts := []config.Host{{
		Match:       "site.ru",
		Replacement: "www.site.ru",
	}}

	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&hosts,
	)

	host := w.detectHostLabel(map[string]string{"$host": "www.site.ru"})
	c.Assert(host, Equals, "www.site.ru")

	host = w.detectHostLabel(map[string]string{"$host": "site.ru"})
	c.Assert(host, Equals, "www.site.ru")

	host = w.detectHostLabel(map[string]string{})
	c.Assert(host, Equals, "unknown")
}

func (s WorkerSuite) TestDetectResponseDuration(c *C) {
	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	duration, exists, err := w.detectResponseDuration(map[string]string{"$request_time": "340"})
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, true)
	c.Assert(duration, Equals, float64(340))

	duration, exists, err = w.detectResponseDuration(map[string]string{})
	c.Assert(err, IsNil)
	c.Assert(exists, Equals, false)
	c.Assert(duration, Equals, float64(0))

	duration, exists, err = w.detectResponseDuration(map[string]string{"$request_time": "invalid"})
	c.Assert(err.Error(), Equals, `strconv.ParseFloat: parsing "invalid": invalid syntax`)
	c.Assert(exists, Equals, false)
	c.Assert(duration, Equals, float64(0))
}

func (s WorkerSuite) TestDetectDeviceType(c *C) {
	w := NewExportWorker(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	deviceType := w.detectDeviceType(
		map[string]string{},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "desktop")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla ipad bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla iphone bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla ipod bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"someOs",
		"generic tablet",
	)
	c.Assert(deviceType, Equals, "tablet")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"someOs",
		"generic smartphone",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"someOs",
		"generic feature phone",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"blackberry tablet os",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "tablet")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"blackberry os",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla windows touch bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"mobile",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla mobile bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "mobile")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"android",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "tablet")

	deviceType = w.detectDeviceType(
		map[string]string{"$http_user_agent": "bla bla bla"},
		"someUserAgent",
		"someOs",
		"someDevice",
	)
	c.Assert(deviceType, Equals, "desktop")
}
