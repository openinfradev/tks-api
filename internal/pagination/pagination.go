package pagination

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type Pagination struct {
	Limit      int
	Page       int
	SortColumn string
	SortOrder  string
	Filters    []Filter
	TotalRows  int64
	TotalPages int
}

type Filter struct {
	Column string
	Value  string
}

var DEFAULT_LIMIT = 10
var MAX_LIMIT = 1000

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = DEFAULT_LIMIT
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSortColumn() string {
	return p.SortColumn
}

func (p *Pagination) GetSortOrder() string {
	return p.SortOrder
}

func (p *Pagination) GetFilter() []Filter {
	return p.Filters
}

/*
	{
		sortingColumn : "id",
		order : "ASC",
		page : 1,
		limit : 10,
	}
*/
func NewPagination(urlParams *url.Values) Pagination {
	var pg Pagination

	pg.SortColumn = urlParams.Get("sortColumn")
	if pg.SortColumn == "" {
		pg.SortColumn = "created_at"
	}
	pg.SortOrder = urlParams.Get("sortOrder")
	if pg.SortOrder == "" {
		pg.SortOrder = "ASC"
	}

	page := urlParams.Get("pageNumber")
	if page == "" {
		pg.Page = 1
	} else {
		pg.Page, _ = strconv.Atoi(page)
	}

	limit := urlParams.Get("pageSize")
	if limit == "" {
		pg.Limit = DEFAULT_LIMIT
	} else {
		limitNum, err := strconv.Atoi(limit)
		if err == nil {
			pg.Limit = limitNum
		}
	}

	// [TODO] filter
	filter := urlParams.Get("filter")
	if filter != "" {
		_ = json.Unmarshal([]byte(filter), &pg.Filters)
	}

	return pg
}

func NewDefaultPagination() Pagination {
	return Pagination{
		SortColumn: "created_at",
		SortOrder:  "ASC",
		Page:       1,
		Limit:      MAX_LIMIT,
	}
}
