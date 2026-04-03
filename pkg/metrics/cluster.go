package metrics

import "github.com/efucloud/cloud-claw-manager/pkg/dtos"

func clusterServer() {
	clusterMetrics = map[string]dtos.PrometheusQuery{
		"Node Status": {
			Query: `count by (status) (  kube_node_status_condition{condition="Ready", status=~"true|false|unknown"} == 1)`,
		},
		"Namespace Status": {
			Query: `count by (phase) (  kube_namespace_status_phase{namespace!=""} == 1)`,
		},
		"Network Received": {
			Query: `sum(rate(node_network_receive_bytes_total{device=~"eth0|ens.*|bond0"}[5m])) / 1024 / 1024`,
		},
		"Network Transmit": {
			Query: `sum(rate(node_network_transmit_bytes_total{device=~"eth0|ens.*|bond0"}[5m])) / 1024 / 1024`,
		},
		"CPU Total": {
			Query: `sum(machine_cpu_cores)`,
		},
		"Memory Total": {
			Query: `sum(machine_memory_bytes/(1024^3))`,
		},
		"CPU Utilisation": {
			Query: `cluster:node_cpu:ratio_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}`,
		},
		"CPU Requests Commitment": {
			Query: `sum(namespace_cpu:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum(kube_node_status_allocatable{resource="cpu" _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
		},
		"CPU Limits Commitment": {
			Query: `sum(namespace_cpu:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum(kube_node_status_allocatable{resource="cpu" _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
		},
		"Memory Utilisation": {
			Query: `1 - sum(:node_memory_MemAvailable_bytes:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum(node_memory_MemTotal_bytes{job="node-exporter" _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
		},
		"Memory Requests Commitment": {
			Query: `sum(namespace_memory:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum(kube_node_status_allocatable{resource="memory" _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
		},
		"Memory Limits Commitment": {
			Query: `sum(namespace_memory:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) / sum(kube_node_status_allocatable{resource="memory" _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_})`,
		},
		"CPU Usage": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"PodsByNamespace": {
			Query: `sum(kube_pod_owner{ cluster="_{{_ .Cluster _}}_"}) by (namespace)`,
		},
		"WorkloadsByNamespace": {
			Query: `count(avg(namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (workload, namespace)) by (namespace)`,
		},
		"CPUUsageByNamespace": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"CPURequestsByNamespace": {
			Query: `sum(namespace_cpu:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"CPURequestsByNamespace%": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace) / sum(namespace_cpu:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"CPULimitsByNamespace": {
			Query: `sum(namespace_cpu:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"CPUByNamespaceLimits%": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace) / sum(namespace_cpu:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"MemoryUsageByNamespace": {
			Query: `sum(container_memory_rss{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ container!=""}) by (namespace)`,
		},
		"MemoryRequestsByNamespace": {
			Query: `sum(namespace_memory:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"MemoryRequestsByNamespace%": {
			Query: `sum(container_memory_rss{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ container!=""}) by (namespace) / sum(namespace_memory:kube_pod_container_resource_requests:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"MemoryLimitsByNamespace": {
			Query: `sum(namespace_memory:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"MemoryByNamespaceLimits%": {
			Query: `sum(container_memory_rss{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ container!=""}) by (namespace) / sum(namespace_memory:kube_pod_container_resource_limits:sum{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_"_{{_ end _}}_}) by (namespace)`,
		},
		"Memory": {
			Query: `sum(container_memory_rss{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ container!=""}) by (namespace)`,
		},
		"Receive Bandwidth": {
			Query: `sum(rate(container_network_receive_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Transmit Bandwidth": {
			Query: `sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Average Container Bandwidth by Namespace: Received": {
			Query: `sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Average Container Bandwidth by Namespace: Transmitted": {
			Query: `sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Rate of Received Packets": {
			Query: `sum(irate(container_network_receive_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Rate of Transmitted Packets": {
			Query: `sum(irate(container_network_transmit_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Rate of Received Packets Dropped": {
			Query: `sum(irate(container_network_receive_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"Rate of Transmitted Packets Dropped": {
			Query: `sum(irate(container_network_transmit_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace=~".+"}[1m])) by (namespace)`,
		},
		"IOPS(Reads)": {
			Query: `sum by(namespace) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]))`,
		},
		"IOPS(Writes)": {
			Query: `sum by(namespace) (rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]))`,
		},
		"IOPS(Reads+Writes)": {
			Query: `ceil(sum by(namespace) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]) + rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m])))`,
		},
		"ThroughPut(Read)": {
			Query: `sum by(namespace) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]))`,
		},
		"ThroughPut(Write)": {
			Query: `sum by(namespace) (rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]))`,
		},
		"ThroughPut(Read+Write)": {
			Query: `sum by(namespace) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]) + rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace!=""}[1m]))`,
		},
	}

}
