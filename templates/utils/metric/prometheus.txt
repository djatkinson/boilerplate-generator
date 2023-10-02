package metric

import "github.com/prometheus/client_golang/prometheus"

func SetupPrometheus() *prometheus.CounterVec {
	outboundCall := prometheus.CounterOpts{
		Name: "http_request_outbound_callbacks",
		Help: "Get Information Callback from outbound http call",
	}
	outboundCallLabelName := []string{
		"host",
		"path",
		"response_code",
	}

	metricOutbound := prometheus.NewCounterVec(outboundCall, outboundCallLabelName)
	return metricOutbound
}
