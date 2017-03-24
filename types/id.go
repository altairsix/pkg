package types

import (
	"database/sql/driver"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

// -- ID -------------------------------------------------

const (
	ZeroID = ID(0)
)

type ID int64

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id ID) String36() string {
	return strings.ToUpper(strconv.FormatInt(int64(id), 36))
}

func (id ID) IsZero() bool {
	return id == ZeroID
}

func (id ID) IsPresent() bool {
	return !id.IsZero()
}

func (id ID) Key() Key {
	if id == 0 {
		return ZeroKey
	}

	return Key(id.String36())
}

func (id ID) Value() (driver.Value, error) {
	return int64(id), nil
}

func (id *ID) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case []byte:
		x, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}

		*id = ID(x)

	case int64:
		*id = ID(v)
	}

	return nil
}

func (id ID) AttributeValue() *dynamodb.AttributeValue {
	if id == 0 {
		return nil
	}

	return &dynamodb.AttributeValue{
		N: aws.String(id.String()),
	}
}

func NewID(in string) (ID, error) {
	v, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return ZeroID, errors.Wrap(err, "types:new_id:err")
	}

	return ID(v), nil
}
