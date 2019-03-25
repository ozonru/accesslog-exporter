package exposer

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "accesslog"

	LogsDroppedTotalName                   = "logs_dropped_total"
	LogsFailParsedTotalName                = "logs_fail_parsed_total"
	LogsTotal                              = "logs_total"
	LogsFilteredTotal                      = "logs_filtered_total"
	UserAgentCachedTotal                   = "user_agent_cached_total"
	UserAgentCurrentCachedTotal            = "user_agent_current_cached_total"
	HostResponseTimeSecondsMetricName      = "host_response_time_seconds"
	UserAgentResponseTimeSecondsMetricName = "user_agent_response_time_seconds"
	UserAgentRequestsTotalMetricName       = "user_agent_requests_total"
	OsDeviceTypeRequestsTotalMetricName    = "os_device_type_requests_total"
	URIResponseTimeSecondsMetricName       = "uri_response_time_seconds"
	NginxRequestsTotal                     = "nginx_requests_total"
)

var (
	Version  string
	Revision string
	Branch   string
)

var (
	// business metrics
	hostResponseTimeSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      HostResponseTimeSecondsMetricName,
		Help:      "Response time by host in seconds",
	}, []string{"host", "code"})
	userAgentResponseTimeSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      UserAgentResponseTimeSecondsMetricName,
		Help:      "Response time by user agent in seconds",
	}, []string{"host", "user_agent", "code"})
	userAgentRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      UserAgentRequestsTotalMetricName,
		Help:      "Requests total by user agent",
	}, []string{"host", "user_agent", "code"})
	osDeviceTypeRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      OsDeviceTypeRequestsTotalMetricName,
		Help:      "Requests total by os and device type",
	}, []string{"host", "os", "device_type"})
	URIResponseTimeSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      URIResponseTimeSecondsMetricName,
		Help:      "Response time by uri in seconds",
	}, []string{"host", "uri", "code"})
	nginxRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      NginxRequestsTotal,
		Help:      "Total requests by nginx host",
	}, []string{"host"})

	// internal accesslog exporter metrics
	accesslogBuildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "build_info",
		Help:      "A metric with a constant '1' value labeled by version, revision, and branch from which the node_exporter was built.",
	}, []string{"version", "revision", "branch"})
	logsDropped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      LogsDroppedTotalName,
		Help:      "Logs that were dropped",
	}, []string{"nginx_host"})
	logsFailParsedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      LogsFailParsedTotalName,
		Help:      "Total fail parsed logs",
	}, []string{"nginx_host"})
	logsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      LogsTotal,
		Help:      "Total log lines",
	}, []string{"nginx_host"})
	logsFilteredTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      LogsFilteredTotal,
		Help:      "Total filtered logs by subnet",
	}, []string{"nginx_host"})
	userAgentCachedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      UserAgentCachedTotal,
		Help:      "Total cached user agents",
	}, []string{"nginx_host"})
	userAgentCurrentCachedTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      UserAgentCurrentCachedTotal,
		Help:      "Total current cached user agents",
	}, []string{"nginx_host"})
)

func init() {
	prometheus.MustRegister(
		nginxRequestsTotal,
		hostResponseTimeSeconds,
		userAgentResponseTimeSeconds,
		userAgentRequestsTotal,
		osDeviceTypeRequestsTotal,
		URIResponseTimeSeconds,

		accesslogBuildInfo,
		logsDropped,
		logsFailParsedTotal,
		logsTotal,
		logsFilteredTotal,
		userAgentCachedTotal,
		userAgentCurrentCachedTotal,
	)

	accesslogBuildInfo.WithLabelValues(Version, Revision, Branch).Set(1)
}
