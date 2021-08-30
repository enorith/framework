package rules

import (
	"fmt"

	"github.com/enorith/http/contracts"
	"gorm.io/gorm"
)

type Unique struct {
	db    *gorm.DB
	table string
	field string
}

func (u *Unique) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	var count int64
	u.db.Table(u.table).Where(fmt.Sprintf("%s = ?", u.field), input).Count(&count)
	if count > 0 {
		return false, false
	}
	return true, false
}

func NewUnique(db *gorm.DB, table, field string) *Unique {
	return &Unique{db: db, table: table, field: field}
}
