package av

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func String(v string) *dynamodb.AttributeValue {
	if v == "" {
		return nil
	}

	return &dynamodb.AttributeValue{
		S: aws.String(v),
	}
}

func Int(v int) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{
		N: aws.String(strconv.Itoa(v)),
	}
}

func Int64(v int64) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{
		N: aws.String(strconv.FormatInt(v, 10)),
	}
}

func Struct(v interface{}) (*dynamodb.AttributeValue, error) {
	if v == nil {
		return nil, nil
	}

	return dynamodbattribute.Marshal(v)
}
