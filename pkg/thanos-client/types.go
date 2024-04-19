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

// PolicyMetric dedicated policy metric struct
type PolicyMetric struct {
	Data   PolicyMetricData `json:"data"`
	Status string           `json:"status"`
}

type PolicyMetricData struct {
	Result     []PolicyMetricResult `json:"result"`
	ResultType string               `json:"resultType"`
}

type PolicyMetricResult struct {
	Metric PolicyMetricDataResultMetric `json:"metric"`
	Value  []interface{}                `json:"value"`
}

type PolicyMetricDataResultMetric struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Violation string `json:"violation_enforcement"`
}

type WorkloadMetric struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type PolicyTemplateMetric struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Kind string `json:"kind"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type PolicyViolationCountMetric struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				ViolationEnforcement string `json:"violation_enforcement,omitempty"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
