package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Db struct {
	Path          string
	Disabled      bool
	SkipMigration bool
}

func NewDb() *Db {
	return &Db{}
}

func (db *Db) Get() (*gorm.DB, error) {
	if db.Disabled {
		return nil, nil
	}

	conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}

	if !db.SkipMigration {
		conn.AutoMigrate(&Event{}, &EventCustomTitle{}, &EventCustomDescription{}, &EventCustomDate{}, &EventCustomURL{})
	}

	return conn, nil
}
