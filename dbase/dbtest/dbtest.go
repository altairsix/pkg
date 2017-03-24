package dbtest

import (
	"database/sql"
	"log"
	"os"

	"github.com/altairsix/pkg/dbase"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var connectString = dbase.ConnectString(dbase.Config{
	Username: getOrElse("DB_USERNAME", "altairsix"),
	Password: getOrElse("DB_PASSWORD", "password"),
	Hostname: getOrElse("DB_HOSTNAME", "127.0.0.1"),
	Port:     getOrElse("DB_PORT", "3306"),
	Database: getOrElse("DB_DATABASE", "altairsix"),
})

func getOrElse(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}

	return v
}

func Open() (*gorm.DB, error) {
	return dbase.Open(connectString)
}

func OpenNative() (*sql.DB, error) {
	return dbase.OpenNative(connectString)
}

func MustOpen() *gorm.DB {
	db, err := dbase.Open(connectString)
	if err != nil {
		log.Fatalln(errors.Wrap(err, "dbtest:must_open:err"))
	}

	return db
}

func Do(fn func(db *gorm.DB)) {
	db := MustOpen()
	defer db.Close()

	fn(db)
}

func WithRollback(fn func(db *gorm.DB)) {
	db := MustOpen()
	defer db.Close()

	tx := db.Begin()
	defer tx.Rollback()

	fn(tx)
}
