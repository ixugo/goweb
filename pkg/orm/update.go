package orm

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// QueryOption ..
type QueryOption func(*gorm.DB) *gorm.DB

// Where 查询条件
func Where(query interface{}, args ...interface{}) QueryOption {
	return func(d *gorm.DB) *gorm.DB {
		return d.Where(query, args...)
	}
}

// OrderBy 排序条件
func OrderBy(value interface{}) QueryOption {
	return func(d *gorm.DB) *gorm.DB {
		return d.Order(value)
	}
}

// Universal 通用增删改查
type Universal[T any] interface {
	First(*T, ...QueryOption) error
	Update(*T, func(*T), ...QueryOption) error
	Delete(*T, ...QueryOption) error
	Create(*T) error
	Find(out *[]*T, p Pager, opts ...QueryOption) (int64, error)
}

type Type[T any] struct {
	db *gorm.DB
}

func NewUniversal[T any](db *gorm.DB) Universal[T] {
	return &Type[T]{db: db}
}

// First 通用查询
func (t *Type[T]) First(out *T, opts ...QueryOption) error {
	return First(t.db, out, opts...)
}

func First(db *gorm.DB, out any, opts ...QueryOption) error {
	if len(opts) == 0 {
		return fmt.Errorf("where is empty")
	}
	for _, opt := range opts {
		db = opt(db)
	}
	return db.First(out).Error
}

// Update 通用更新
func (t Type[T]) Update(model *T, changeFn func(*T), opts ...QueryOption) error {
	return Update(t.db, model, changeFn, opts...)
}

func Update[T any](db *gorm.DB, model *T, changeFn func(*T), opts ...QueryOption) error {
	if len(opts) == 0 {
		return fmt.Errorf("where is empty")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		db := tx.Clauses(clause.Locking{Strength: "UPDATE"})
		for _, opt := range opts {
			db = opt(db)
		}
		if err := db.First(model).Error; err != nil {
			return err
		}
		changeFn(model)
		return tx.Save(model).Error
	})
}

// Delete 通用删除
func (t Type[T]) Delete(model *T, opts ...QueryOption) error {
	return Delete(t.db, model, opts...)
}

func Delete(db *gorm.DB, model any, opts ...QueryOption) error {
	if len(opts) == 0 {
		return fmt.Errorf("where is empty")
	}
	db = db.Clauses(clause.Returning{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(model).Error
}

func (t Type[T]) Create(model *T) error {
	return t.db.Create(model).Error
}

type Pager interface {
	Limit() int
	Offset() int
}

func (t Type[T]) Find(out *[]*T, p Pager, opts ...QueryOption) (int64, error) {
	return Find(t.db, out, p, opts...)
}

func Find[T any](db *gorm.DB, out *[]*T, p Pager, opts ...QueryOption) (int64, error) {
	db = db.Model(new(T))
	for _, opt := range opts {
		db = opt(db)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil || total <= 0 {
		return total, err
	}
	return total, db.Limit(p.Limit()).Offset(p.Offset()).Find(out).Error
}