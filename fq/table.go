package fq

import "strings"

// DynamoDBTableName returns the fq dynamodb table name for the
// specified env, service, and base table name
func DynamoDBTableName(env, service, table string) string {
	return strings.Join([]string{env, service, table}, "-")
}
