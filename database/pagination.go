package database

import (
	"math"

	"github.com/enorith/http/contracts"
	"gorm.io/gorm"
)

type PageParams interface {
	Page() int
	PerPage() int
}

type PageMeta struct {
	Total    int64 `json:"total"`
	PerPage  int   `json:"per_page"`
	Page     int   `json:"page"`
	LastPage int   `json:"last_page"`
	From     int   `json:"from"`
	To       int   `json:"to"`
}

type PageResult[T interface{}] struct {
	Meta PageMeta `json:"meta"`
	Data []T      `json:"data"`
}

type Paginator[T interface{}] struct {
	Params PageParams
}

func (p Paginator[T]) Paginate(tx *gorm.DB) (*PageResult[T], error) {
	var (
		meta    PageMeta
		targets []T
	)
	page := p.Params.Page()
	perPage := p.Params.PerPage()
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPageSize
	}

	meta.Page = page
	meta.PerPage = perPage
	meta.From = perPage*(page-1) + 1
	newTx := tx.Session(&gorm.Session{
		NewDB: true,
	})
	tx.Statement.Dest = targets
	e := newTx.Table("(?) as `aggragate`", tx).Count(&meta.Total).Error
	if e != nil {
		return nil, e
	}
	db := tx.Limit(int(perPage)).Offset(int(perPage * (page - 1))).Find(&targets)
	e = db.Error
	if e != nil {
		return nil, e
	}
	meta.LastPage = int(math.Ceil(float64(meta.Total) / float64(perPage)))
	meta.To = meta.From + int(db.RowsAffected-1)

	return &PageResult[T]{
		Data: targets,
		Meta: meta,
	}, nil
}

func NewPaginator[T interface{}](params PageParams) Paginator[T] {
	return Paginator[T]{
		Params: params,
	}
}

type RequestPageParams struct {
	request contracts.RequestContract
}

func (r RequestPageParams) Page() int {
	p, _ := r.request.GetInt(PageKey)
	return p
}

func (r RequestPageParams) PerPage() int {
	p, _ := r.request.GetInt(PageSizeKey)
	return p
}