package trigger

import (
	"testing"

	"github.com/apex/go-apex/sns"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

type Mock struct {
	Err        error
	ReceiveIn  *sqs.ReceiveMessageInput
	ReceiveOut *sqs.ReceiveMessageOutput
	DeleteIn   *sqs.DeleteMessageBatchInput
	DeleteOut  *sqs.DeleteMessageBatchOutput
}

func (m *Mock) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	m.ReceiveIn = input
	return m.ReceiveOut, m.Err
}

func (m *Mock) DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error) {
	m.DeleteIn = input
	return m.DeleteOut, m.Err
}

func TestReceiveMessage(t *testing.T) {
	m := &Mock{
		ReceiveOut: &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body: aws.String("body"),
				},
			},
		},
	}
	queryUrl := aws.String("blah")
	ch, err := receiveMessages(m, queryUrl, &sns.Event{})
	assert.Nil(t, err)
	assert.NotNil(t, ch)
	assert.Equal(t, int64(1), *m.ReceiveIn.WaitTimeSeconds)

	message := <-ch
	assert.Equal(t, m.ReceiveOut.Messages[0], message)

	_, ok := <-ch
	assert.False(t, ok, "expected chan to be closed")
}

func TestReceiveMessageFromCloudWatch(t *testing.T) {
	m := &Mock{
		ReceiveOut: &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body: aws.String("body"),
				},
			},
		},
	}
	queryUrl := aws.String("blah")

	event := &sns.Event{
		Records: []*sns.Record{
			{},
		},
	}
	event.Records[0].SNS.Message = `{"source":"aws.events"}`

	_, err := receiveMessages(m, queryUrl, event)
	assert.Nil(t, err)
	assert.Zero(t, *m.ReceiveIn.WaitTimeSeconds)
}

func TestProcessMessages(t *testing.T) {
	a := &sqs.Message{Body: aws.String("a")}
	b := &sqs.Message{Body: aws.String("b")}
	c := &sqs.Message{Body: aws.String("c")}
	d := &sqs.Message{Body: aws.String("d")}

	ch := make(chan *sqs.Message, 4)
	ch <- a
	ch <- b
	ch <- c
	ch <- d
	close(ch)

	calls := 0
	del := processMessages(ch, 1, HandlerFunc(func(string) error {
		calls++
		return nil
	}))

	assert.Equal(t, a, <-del)
	assert.Equal(t, b, <-del)
	assert.Equal(t, c, <-del)
	assert.Equal(t, d, <-del)
	assert.Equal(t, 4, calls)

	_, ok := <-del
	assert.False(t, ok)
}

func TestDeleteMessages(t *testing.T) {
	a := &sqs.Message{Body: aws.String("a")}
	b := &sqs.Message{Body: aws.String("b")}

	del := make(chan *sqs.Message, 4)
	del <- a
	del <- b
	close(del)

	m := &Mock{}
	queryUrl := aws.String("blah")
	deleteMessages(m, queryUrl, del)

	assert.Len(t, m.DeleteIn.Entries, 2)
}

func TestHandle(t *testing.T) {
	m := &Mock{
		ReceiveOut: &sqs.ReceiveMessageOutput{
			Messages: []*sqs.Message{
				{
					Body: aws.String("body"),
				},
			},
		},
	}
	queryUrl := aws.String("blah")

	calls := 0
	fn := HandleFunc(m, queryUrl, 1, func(in string) error {
		calls++
		return nil
	})
	err := fn(&sns.Event{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, calls)
	assert.NotNil(t, m.ReceiveIn)
	assert.NotNil(t, m.DeleteIn)
	assert.Len(t, m.DeleteIn.Entries, 1)
}
