package repository

import (
	"fmt"
	"net/url"
	"strconv"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type Repository struct {
	db *gorm.DB
}

type Pagination struct {
	Limit      int    `json:"limit,omitempty;query:limit"`
	Page       int    `json:"page,omitempty;query:page"`
	Sort       string `json:"sort,omitempty;query:sort"`
	TotalRows  int64  `json:"totalRows"`
	TotalPages int    `json:"totalPages"`
}

var DEFAULT_LIMIT = 10

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

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func nilUuid() uuid.UUID {
	nilId, _ := uuid.Parse("")
	return nilId
}

func NewPagination(urlParams *url.Values) *Pagination {
	var pagination Pagination

	sortingColumn := urlParams.Get("sortingColumn")
	if sortingColumn == "" {
		sortingColumn = "created_at"
	}
	order := urlParams.Get("order")
	if order == "" {
		order = "asc"
	}
	pagination.Sort = fmt.Sprintf("%s %s", sortingColumn, order)

	page := urlParams.Get("page")
	if page == "" {
		pagination.Page = 1
	} else {
		pagination.Page, _ = strconv.Atoi(page)
	}

	limit := urlParams.Get("limit")
	if limit == "" {
		pagination.Limit = DEFAULT_LIMIT
	} else {
		limitNum, err := strconv.Atoi(limit)
		if err == nil {
			pagination.Limit = limitNum
		}
	}

	return &pagination
}
