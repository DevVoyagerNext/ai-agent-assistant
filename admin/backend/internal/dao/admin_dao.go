package dao

import "gorm.io/gorm"

type AdminDao struct {
	db *gorm.DB
}

func NewAdminDao(db *gorm.DB) *AdminDao {
	return &AdminDao{db: db}
}

func (d *AdminDao) Transaction(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}
