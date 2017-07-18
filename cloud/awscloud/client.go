package awscloud

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDB instantiates a dynamodb client; endpoint is an optional endpoint useful for testing
func DynamoDB(endpoint string) (*dynamodb.DynamoDB, error) {
	s, err := EnvSession(endpoint)
	if err != nil {
		return nil, err
	}
	return dynamodb.New(s), nil
}

// CreateTable creates a new table using parameters specified
func CreateTable(w io.Writer, dynamodbAPI *dynamodb.DynamoDB, input *dynamodb.CreateTableInput) error {
	fmt.Fprintf(w, "creating table, %v ... ", *input.TableName)

	out, err := dynamodbAPI.CreateTable(input)
	if err != nil {
		if v, ok := err.(awserr.Error); ok && v.Code() == dynamodb.ErrCodeResourceInUseException {
			fmt.Fprintln(w, "exists, skipping")
			return nil
		}

		fmt.Fprintf(w, "failed, %v\n", err)
		return nil
	}

	fmt.Fprintln(w, *out.TableDescription.TableArn)
	return nil
}

// DeleteTable deletes a table using parameters specified
func DeleteTable(w io.Writer, dynamodbAPI *dynamodb.DynamoDB, tableName string) error {
	fmt.Fprintf(w, "deleting table, %v ... ", tableName)
	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	if _, err := dynamodbAPI.DeleteTable(input); err != nil {
		if v, ok := err.(awserr.Error); ok {
			switch v.Code() {
			case dynamodb.ErrCodeResourceInUseException:
				fmt.Fprintln(w, "in use, try again later")
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Fprintln(w, "not found, ok")
			}
			return nil
		}

		fmt.Fprintf(w, "failed, %v\n", err)
		return nil
	}

	fmt.Fprintln(w, "ok")
	return nil
}
