package thanos

type Metric struct {
	Data   MetricData `json:"data"`
	Status string     `json:"status"`
}

type MetricData struct {
	Result     []MetricDataResult `json:"result"`
	ResultType string             `json:"vector"`
}

type MetricDataResult struct {
	Metric MetricDataResultMetric `json:"metric"`
	Value  []interface{}          `json:"value"`
}

type MetricDataResultMetric struct {
	TacoCluster string `json:"taco_cluster"`
}
