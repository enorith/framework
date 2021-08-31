package database

import (
	"github.com/enorith/gormdb"
	"gorm.io/gorm"
)

type Model struct {
	*gorm.DB  `gorm:"-" json:"-"`
	Paginator *gormdb.Paginator `gorm:"-" json:"-"`
	Dest      interface{}       `gorm:"-" json:"-"`
	DestSlice interface{}       `gorm:"-" json:"-"`
}

func (m Model) Paginate() (map[string]interface{}, error) {

	return m.Paginator.Paginate(m.DB, m.DestSlice)
}
