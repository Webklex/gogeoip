package paginator

import (
	"gorm.io/gorm"
)

type GormPaginator struct {
	*DefaultPaginator
}

func NewGormPaginator(paginator *DefaultPaginator) *GormPaginator {
	return &GormPaginator{
		DefaultPaginator: paginator,
	}
}

func (p *GormPaginator) Paginate(db *gorm.DB, models interface{}) error {
	tx := db.Scopes(p.paginate(models, db)).Find(models)
	p.Rows = models
	return tx.Error
}

func (p *GormPaginator) paginate(value interface{}, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)
	p.TotalRows = totalRows
	p.TotalPages = 1
	if p.TotalRows > int64(p.Limit) && totalRows > 0 {
		p.TotalPages = totalRows / int64(p.GetLimit())
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.GetOffset()).Limit(p.GetLimit()).Order(p.GetOrder())
	}
}

func (p *GormPaginator) GetSortKey() string {
	if p.SortKey == "" {
		p.SortKey = "Id"
	}
	return p.SortKey
}

func (p *GormPaginator) GetOrder() string {
	return p.GetSortKey() + " " + p.GetSortDirection()
}
