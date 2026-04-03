package metrics

import "github.com/efucloud/cloud-claw-manager/pkg/dtos"

func podServer() {
	podMetrics = map[string]dtos.PrometheusQuery{
		"CPU Requests": {
			I18N:  dtos.I18N{ZH: "CPU 请求量", EN: "CPU Requests"},
			Query: `sum(kube_pod_container_resource_requests{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", resource="cpu"})`,
		},
		"CPU Limits": {
			I18N:  dtos.I18N{ZH: "CPU 限制量", EN: "CPU Limits"},
			Query: `sum(  kube_pod_container_resource_limits{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", resource="cpu"})`,
		},
		"Container CPU Usage": {
			I18N:  dtos.I18N{ZH: "CPU 使用量", EN: "CPU Usage"},
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container)`,
		},
		"CPU Throttling": {
			Query: `sum(increase(container_cpu_cfs_throttled_periods_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}[1m])) by (container) /sum(increase(container_cpu_cfs_periods_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}[1m])) by (container)`,
		},
		"Container CPU Requests": {
			Query: `sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container)`,
		},
		"Container CPU Requests%": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container) / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container)`,
		},
		"Container CPU Limits": {
			Query: `sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container)`,
		},
		"Container CPU Limits%": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container) / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!=""}) by (container)`,
		},
		"Memory Usage (WSS)": {
			Query: `sum(container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!="", image!=""}) by (container)`,
		},
		"Container Memory Requests": {
			Query: `sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}) by (container)`,
		},
		"Container Memory Requests%": {
			Query: `sum(container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", image!=""}) by (container) / sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}) by (container)`,
		},
		"Container Memory Limits": {
			Query: `sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}) by (container)`,
		},
		"Container Memory Limits%": {
			Query: `sum(container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container!="", image!=""}) by (container) / sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits{ namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}) by (container)`,
		},
		"Memory Usage (RSS)": {
			Query: `sum(container_memory_rss{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container != "", container != "POD"}) by (container)`,
		},
		"Memory Usage (Cache)": {
			Query: `sum(container_memory_cache{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container != "", container != "POD"}) by (container)`,
		},
		"Memory Usage (Swap)": {
			Query: `sum(container_memory_swap{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_", container != "", container != "POD"}) by (container)`,
		},
		"Receive Bandwidth": {
			Query: `sum(irate(container_network_receive_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"Transmit Bandwidth": {
			Query: `sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"Rate of Received Packets": {
			Query: `sum(rate(container_network_receive_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"Rate of Transmitted Packets": {
			Query: `sum(rate(container_network_transmit_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"Rate of Received Packets Dropped": {
			Query: `sum(rate(container_network_receive_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"Rate of Transmitted Packets Dropped": {
			Query: `sum(rate(container_network_transmit_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])) by (pod)`,
		},
		"IOPS Reads (Pod)": {
			Query: `ceil(sum by(pod) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])))`,
		},
		"IOPS Writes (Pod)": {
			Query: `ceil(sum by(pod) (rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="",namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m])))`,
		},
		"ThroughPut Reads (Pod)": {
			Query: `sum by(pod) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m]))`,
		},
		"ThroughPut Writes (Pod)": {
			Query: `sum by(pod) (rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod=~"_{{_ .Pod _}}_"}[1m]))`,
		},
		"IOPS (Containers)": {
			Query: `ceil(sum by(container) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]) + rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m])))`,
		},
		"ThroughPut (Containers)": {
			Query: `sum by(container) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]) + rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"IOPS Reads Container": {
			Query: `sum by(container) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"IOPS Writes Container": {
			Query: `sum by(container) (rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor",device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"IOPS(Reads + Writes) Container": {
			Query: `sum by(container) (rate(container_fs_reads_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]) + rate(container_fs_writes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"Throughput(Read) Container": {
			Query: `sum by(container) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"Throughput(Write) Container": {
			Query: `sum by(container) (rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
		"Throughput(Read+Write) Container": {
			Query: `sum by(container) (rate(container_fs_reads_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]) + rate(container_fs_writes_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", device=~"(/dev/)?(mmcblk.p.+|nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+|md.+|dasd.+)", container!="", namespace="_{{_ .Namespace _}}_", pod="_{{_ .Pod _}}_"}[1m]))`,
		},
	}
}
