package rules

import (
	"fmt"

	"github.com/enorith/http/contracts"
	"gorm.io/gorm"
)

type Unique struct {
	db            *gorm.DB
	table         string
	field         string
	inputFormater func(input contracts.InputValue) interface{}
}

func (u *Unique) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	var count int64
	u.db.Table(u.table).Where(fmt.Sprintf("%s = ?", u.field), u.inputFormater(input)).Count(&count)
	if count > 0 {
		return false, false
	}
	return true, false
}

func (u *Unique) FormatInput(f func(input contracts.InputValue) interface{}) *Unique {
	u.inputFormater = f
	return u
}

func NewUnique(db *gorm.DB, table, field string) *Unique {
	return &Unique{db: db, table: table, field: field, inputFormater: func(input contracts.InputValue) interface{} {
		return input.GetString()
	}}
}
