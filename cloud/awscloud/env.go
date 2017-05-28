package awscloud

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
)

func EnvRegion() string {
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-west-2"
	}

	return region
}

// EnvSession returns an AWS session constructed from env variables
func EnvSession(endpoint string) (*session.Session, error) {
	region := EnvRegion()

	cfg := &aws.Config{
		Region: aws.String(region),
	}

	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	s, err := session.NewSession(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "awscloud:env_session:err")
	}

	return s, nil
}
