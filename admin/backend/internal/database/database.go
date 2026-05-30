package database

import (
	"admin/backend/internal/config"
	"admin/backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.Mysql.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Mysql.MaxOpenConns)

	if err := db.AutoMigrate(&model.AuditLog{}); err != nil {
		return nil, err
	}

	return db, nil
}
