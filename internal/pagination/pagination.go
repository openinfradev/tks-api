package pagination

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	filter "github.com/openinfradev/tks-api/internal/filter"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/serializer"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"

	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/database"
)

const SORT_COLUMN = "sortColumn"
const SORT_ORDER = "sortOrder"
const PAGE_NUMBER = "pageNumber"
const PAGE_SIZE = "pageSize"
const FILTER = "filter"
const FILTER_ARRAY = "filter[]"
const OR = "or"
const OR_ARRAY = "or[]"
const COMBINED_FILTER = "combinedFilter" // deprecated

var DEFAULT_LIMIT = 10000

type Pagination struct {
	Limit          int
	Page           int
	SortColumn     string
	SortOrder      string
	Filters        []Filter
	CombinedFilter CombinedFilter // deprecated
	TotalRows      int64
	TotalPages     int

	PaginationRequest *goyave.Request
	Paginator         *database.Paginator
}

type Filter struct {
	Or       bool
	Relation string
	Column   string
	Operator string
	Values   []string
}

type CombinedFilter struct {
	Columns []string
	Value   string
}

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

func (p *Pagination) GetFilters() []Filter {
	return p.Filters
}

func (p *Pagination) AddFilter(f Filter) {
	p.Filters = append(p.Filters, f)
}

func (p *Pagination) MakePaginationRequest() {
	if p.PaginationRequest == nil {
		p.PaginationRequest = &goyave.Request{}
	}

	pgFilters := make([]*filter.Filter, 0)
	pgSorts := make([]*filter.Sort, 0)

	for _, f := range p.Filters {
		field := f.Column
		if f.Relation != "" {
			field = f.Relation + "." + f.Column
		}

		pgFilter := filter.Filter{
			Field:    field,
			Operator: convertOperator(f.Operator),
			Args:     f.Values,
			Or:       f.Or,
		}

		pgFilters = append(pgFilters, &pgFilter)
	}

	pgSort := filter.Sort{
		Field: p.SortColumn,
		Order: filter.SortOrder(p.SortOrder),
	}
	pgSorts = append(pgSorts, &pgSort)

	pgJoins := filter.Join{}

	p.PaginationRequest.Data = map[string]interface{}{
		"filter":   pgFilters,
		"join":     pgJoins,
		"page":     p.Page,
		"per_page": p.Limit,
		"sort":     pgSorts,
	}
}

func (p *Pagination) Fetch(db *gorm.DB, dest interface{}) (*database.Paginator, *gorm.DB) {
	paginator, db := filter.Scope(db, p.PaginationRequest, dest)

	p.Paginator = paginator

	p.Page = paginator.CurrentPage
	p.TotalPages = int(paginator.MaxPage)
	p.TotalRows = paginator.Total
	p.Limit = int(paginator.PageSize)

	return paginator, db
}

func (p *Pagination) Response(ctx context.Context) (out domain.PaginationResponse, err error) {
	if err := serializer.Map(ctx, *p, &out); err != nil {
		return out, err
	}
	out.Filters = make([]domain.FilterResponse, len(p.Filters))
	for i, f := range p.Filters {
		if err := serializer.Map(ctx, f, &out.Filters[i]); err != nil {
			continue
		}
	}

	return out, err
}

func NewPaginationWithFilter(column string, releation string, op string, values []string) *Pagination {
	pg := newDefaultPagination()

	pg.Filters = append(pg.Filters, Filter{
		Column:   helper.ToSnakeCase(strings.Replace(column, "[]", "", -1)),
		Relation: releation,
		Operator: op,
		Values:   values,
		Or:       false,
	})
	pg.MakePaginationRequest()

	return pg
}

func NewPagination(urlParams *url.Values) *Pagination {
	pg := newDefaultPagination()

	if urlParams != nil {
		for key, value := range *urlParams {
			switch key {
			case SORT_COLUMN:
				if value[0] != "" {
					pg.SortColumn = value[0]
				}
			case SORT_ORDER:
				if value[0] != "" {
					pg.SortOrder = value[0]
				}
			case PAGE_NUMBER:
				if value[0] != "" {
					pg.Page, _ = strconv.Atoi(value[0])
				}
			case PAGE_SIZE:
				if value[0] != "" {
					if limitNum, err := strconv.Atoi(value[0]); err == nil {
						pg.Limit = limitNum
					}
				}
			case COMBINED_FILTER: // deprecated
				log.Error(context.TODO(), "DEPRECATED filter scheme. COMBINEND_FILTER")
			case FILTER, FILTER_ARRAY, OR, OR_ARRAY:
				for _, filterValue := range value {
					arr := strings.Split(filterValue, "|")

					columns := strings.Split(arr[0], ",")
					for i, column := range columns {
						releation := ""
						arrColumns := strings.Split(column, ".")
						if len(arrColumns) > 1 {
							releation = strcase.ToCamel(arrColumns[0])
							column = arrColumns[1]
						}

						trimmedStr := strings.Trim(arr[1], "[]")
						values := strings.Split(trimmedStr, ",")

						op := "$cont"
						if len(arr) == 3 {
							op = arr[2]
						}

						or := false
						if i > 0 || key == OR || key == OR_ARRAY {
							or = true
						}

						pg.Filters = append(pg.Filters, Filter{
							Column:   helper.ToSnakeCase(strings.Replace(column, "[]", "", -1)),
							Relation: releation,
							Operator: op,
							Values:   values,
							Or:       or,
						})

					}
				}
			}
		}
	}
	pg.MakePaginationRequest()

	return pg
}

func newDefaultPagination() *Pagination {
	return &Pagination{
		SortColumn: "created_at",
		SortOrder:  "DESC",
		Page:       1,
		Limit:      DEFAULT_LIMIT,
	}
}

func convertOperator(op string) *filter.Operator {
	if _, ok := filter.Operators[op]; ok {
		return filter.Operators[op]
	}
	return filter.Operators["$cont"]
}
