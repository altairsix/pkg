package dbase

import "github.com/go-sql-driver/mysql"

const (
	MySQLDuplicateKeyNum   = 1061
	MySQLDuplicateEntryNum = 1062
)

func IsDuplicateKey(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == MySQLDuplicateKeyNum {
			return true
		}
	}

	return false
}

func IsDuplicateEntry(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == MySQLDuplicateEntryNum {
			return true
		}
	}

	return false
}

func IsDuplicateErr(err error) bool {
	return IsDuplicateEntry(err)
}
