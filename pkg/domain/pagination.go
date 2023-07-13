package domain

type PaginationResponse struct {
	Limit      int              `json:"pageSize"`
	Page       int              `json:"pageNumber"`
	SortColumn string           `json:"sortColumn"`
	SortOrder  string           `json:"sortOrder"`
	Filters    []FilterResponse `json:"filters,omitempty"`
	TotalRows  int64            `json:"totalRows"`
	TotalPages int              `json:"totalPages"`
}

type FilterResponse struct {
	Column string   `json:"column"`
	Values []string `json:"values"`
}
