package dbase

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	Key = "_db"
)

type Config struct {
	Username    string
	Password    string
	Hostname    string
	Port        string
	Database    string
	TxIsolation string
}

func ConnectString(cfg Config) string {
	isolation := cfg.TxIsolation
	if isolation == "" {
		isolation = "READ-COMMITTED"
	}

	return fmt.Sprintf(`%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local&tx_isolation="%v"`,
		cfg.Username,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
		isolation,
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
	Tx(ctx context.Context, callback func(ctx context.Context, db *gorm.DB) error) error
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

func (fn OpenFunc) Tx(ctx context.Context, callback func(ctx context.Context, db *gorm.DB) error) error {
	db, err := fn.Open()
	if err != nil {
		return err
	}

	db = db.Begin()
	ctx = context.WithValue(ctx, Key, db)

	err = callback(ctx, db)
	if err != nil {
		db.Rollback()
		return err
	}

	db.Commit()
	return nil
}

// FromContext retrieves a db instance from a context
func FromContext(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(Key)
	if v == nil {
		return nil, false
	}

	db, ok := v.(*gorm.DB)
	return db, ok
}

func NewOpenFunc(connectString string) OpenFunc {
	return func() (*gorm.DB, error) {
		return Open(connectString)
	}
}

type Mock struct {
	DB            *gorm.DB
	OpenCount     int
	CloseCount    int
	BeginCount    int
	CommitCount   int
	RollbackCount int
	TxCount       int
}

func (m *Mock) Open() (*gorm.DB, error) {
	m.OpenCount++
	return m.DB, nil
}

func (m *Mock) Close(db *gorm.DB) error {
	m.CloseCount++
	return nil
}

func (m *Mock) Begin(db *gorm.DB) *gorm.DB {
	m.BeginCount++
	return m.DB
}

func (m *Mock) Commit(db *gorm.DB) *gorm.DB {
	m.CommitCount++
	return m.DB
}

func (m *Mock) Rollback(db *gorm.DB) *gorm.DB {
	m.RollbackCount++
	return m.DB
}

func (m *Mock) Tx(ctx context.Context, callback func(ctx context.Context, db *gorm.DB) error) error {
	m.TxCount++
	return callback(ctx, m.DB)
}
