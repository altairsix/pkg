package dbase

import (
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Config struct {
	Username string
	Password string
	Hostname string
	Port     string
	Database string
}

func ConnectString(cfg Config) string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
	)
}

func Open(connectString string) (*gorm.DB, error) {
	return gorm.Open("mysql", connectString)
}

func OpenNative(connectString string) (*sql.DB, error) {
	return sql.Open("mysql", connectString)
}

type Accessor interface {
	Open() (*gorm.DB, error)
	Close(*gorm.DB) error
	Begin(*gorm.DB) *gorm.DB
	Commit(*gorm.DB) *gorm.DB
	Rollback(*gorm.DB) *gorm.DB
}

type OpenFunc func() (*gorm.DB, error)

func (fn OpenFunc) Open() (*gorm.DB, error) {
	return fn()
}

func (fn OpenFunc) Close(db *gorm.DB) error {
	return db.Close()
}

func (fn OpenFunc) Begin(db *gorm.DB) *gorm.DB {
	return db.Begin()
}

func (fn OpenFunc) Commit(db *gorm.DB) *gorm.DB {
	return db.Commit()
}

func (fn OpenFunc) Rollback(db *gorm.DB) *gorm.DB {
	return db.Rollback()
}

func NewOpenFunc(connectString string) OpenFunc {
	return func() (*gorm.DB, error) {
		return Open(connectString)
	}
}

type Mock struct {
	DB *gorm.DB
}

func (m Mock) Open() (*gorm.DB, error) {
	return m.DB, nil
}

func (m Mock) Close(db *gorm.DB) error {
	return nil
}

func (m Mock) Begin(db *gorm.DB) *gorm.DB {
	return m.DB
}

func (m Mock) Commit(db *gorm.DB) *gorm.DB {
	return m.DB
}

func (m Mock) Rollback(db *gorm.DB) *gorm.DB {
	return m.DB
}
