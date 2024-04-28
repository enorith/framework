package database

import "gorm.io/gorm"

type Builder[T interface{}] struct {
	*gorm.DB
	Paginator Paginator[T]
}

func (b *Builder[T]) Query(fn func(*gorm.DB) *gorm.DB) *Builder[T] {
	var model T
	return NewBuilder(fn(b.DB.Model(&model)), b.Paginator)
}

func (b *Builder[T]) Get() (result []T, err error) {
	var model T
	err = b.DB.Model(model).Find(&result).Error
	return
}

func (b *Builder[T]) First(conds ...interface{}) (result T, err error) {
	var model T
	err = b.DB.Model(model).First(&result, conds...).Error
	return
}
func (b *Builder[T]) Paginate(opts ...PaginateOptions) (*PageResult[T], error) {
	return b.Paginator.Paginate(b.DB, opts...)
}

func NewBuilder[T interface{}](tx *gorm.DB, paginator Paginator[T]) *Builder[T] {
	return &Builder[T]{DB: tx, Paginator: paginator}
}
