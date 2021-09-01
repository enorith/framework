package database

import (
	"github.com/enorith/gormdb"
	"gorm.io/gorm"
)

type Model struct {
	*gorm.DB  `gorm:"-" json:"-" msgpack:"-"`
	Paginator *gormdb.Paginator `gorm:"-" json:"-"  msgpack:"-"`
	Dest      interface{}       `gorm:"-" json:"-"  msgpack:"-"`
	DestSlice interface{}       `gorm:"-" json:"-"  msgpack:"-"`
}

func (m Model) Paginate() (map[string]interface{}, error) {

	return m.Paginator.Paginate(m.DB, m.DestSlice)
}
