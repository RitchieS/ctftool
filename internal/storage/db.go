package storage

import (
	"os"
	"path"

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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dbCwdPath := path.Join(homeDir, db.Path)
	dbHomePath := path.Join(homeDir, ".config", "ctftool", db.Path)
	dbHomeCwdPath := path.Join(cwd, db.Path)

	if _, err := os.Stat(dbCwdPath); err == nil {
		db.Path = dbCwdPath
	} else if _, err := os.Stat(dbHomePath); err == nil {
		db.Path = dbHomePath
	} else if _, err := os.Stat(dbHomeCwdPath); err == nil {
		db.Path = dbHomeCwdPath
	} else {
		db.Path = dbHomePath
	}

	conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}

	if !db.SkipMigration {
		err := conn.AutoMigrate(&Event{}, &EventCustomTitle{}, &EventCustomDescription{}, &EventCustomDate{}, &EventCustomURL{})
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}
