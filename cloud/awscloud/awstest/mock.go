package awstest

import "github.com/aws/aws-sdk-go/service/sns"

type MockSNS struct {
	PublishInput *sns.PublishInput
	Err          error
}

func (m *MockSNS) Publish(in *sns.PublishInput) (*sns.PublishOutput, error) {
	m.PublishInput = in
	return &sns.PublishOutput{}, m.Err
}
