package gormx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/altairsix/pkg/dbase"
	"github.com/jinzhu/gorm"
)

// Accessor provides an adapter between dbase.Accessor and mysqlstore.Accessor
type Accessor struct {
	Target dbase.Accessor
}

// DB provides the shape required by mysqlstore.Accessor
type DB interface {
	// Exec is implemented by *sql.DB and *sql.Tx
	Exec(query string, args ...interface{}) (sql.Result, error)
	// PrepareContext is implemented by *sql.DB and *sql.Tx
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	// Query is implemented by *sql.DB and *sql.Tx
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type gormDB struct {
	db *gorm.DB
}

// Exec is implemented by *sql.DB and *sql.Tx
func (g *gormDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return g.db.DB().Exec(query, args...)
}

// PrepareContext is implemented by *sql.DB and *sql.Tx
func (g *gormDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return g.db.DB().PrepareContext(ctx, query)
}

// Query is implemented by *sql.DB and *sql.Tx
func (g *gormDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return g.db.DB().Query(query, args...)
}

// Open open a db connection
func (a *Accessor) Open(ctx context.Context) (DB, error) {
	db, err := a.Target.Open()
	if err != nil {
		return nil, err
	}

	return &gormDB{db: db}, nil
}

// Close the db connection
func (a *Accessor) Close(db DB) error {
	v, ok := db.(*gormDB)
	if !ok {
		return fmt.Errorf("unable to close unexpected type, %#v", db)
	}

	return a.Target.Close(v.db)
}
