package awscloud

import "github.com/aws/aws-sdk-go/service/dynamodb"

// DynamoDB instantiates a dynamodb client; endpoint is an optional endpoint useful for testing
func DynamoDB(endpoint string) (*dynamodb.DynamoDB, error) {
	s, err := EnvSession(endpoint)
	if err != nil {
		return nil, err
	}
	return dynamodb.New(s), nil
}
