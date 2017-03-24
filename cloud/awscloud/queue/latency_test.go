package queue_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync/atomic"
	"testing"
	"time"

	"github.com/altairsix/pkg/cloud/awscloud/awstest"
	"github.com/altairsix/pkg/cloud/awscloud/queue"
	"github.com/altairsix/pkg/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/montanaflynn/stats"
	"github.com/stretchr/testify/assert"
)

const (
	Name = "latency-test"
)

type Event struct {
	ID      int   `json:",string"`
	Time    int64 `json:",string"`
	Message string
}

func TestLatency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background(awstest.Env))

	queueUrl, err := queue.Subscribe(ioutil.Discard, awstest.SNS, awstest.SQS, Name, Name)
	if !assert.Nil(t, err) {
		return
	}

	workers := 2
	sent := int64(8)
	received := int64(0)

	results := make(chan float64, int(sent))

	fn := func(k context.Kontext, message *sqs.Message) error {
		v := atomic.AddInt64(&received, 1)

		if v > sent {
			return nil
		}

		now := time.Now().UnixNano()
		outer := &Event{}
		err := json.Unmarshal([]byte(*message.Body), outer)
		assert.Nil(t, err)

		event := &Event{}
		err = json.Unmarshal([]byte(outer.Message), event)
		assert.Nil(t, err)

		elapsed := float64(now-event.Time) / 1000000
		fmt.Printf("%.2f\n", elapsed)
		results <- elapsed

		if v == sent {
			cancel()
			close(results)
		}

		return nil
	}
	done := queue.Start(ctx, awstest.SQS, queueUrl, fn, queue.Workers(workers))

	topic, err := awstest.SNS.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String(Name),
	})
	assert.Nil(t, err)

	for id := int64(0); id < sent; id++ {
		_, err := awstest.SNS.Publish(&sns.PublishInput{
			TopicArn: topic.TopicArn,
			Message:  aws.String(fmt.Sprintf(`{"ID": "%v", "Time": "%v"}`, id, time.Now().UnixNano())),
		})
		assert.Nil(t, err)
		//_, err := awstest.SQS.SendMessage(&sqs.SendMessageInput{
		//	QueueUrl:    queueUrl,
		//	MessageBody: aws.String(fmt.Sprintf(`{"ID": "%v", "Time": "%v"}`, id, time.Now().UnixNano())),
		//})
		//assert.Nil(t, err)
	}

	data := make([]float64, 0, int(sent))
	for v := range results {
		data = append(data, v)
	}
	fmt.Println(stats.Median(data))

	<-done
}
