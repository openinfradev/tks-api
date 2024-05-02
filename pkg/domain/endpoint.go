package domain

type EndpointResponse struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type ListEndpointResponse struct {
	Endpoints []EndpointResponse `json:"endpoints"`
}
