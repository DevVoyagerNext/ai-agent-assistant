package initialize

import (
	"backend/global"
	"backend/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	m := global.GVA_CONFIG.Mysql
	if m.Dbname == "" {
		return nil
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(),
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{}); err != nil {
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		// 自动迁移表
		err = db.AutoMigrate(
			&model.User{},
			&model.Subject{},
			&model.KnowledgeNode{},
			&model.KnowledgeContent{},
			&model.UserStudyNote{},
			&model.UserStudyStatus{},
			&model.Image{},
			&model.NoteShare{},
			&model.UserNodeDifficulty{},
			&model.UserActivityLog{},
			&model.UserDailyActionStat{},
			&model.UserFollow{},
			&model.UserPrivateNote{},
		)
		if err != nil {
			return nil
		}
		return db
	}
}
