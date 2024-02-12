package rules

import (
	"fmt"

	"github.com/enorith/gormdb"
	"github.com/enorith/http/contracts"
	"gorm.io/gorm"
)

type Unique struct {
	db            *gorm.DB
	table         string
	field         string
	inputFormater func(input contracts.InputValue) interface{}
	scopes        []func(*gorm.DB) *gorm.DB
	connection    string
}

func (u *Unique) RoleName() string {
	return "unique"
}

func (u *Unique) Passes(input contracts.InputValue) (success bool, skipAll bool) {
	var count int64
	if u.db == nil {
		var e error
		if u.connection != "" {
			u.db, e = gormdb.DefaultManager.GetConnection(u.connection)
		} else {
			u.db, e = gormdb.DefaultManager.GetConnection()
		}
		if e != nil {
			return false, false
		}
	}

	builder := u.db.Table(u.table).Scopes(u.scopes...)
	builder.Where(fmt.Sprintf("%s = ?", u.field), u.inputFormater(input)).Count(&count)

	if count > 0 {
		return false, false
	}
	return true, false
}

func (u *Unique) FormatInput(f func(input contracts.InputValue) interface{}) *Unique {
	u.inputFormater = f
	return u
}

func (u *Unique) Scope(s func(*gorm.DB) *gorm.DB) *Unique {
	u.scopes = append(u.scopes, s)

	return u
}

func (u *Unique) Ignore(id interface{}, idField ...string) *Unique {
	f := "id"
	if len(idField) > 0 {
		f = idField[0]
	}

	return u.Scope(func(d *gorm.DB) *gorm.DB {
		return d.Where(fmt.Sprintf("%s != ?", f), id)
	})
}

func NewUnique(db *gorm.DB, table, field string) *Unique {
	return &Unique{db: db, table: table, field: field, inputFormater: func(input contracts.InputValue) interface{} {
		return input.GetString()
	}, scopes: make([]func(*gorm.DB) *gorm.DB, 0)}
}

func UniqueDefault(table, field string) *Unique {
	return NewUnique(nil, table, field)
}
