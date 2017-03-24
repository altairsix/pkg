package queue_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/pkg/cloud/awscloud/awstest"
	"github.com/altairsix/pkg/cloud/awscloud/queue"
	"github.com/altairsix/pkg/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background(awstest.Env))
	go func() {
		time.Sleep(time.Second * 8)
	}()

	sent := int64(20)
	received := int64(0)
	body := `{"hello":"world"}`

	// -- obtain a reference to the queue ------------------------------------

	out, err := awstest.SQS.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("queue-test"),
	})
	assert.Nil(t, err)
	queueUrl := out.QueueUrl

	// -- populate the queue with messages -----------------------------------

	go func() {
		for i := 1; i <= int(sent); i++ {
			_, err = awstest.SQS.SendMessage(&sqs.SendMessageInput{
				MessageBody: aws.String(body),
				QueueUrl:    queueUrl,
			})
			assert.Nil(t, err)
		}
	}()

	// -- receive messages ---------------------------------------------------

	fn := func(k context.Kontext, m *sqs.Message) error {
		if m.Body != nil {
			assert.Equal(t, body, *m.Body)
			if v := atomic.AddInt64(&received, 1); v >= sent {
				cancel() // stop the queue
			}
		}

		return nil
	}

	done := queue.Start(ctx, awstest.SQS, queueUrl, fn, queue.Workers(5))
	<-done
}
