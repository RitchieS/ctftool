package storage

import (
	"fmt"
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

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbCwdPath := path.Join(cwd, db.Path)
	dbHomePath := path.Join(homeDir, ".config", "ctftool")

	if _, err := os.Stat(dbCwdPath); err == nil {
		db.Path = dbCwdPath
	} else if _, err := os.Stat(dbHomePath); err == nil {
		dbHomePath = path.Join(dbHomePath, db.Path)
		if _, err := os.Stat(dbHomePath); err == nil {
			db.Path = dbHomePath
		}
	} else {
		err := os.MkdirAll(dbHomePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
		db.Path = path.Join(dbHomePath, db.Path)
	}

	fmt.Fprintln(os.Stderr, "Connecting to database:", db.Path)

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
