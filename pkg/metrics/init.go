package metrics

import (
	"fmt"
	"github.com/efucloud/cloud-claw-manager/pkg/dtos"
	"github.com/efucloud/common"
)

var allMetrics map[string]map[string]dtos.PrometheusQuery
var apiServerMetrics map[string]dtos.PrometheusQuery
var podMetrics map[string]dtos.PrometheusQuery
var workloadMetrics map[string]dtos.PrometheusQuery
var namespaceMetrics map[string]dtos.PrometheusQuery
var clusterMetrics map[string]dtos.PrometheusQuery
var nodeMetrics map[string]dtos.PrometheusQuery

func GetPrometheusQuery(view, key string) (result dtos.PrometheusQuery, errorData common.ErrorData) {
	if metrics, exist := allMetrics[view]; exist {
		if query, ex := metrics[key]; ex {
			return query, errorData
		} else {
			errorData.Err = fmt.Errorf("code: %s not found", key)
		}
	} else {
		errorData.Err = fmt.Errorf("view: %s not found", view)
	}
	return
}
func init() {
	apiServerMetrics = map[string]dtos.PrometheusQuery{}
	apiServer()
	podServer()
	workloadServer()
	namespaceServer()
	clusterServer()
	nodeServer()
	allMetrics = make(map[string]map[string]dtos.PrometheusQuery)
	allMetrics["apiserver"] = apiServerMetrics
	allMetrics["pod"] = podMetrics
	allMetrics["workload"] = workloadMetrics
	allMetrics["namespace"] = namespaceMetrics
	allMetrics["cluster"] = clusterMetrics
	allMetrics["node"] = nodeMetrics

}
