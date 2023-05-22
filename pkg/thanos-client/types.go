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
	Values []interface{}          `json:"values"`
}

type MetricDataResultMetric struct {
	Name        string `json:"__name__"`
	TacoCluster string `json:"taco_cluster"`
}
