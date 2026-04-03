package metrics

import "github.com/efucloud/cloud-claw-manager/pkg/dtos"

func apiServer() {
	apiServerMetrics = map[string]dtos.PrometheusQuery{
		"Availability": {
			I18N:        dtos.I18N{ZH: "可用性", EN: "Availability"},
			Query:       `apiserver_request:availability30d{verb="all",  cluster="_{{_ .Cluster _}}_"}`,
			Description: "How many percent of requests (both read and write) in 30 days have been answered successfully and fast enough?",
		},
		"ReadAvailability": {
			I18N:        dtos.I18N{ZH: "读可用性", EN: "Read Availability"},
			Query:       `apiserver_request:availability30d{verb="read" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
			Description: "How many percent of read requests (LIST,GET) in 30 days have been answered successfully and fast enough?",
		},
		"WriteAvailability": {
			I18N:        dtos.I18N{ZH: "写可用性", EN: "Write Availability"},
			Query:       `apiserver_request:availability30d{verb="write" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
			Description: "How many percent of write requests (POST|PUT|PATCH|DELETE) in 30 days have been answered successfully and fast enough?",
		},
		"ErrorBudget": {
			I18N:        dtos.I18N{ZH: "ErrorBudget", EN: "ErrorBudget"},
			Query:       `100 * (apiserver_request:availability30d{verb="all" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_} - 0.990000)`,
			Description: "How much error budget is left looking at our 0.990% availability guarantees?",
		},
		"Read SLI - Requests": {
			I18N:        dtos.I18N{ZH: "Read SLI - Requests", EN: "Read SLI - Requests"},
			Query:       `sum by (code) (code_resource:apiserver_request_total:rate5m{verb="read" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
			Description: "How many read requests (LIST,GET) per second do the apiservers get by code?",
		},
		"Read SLI - Errors": {
			I18N:        dtos.I18N{ZH: "Read SLI - Error", EN: "Read SLI - Error"},
			Query:       `sum by (resource) (code_resource:apiserver_request_total:rate5m{verb="read",code=~"5.." _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum by (resource) (code_resource:apiserver_request_total:rate5m{verb="read" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
			Description: "How many read requests (LIST,GET) per second do the apiservers get by code?",
		},
		"Read SLI - Duration": {
			I18N:        dtos.I18N{ZH: "Read SLI - Duration", EN: "Read SLI - Duration"},
			Query:       `cluster_quantile:apiserver_request_sli_duration_seconds:histogram_quantile{verb="read" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
			Description: "How many seconds is the 99th percentile for reading (LIST|GET) a given resource?",
		},
		"Write SLI - Requests": {
			Query:       `sum by (code) (code_resource:apiserver_request_total:rate5m{verb="write" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
			Description: "How many write requests (POST|PUT|PATCH|DELETE) per second do the apiservers get by code?",
		},
		"Write SLI - Errors": {
			Query:       `sum by (resource) (code_resource:apiserver_request_total:rate5m{verb="write",code=~"5.." _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum by (resource) (code_resource:apiserver_request_total:rate5m{verb="write" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
			Description: "How many percent of write requests (POST|PUT|PATCH|DELETE) per second are returned with errors (5xx)?",
		},
		"Write SLI - Duration": {
			Query:       `cluster_quantile:apiserver_request_sli_duration_seconds:histogram_quantile{verb="write" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
			Description: "How many seconds is the 99th percentile for writing (POST|PUT|PATCH|DELETE) a given resource??",
		},
		"Work Queue Add Rate": {
			Query: `sum(rate(workqueue_adds_total{job="apiserver",  cluster="_{{_ .Cluster _}}_"}[1m])) by (instance, name)`,
		},
		"Work Queue Depth": {
			Query: `sum(rate(workqueue_depth{job="apiserver" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}[1m])) by (instance, name)`,
		},
		"Work Queue Latency": {
			Query: `histogram_quantile(0.99, sum(rate(workqueue_queue_duration_seconds_bucket{job="apiserver" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}[1m])) by (instance, name, le))`,
		},
		"Memory": {
			Query: `process_resident_memory_bytes{job="apiserver" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
		},
		"CPU usage": {
			Query: `rate(process_cpu_seconds_total{job="apiserver" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}[1m])`,
		},
		"Goroutines": {
			Query: `go_goroutines{job="apiserver" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
		},
	}
}
