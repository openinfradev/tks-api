package domain

type PaginationResponse struct {
	Limit      int              `json:"limit"`
	Page       int              `json:"page"`
	SortColumn string           `json:"sortColumn"`
	SortOrder  string           `json:"sortOrder"`
	Filters    []FilterResponse `json:"filters,omitempty"`
	TotalRows  int64            `json:"totalRows"`
	TotalPages int              `json:"totalPages"`
}

type FilterResponse struct {
	Column string `json:"column"`
	Value  string `json:"value"`
}
