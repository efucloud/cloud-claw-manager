package metrics

import "github.com/efucloud/cloud-claw-manager/pkg/dtos"

func namespaceServer() {
	namespaceMetrics = map[string]dtos.PrometheusQuery{
		//总览
		"CPU Usage Total": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"})`,
		},

		"Receive Bandwidth": {
			I18N: dtos.I18N{ZH: "Receive Bandwidth", EN: "Receive Bandwidth"},
			Query: `(sum(rate(container_network_receive_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload=~".+", workload_type=~"(daemonset|deployment|statefulset)"}))`,
		},
		"Transmit Bandwidth": {
			I18N: dtos.I18N{ZH: "Transmit Bandwidth", EN: "Transmit Bandwidth"},
			Query: `(sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload=~".+", workload_type=~"(daemonset|deployment|statefulset)"}))`,
		},
		"Memory Usage Total": {
			Query: `sum(container_memory_working_set_bytes{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", container!="", image!=""}
  * on(namespace,pod)
    group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"})`,
		},
		"Current Rate of Bytes Received": {
			I18N: dtos.I18N{ZH: "当前字节接收速率", EN: "Current Rate of Bytes Received"},
			Query: `sort_desc(sum(rate(container_network_receive_bytes_total{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (cluster,namespace,pod) group_left ()
    topk by (cluster,namespace,pod) (
      1,
      max by (cluster,namespace,pod) (kube_pod_info{host_network="false"})
    )
* on (cluster,namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload=~".+", workload_type=~"(daemonset|deployment|statefulset)"}))`,
		},
		"Current Rate of Bytes Transmitted": {
			I18N: dtos.I18N{ZH: "当前字节发送速率", EN: "Current Rate of Bytes Transmitted"},
			Query: `sort_desc(sum(rate(container_network_transmit_bytes_total{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (cluster,namespace,pod) group_left ()
    topk by (cluster,namespace,pod) (
      1,
      max by (cluster,namespace,pod) (kube_pod_info{host_network="false"})
    )
* on (cluster,namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload=~".+", workload_type=~"(daemonset|deployment|statefulset)"}))`,
		},
		//工作负载
		"Workload Type": {
			Query: `count(namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}) by (workload, workload_type)`,
		},
		"Workload CPU Usage": {
			Query: `sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload CPU Request": {
			Query: `sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload CPU Request Total": {
			Query: `sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
)`,
		},
		"Workload CPU Request%": {
			Query: `sum(
  node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)
/sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)
`,
		},
		"Workload CPU Limit Total": {
			Query: `sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
)`,
		},
		"Workload CPU Limit": {
			Query: `sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload CPU Limit%": {
			Query: `sum(
  node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate5m{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)
/sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="cpu"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload Memory Request%": {
			Query: `sum(
    container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", container!="", image!=""}
  * on(namespace,pod)
    group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}
) by (workload, workload_type)
/sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}
) by (workload, workload_type)`,
		},
		"Workload Memory Limit%": {
			Query: `sum(
    container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", container!="", image!=""}
  * on(namespace,pod)
    group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}
) by (workload, workload_type)
/sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}
) by (workload, workload_type)
`,
		},
		"Workload Memory Usage": {
			Query: `sum(
    container_memory_working_set_bytes{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", container!="", image!=""}
  * on(namespace,pod)
    group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload Memory Request": {
			Query: `sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}
) by (workload, workload_type)
`,
		},
		"Workload Memory Limit": {
			Query: `sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
) by (workload, workload_type)`,
		},
		"Workload Memory Request Total": {
			Query: `sum(
  kube_pod_container_resource_requests{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
)`,
		},
		"Workload Memory Limit Total": {
			Query: `sum(
  kube_pod_container_resource_limits{ _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", resource="memory"}
* on(namespace,pod)
  group_left(workload, workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}
)`,
		},
		"Workload Current Receive Bandwidth": {
			Query: `(sum(rate(container_network_receive_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
		"Workload Current Receive Transmitted": {
			Query: `(sum(rate(container_network_transmit_bytes_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
		"Workload Rate of Received Packets": {
			Query: `(sum(rate(container_network_receive_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
		"Workload Rate of Transmitted Packets": {
			Query: `(sum(rate(container_network_transmit_packets_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
		"Workload Rate of Received Packets Dropped": {
			Query: `(sum(rate(container_network_receive_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
		"Workload Rate of Transmitted Packets Dropped": {
			Query: `(sum(rate(container_network_transmit_packets_dropped_total{job="kubelet", metrics_path="/metrics/cadvisor", _{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_"}[1m])
* on (namespace,pod)
group_left(workload,workload_type) namespace_workload_pod:kube_pod_owner:relabel{_{{_ if .Cluster _}}_ cluster="_{{_ .Cluster _}}_",_{{_ end _}}_ namespace="_{{_ .Namespace _}}_", workload_type=~"(daemonset|deployment|statefulset)"}) by (workload))
`,
		},
	}
}
