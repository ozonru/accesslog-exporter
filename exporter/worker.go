package exporter

import (
	"context"
	"strconv"
	"strings"

	"github.com/vlamug/accesslog-exporter/cache"
	"github.com/vlamug/accesslog-exporter/config"
	"github.com/vlamug/accesslog-exporter/exposer"
	"github.com/vlamug/accesslog-exporter/input"
	"github.com/vlamug/accesslog-exporter/parser"
	"github.com/vlamug/accesslog-exporter/pkg/logging"
	"github.com/vlamug/accesslog-exporter/pkg/net"
)

const (
	// defaultLogFormat is a default access log format
	defaultLogFormat = `$remote_addr - $remote_user [$time_local] "$request" $request_time $status $body_bytes_sent "$http_referer" "$http_user_agent" "$http_x_forwarded_for" $connection_requests`

	remoteAddrVar    = "$remote_addr"
	statusVar        = "$status"
	httpUserAgentVar = "$http_user_agent"
	requestTimeVar   = "$request_time"
	requestVar       = "$request"
	hostVar          = "$host"

	unknownLabelValue  = "unknown"
	internalLabelValue = "internal"

	userAgentLabelName = "user_agent"
	osLabelName        = "os"
	deviceLabelName    = "device"

	deviceTypeMobile  = "mobile"
	deviceTypeTablet  = "tablet"
	deviceTypeDesktop = "desktop"
)

// pool contains set of workers
type pool chan IWorker

// IWorker is an interface for worker
type IWorker interface {
	Process(line *input.LogLine, ctx context.Context)
}

type ExportWorker struct {
	logLinePsr   parser.LogLineParser
	userAgentPsr parser.UserAgentParser

	cc cache.Cache

	exposeFunc exposer.Exposer

	internalSubNets *[]string
	sources         *[]config.Source
	hosts           *[]config.Host

	userAgentReplacementSettings  *[]config.UserAgentReplacementSetting
	requestURIReplacementSettings *[]config.RequestURIReplacementSetting
}

func NewExportWorker(
	logLinePsr parser.LogLineParser,
	userAgentPsr parser.UserAgentParser,
	cc cache.Cache,
	exposeFunc exposer.Exposer,
	internalSubNets *[]string,
	source *[]config.Source,
	userAgentReplacementSettings *[]config.UserAgentReplacementSetting,
	requestURIReplacements *[]config.RequestURIReplacementSetting,
	hosts *[]config.Host,
) *ExportWorker {
	return &ExportWorker{
		logLinePsr:                    logLinePsr,
		userAgentPsr:                  userAgentPsr,
		cc:                            cc,
		exposeFunc:                    exposeFunc,
		internalSubNets:               internalSubNets,
		sources:                       source,
		userAgentReplacementSettings:  userAgentReplacementSettings,
		requestURIReplacementSettings: requestURIReplacements,
		hosts:                         hosts,
	}
}

// Process processes log lines and exports metrics
func (e *ExportWorker) Process(line *input.LogLine, ctx context.Context) {
	format := e.detectFormat(ctx, line.NginxHost)

	data, err := e.logLinePsr(format, line.Content)
	if err != nil {
		logging.WithContext(ctx).Sugar().With(
			"format", format,
			"content", line.Content,
		).Warnf("could not parse log line: %s", err)

		e.exposeFunc(exposer.LogsFailParsedTotalName, []string{line.NginxHost}, float64(0))
	}

	e.exportMetrics(data, line.NginxHost, ctx)
}

// exportMetrics exports defined metrics.
func (e *ExportWorker) exportMetrics(data map[string]string, nginxHost string, ctx context.Context) {
	// try to detect user agent, os, device using custom settings from config
	uaLbs := e.tryDetectCustomUserAgentLabels(data)
	if uaLbs == nil {
		uaLbs = &uaLabels{internalLabelValue, internalLabelValue, internalLabelValue}
		// check if it is needed to parse user agent labels
		if e.needParseUserAgent(data, ctx) {
			uaLbs = e.detectUserAgentLabels(data, nginxHost)
		} else {
			e.exposeFunc(exposer.LogsFilteredTotal, []string{nginxHost}, float64(0))
		}
	}

	// detect device type label
	deviceType := e.detectDeviceType(data, uaLbs.userAgent, uaLbs.os, uaLbs.device)

	// detect http code label
	httpCode, err := e.detectHttpCodeLabel(data)
	if err != nil {
		logging.WithContext(ctx).Sugar().Errorf("could not parse http code: %s", err)
	}

	// detect host label
	host := e.detectHostLabel(data)

	// detect URI label
	URI := e.detectURILabel(data)

	// detect response duration metric value
	responseDuration, ok, err := e.detectResponseDuration(data)
	if err != nil {
		logging.WithContext(ctx).Sugar().Warnf("could not detect response duration: %s", err)
	}

	// expose metrics
	if ok {
		// response time by host and http code
		e.exposeFunc(exposer.HostResponseTimeSecondsMetricName, []string{host, httpCode}, float64(responseDuration))
		// response time by host, user agent and http code
		e.exposeFunc(exposer.UserAgentResponseTimeSecondsMetricName, []string{host, uaLbs.userAgent, httpCode}, float64(responseDuration))
		// response time by host, URI and http code
		e.exposeFunc(exposer.URIResponseTimeSecondsMetricName, []string{host, URI, httpCode}, float64(responseDuration))
	}

	// requests count by host, user agent and http code
	e.exposeFunc(exposer.UserAgentRequestsTotalMetricName, []string{host, uaLbs.userAgent, httpCode}, float64(0))
	// requests count by host, os and device type
	e.exposeFunc(exposer.OsDeviceTypeRequestsTotalMetricName, []string{host, uaLbs.os, deviceType}, float64(0))
	// requests by nginx host
	e.exposeFunc(exposer.NginxRequestsTotal, []string{nginxHost}, float64(0))
}

// tryDetectCustomUserAgentLabels tries to detect user agent using custom settings from config.
func (e *ExportWorker) tryDetectCustomUserAgentLabels(data map[string]string) *uaLabels {
	var uaLbs *uaLabels

	if v, ok := data[httpUserAgentVar]; ok {
		// detect custom user agents
		for _, rep := range *e.userAgentReplacementSettings {
			if rep.MatchRe != nil && rep.MatchRe.MatchString(v) {
				uaLbs = &uaLabels{
					rep.MatchRe.ReplaceAllString(rep.MatchRe.FindString(v), rep.Replacements.UserAgent),
					rep.MatchRe.ReplaceAllString(rep.MatchRe.FindString(v), rep.Replacements.Os),
					rep.MatchRe.ReplaceAllString(rep.MatchRe.FindString(v), rep.Replacements.Device),
				}
			} else if rep.Match != "" && strings.ToLower(rep.Match) == strings.ToLower(v) {
				uaLbs = &uaLabels{
					rep.Replacements.UserAgent,
					rep.Replacements.Os,
					rep.Replacements.Device,
				}
			}
		}
	}

	return uaLbs
}

// detectFormat tries to detect log format according config.
func (e *ExportWorker) detectFormat(ctx context.Context, nginxHost string) string {
	for _, source := range *e.sources {
		if source.Host == nginxHost {
			return source.LogFormat
		}
	}

	logging.WithContext(ctx).Sugar().Infof("default log format detected for: %s", nginxHost)

	return defaultLogFormat
}

// detectUserAgentLabels tries to detect user agent data.
func (e *ExportWorker) detectUserAgentLabels(data map[string]string, nginxHost string) *uaLabels {
	uaLbs := &uaLabels{unknownLabelValue, unknownLabelValue, unknownLabelValue}

	if v, ok := data[httpUserAgentVar]; ok {
		labels, ok := e.cc.Get(v)
		if !ok {
			client := e.userAgentPsr.Parse(v)

			labels = map[string]string{
				userAgentLabelName: client.UserAgent.Family,
				osLabelName:        strings.Title(client.Os.Family),
				deviceLabelName:    client.Device.Family,
			}

			e.cc.Set(v, labels)

			e.exposeFunc(exposer.UserAgentCachedTotal, []string{nginxHost}, float64(0))
			e.exposeFunc(exposer.UserAgentCurrentCachedTotal, []string{nginxHost}, float64(e.cc.Len()))
		}

		uaLbs.userAgent = labels[userAgentLabelName]
		uaLbs.os = labels[osLabelName]
		uaLbs.device = labels[deviceLabelName]
	}

	return uaLbs
}

// needParseUserAgent detects whether it is needed to parse user agent.
func (e *ExportWorker) needParseUserAgent(data map[string]string, ctx context.Context) bool {
	rAddr, ok := data[remoteAddrVar]
	if !ok {
		return true
	}

	contains, err := net.IsSubnetContainsIP(rAddr, *e.internalSubNets)
	if err != nil {
		logging.WithContext(ctx).Sugar().Warnf("could not detect if ip belongs to subnet: %s", err)

		return true
	}

	return !contains
}

// detectHttpCodeLabel tries to detect http code.
func (e *ExportWorker) detectHttpCodeLabel(data map[string]string) (string, error) {
	if v, ok := data[statusVar]; ok {
		code, err := strconv.Atoi(v)
		if err != nil || code == 0 {
			return unknownLabelValue, err
		}

		return strconv.Itoa(code), nil
	}

	return unknownLabelValue, nil
}

// detectURILabel detects URI label.
func (e *ExportWorker) detectURILabel(data map[string]string) string {
	if v, ok := data[requestVar]; ok {
		request := strings.Split(v, " ")
		if len(request) < 2 {
			return ""
		}

		method := request[0]
		URI := request[1]

		for _, rep := range *e.requestURIReplacementSettings {
			if rep.Method != "" && rep.Method != method {
				return ""
			}
			if path := rep.Regexp.FindString(URI); path != "" {
				return rep.Regexp.ReplaceAllString(path, rep.Replacements.RequestURI)
			}
		}
	}

	return ""
}

// detectHostLabel detects host of request
func (e *ExportWorker) detectHostLabel(data map[string]string) string {
	if v, ok := data[hostVar]; ok {
		// check if there is replacement for host label
		if e.hosts != nil {
			for _, repl := range *e.hosts {
				if repl.Match == v {
					return repl.Replacement
				}
			}
		}

		return v
	}

	return unknownLabelValue
}

// detectResponseDuration detects response duration
func (e *ExportWorker) detectResponseDuration(data map[string]string) (float64, bool, error) {
	v, ok := data[requestTimeVar]
	if !ok {
		return float64(0), false, nil
	}

	duration, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return float64(0), false, err
	}

	return duration, true, nil
}

// detectDeviceType detects device type
func (e ExportWorker) detectDeviceType(data map[string]string, userAgentFamily, osFamily, deviceFamily string) string {
	userAgent, ok := data[httpUserAgentVar]
	if !ok {
		return deviceTypeDesktop
	}

	// deal with apple devices
	if strings.Contains(userAgent, "ipad") {
		return deviceTypeMobile
	}

	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipod") {
		return deviceTypeMobile
	}

	// check generic cases
	if deviceFamily == "generic tablet" {
		return deviceTypeTablet
	}

	if deviceFamily == "generic smartphone" || deviceFamily == "generic feature phone" {
		return deviceTypeMobile
	}

	// blackberry
	if osFamily == "blackberry tablet os" {
		return deviceTypeTablet
	}

	if osFamily == "blackberry os" {
		return deviceTypeMobile
	}

	// ie doesn't separate tablet/mobile cases, so we treat it as mobile
	// it better to show mobile version to tablet user rather than the opposite
	if strings.Contains(userAgent, "windows") && strings.Contains(userAgent, "touch") {
		return deviceTypeMobile
	}

	// a couple of general basic tests
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgentFamily, "mobile") {
		return deviceTypeMobile
	}

	// at this point we have android device without "mobile" substring. It's more appropriate to treat it as a tablet
	// http://android-developers.blogspot.ru/2010/12/android-browser-user-agent-issues.html
	if osFamily == "android" {
		return deviceTypeTablet
	}

	return deviceTypeDesktop
}
