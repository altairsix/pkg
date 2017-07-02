package checkpoint

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

// TableName provides name of checkpoints table for a given environment
func TableName(env string) string {
	return env + "-checkpoints"
}

// MakeCreateTableInput creates the create table description
func MakeCreateTableInput(env string, readCapacity, writeCapacity int64) *dynamodb.CreateTableInput {
	return &dynamodb.CreateTableInput{
		TableName: aws.String(TableName(env)),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("key"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("key"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(readCapacity),
			WriteCapacityUnits: aws.Int64(writeCapacity),
		},
	}
}

// CreateTable creates tables with specified capacity
func CreateTable(api *dynamodb.DynamoDB, env string, readCapacity, writeCapacity int64) error {
	tableName := TableName(env)
	fmt.Printf("creating table, %v ... ", tableName)

	input := MakeCreateTableInput(env, readCapacity, writeCapacity)
	_, err := api.CreateTable(input)
	if err != nil {
		if v := err.(awserr.Error); v.Code() == dynamodb.ErrCodeResourceInUseException {
			fmt.Println("already exists, skipping")
			return nil
		}
		return errors.Wrapf(err, "unable to create table, %v", tableName)
	}

	fmt.Println("ok")
	return nil
}
