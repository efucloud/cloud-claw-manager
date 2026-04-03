package metrics

import "github.com/efucloud/cloud-claw-manager/pkg/dtos"

func nodeServer() {
	nodeMetrics = map[string]dtos.PrometheusQuery{
		"CPU Utilisation": {
			Query: `instance:node_cpu_utilisation:rate5m{job="node-exporter", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_} != 0`,
		},
		"Memory Utilisation": {
			Query: `instance:node_memory_utilisation:ratio{job="node-exporter", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_} != 0`,
		},
		"CPU Saturation (Load1 per CPU)": {
			Query: `instance:node_load1_per_cpu:ratio{job="node-exporter", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_} != 0`,
		},
		"Memory Saturation (Major Page Faults)": {
			Query: `instance:node_vmstat_pgmajfault:rate5m{job="node-exporter", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_} != 0`,
		},
		"Disk Space Utilisation": {
			Query: `sort_desc(1 -
  (
    max without (mountpoint, fstype) (node_filesystem_avail_bytes{job="node-exporter", fstype!="", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})
    /
    max without (mountpoint, fstype) (node_filesystem_size_bytes{job="node-exporter", fstype!="", instance="_{{_ .Node _}}_" _{{_ if .Cluster _}}_ , cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})
  ) != 0
)
`,
		},
	}
}
