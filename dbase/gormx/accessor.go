package gormx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/altairsix/eventsource/mysqlstore"
	"github.com/altairsix/pkg/dbase"
	"github.com/jinzhu/gorm"
)

// Accessor provides an adapter between dbase.Accessor and mysqlstore.Accessor
type Accessor struct {
	Target dbase.Accessor
}

type gormDB struct {
	db *gorm.DB
}

// Exec is implemented by *sql.DB and *sql.Tx
func (g *gormDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return g.db.CommonDB().Exec(query, args...)
}

// PrepareContext is implemented by *sql.DB and *sql.Tx
func (g *gormDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	db := g.db.CommonDB()

	target, ok := db.(mysqlstore.DB)
	if !ok {
		return nil, fmt.Errorf("unable to handle common db")
	}

	return target.PrepareContext(ctx, query)
}

// Query is implemented by *sql.DB and *sql.Tx
func (g *gormDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return g.db.CommonDB().Query(query, args...)
}

// Open open a db connection
func (a *Accessor) Open(ctx context.Context) (mysqlstore.DB, error) {
	db, err := a.Target.Open()
	if err != nil {
		return nil, err
	}

	return &gormDB{db: db}, nil
}

// Close the db connection
func (a *Accessor) Close(db mysqlstore.DB) error {
	v, ok := db.(*gormDB)
	if !ok {
		return fmt.Errorf("unable to close unexpected type, %#v", db)
	}

	return a.Target.Close(v.db)
}
