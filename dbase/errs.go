package dbase

import "github.com/go-sql-driver/mysql"

const (
	MySQLDuplicateEntryNum = 1062
)

func IsDuplicateErr(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == MySQLDuplicateEntryNum {
			return true
		}
	}

	return false
}
