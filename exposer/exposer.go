package exposer

// Exposer is an interface for service that exposes metrics
type Exposer func(name string, labels []string, value float64)

// NewPromExposer return function that exposes metrics for Prometheus
func PromExposer(name string, labels []string, value float64) {
	switch name {
	case HostResponseTimeSecondsMetricName:
		hostResponseTimeSeconds.WithLabelValues(labels...).Observe(value)
	case UserAgentResponseTimeSecondsMetricName:
		userAgentResponseTimeSeconds.WithLabelValues(labels...).Observe(value)
	case UserAgentRequestsTotalMetricName:
		userAgentRequestsTotal.WithLabelValues(labels...).Inc()
	case OsDeviceTypeRequestsTotalMetricName:
		osDeviceTypeRequestsTotal.WithLabelValues(labels...).Inc()
	case URIResponseTimeSecondsMetricName:
		URIResponseTimeSeconds.WithLabelValues(labels...).Observe(value)
	case NginxRequestsTotal:
		nginxRequestsTotal.WithLabelValues(labels...).Inc()
	case LogsDroppedTotalName:
		logsDropped.WithLabelValues(labels...).Inc()
	case LogsFailParsedTotalName:
		logsFailParsedTotal.WithLabelValues(labels...).Inc()
	case LogsTotal:
		logsTotal.WithLabelValues(labels...).Inc()
	case LogsFilteredTotal:
		logsFilteredTotal.WithLabelValues(labels...).Inc()
	case UserAgentCachedTotal:
		userAgentCachedTotal.WithLabelValues(labels...).Inc()
	case UserAgentCurrentCachedTotal:
		userAgentCurrentCachedTotal.WithLabelValues(labels...).Set(value)
	}
}
