package checkpoint

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// CP is a dynamodb backed implementation of publisher.Checkpointer
type CP struct {
	tableName string
	api       *dynamodb.DynamoDB
}

// New constructs a new dynamodb backed CP that implements publisher.Checkpointer
func New(env string, api *dynamodb.DynamoDB) *CP {
	return &CP{
		tableName: fmt.Sprintf(TableName(env)),
		api:       api,
	}
}

// Save the offset for the specified key to the data store
func (c *CP) Save(ctx context.Context, key string, offset uint64) error {
	offsetStr := aws.String(strconv.FormatUint(offset, 10))
	_, err := c.api.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"key":    {S: aws.String(key)},
			"offset": {N: offsetStr},
		},
		ConditionExpression: aws.String("attribute_not_exists(#key) or (attribute_exists(#key) and #offset <= :offset)"),
		ExpressionAttributeNames: map[string]*string{
			"#key":    aws.String("key"),
			"#offset": aws.String("offset"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":offset": {N: offsetStr},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Load the offset for the specified key from the data store
func (c *CP) Load(ctx context.Context, key string) (uint64, error) {
	out, err := c.api.GetItem(&dynamodb.GetItemInput{
		TableName:      aws.String(c.tableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {S: aws.String(key)},
		},
	})
	if err != nil {
		return 0, err
	}

	if len(out.Item) == 0 || out.Item["offset"] == nil {
		return 0, nil
	}

	offset, err := strconv.ParseUint(*out.Item["offset"].N, 10, 64)
	if err != nil {
		return 0, err
	}

	return offset, nil
}
