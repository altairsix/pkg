package types

import (
	"database/sql/driver"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	ZeroKey = Key("")
)

type Key string

func (k Key) String() string {
	return string(k)
}

func (k Key) AttributeValue() *dynamodb.AttributeValue {
	if k == "" {
		return nil
	}

	return &dynamodb.AttributeValue{
		S: aws.String(k.String()),
	}
}

func (k Key) IsEmpty() bool {
	return k == ZeroKey
}

func (k Key) IsPresent() bool {
	return !k.IsEmpty()
}

func (k Key) Value() (driver.Value, error) {
	return k.String(), nil
}

func (k *Key) Scan(src interface{}) error {
	if v, ok := src.([]byte); ok {
		*k = Key(v)
	}

	return nil
}
