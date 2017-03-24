package awstest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	Region = "us-west-2"
	Env    = "local"
)

var (
	DynamoDB *dynamodb.DynamoDB
	SNS      *sns.SNS
	SQS      *sqs.SQS
)

func init() {
	dir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalln(err)
	}

	env := map[string]string{}
	for i := 0; i < 5; i++ {
		dir = filepath.Join(dir, "..")
		filename := filepath.Join(dir, "test.env")
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}

		err = json.Unmarshal(data, &env)
		if err != nil {
			log.Fatalln(err)
		}
		break
	}

	for k, v := range env {
		os.Setenv(k, v)
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-west-2"
	}

	cfg := &aws.Config{Region: aws.String(region)}
	s, err := session.NewSession(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	DynamoDB = dynamodb.New(s)
	SNS = sns.New(s)
	SQS = sqs.New(s)
}
