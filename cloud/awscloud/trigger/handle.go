package trigger

import (
	"bytes"
	"sync"

	"github.com/apex/go-apex"
	"github.com/apex/go-apex/sns"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/savaki/jq"
)

const (
	maxMessages = int64(10)
)

var (
	op               = jq.Must(jq.Parse(".source"))
	cloudWatchSource = []byte(`"aws.events"`)
)

type Handler interface {
	Apply(v string) error
}

type HandlerFunc func(string) error

func (fn HandlerFunc) Apply(v string) error {
	return fn(v)
}

// SQS provides an interface over *sqs.SQS to simplify testing
type SQS interface {
	// ReceiveMessage is the same as (*sqs.SQS) ReceiveMessage
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)

	// DeleteMessageBatch is the same as (*sqs.SQS) DeleteMessageBatch
	DeleteMessageBatch(input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageBatchOutput, error)
}

func HandleFunc(sqsApi SQS, queryUrl *string, concurrency int, fn HandlerFunc) sns.HandlerFunc {
	return Handle(sqsApi, queryUrl, concurrency, fn)
}

func Handle(sqsApi SQS, queryUrl *string, concurrency int, fn Handler) sns.HandlerFunc {
	return func(event *sns.Event, c *apex.Context) error {
		ch, err := receiveMessages(sqsApi, queryUrl, event)
		if err != nil {
			return err
		}

		del := processMessages(ch, concurrency, fn)
		deleteMessages(sqsApi, queryUrl, del)

		return nil
	}
}

func receiveMessages(sqsApi SQS, queryUrl *string, event *sns.Event) (<-chan *sqs.Message, error) {
	in := &sqs.ReceiveMessageInput{
		MaxNumberOfMessages: aws.Int64(maxMessages),
		QueueUrl:            queryUrl,
		WaitTimeSeconds:     aws.Int64(1),
	}

	if event != nil && event.Records != nil {
		for _, record := range event.Records {
			if v, err := op.Apply([]byte(record.SNS.Message)); err == nil && bytes.Compare(v, cloudWatchSource) == 0 {
				in.SetWaitTimeSeconds(0)
			}
		}
	}

	out, err := sqsApi.ReceiveMessage(in)
	if err != nil {
		return nil, err
	}

	// populate the input chan
	//
	ch := make(chan *sqs.Message, maxMessages)
	for _, message := range out.Messages {
		ch <- message
	}
	close(ch)

	return ch, nil
}

func processMessages(ch <-chan *sqs.Message, concurrency int, h Handler) <-chan *sqs.Message {
	// launch workers
	//
	wg := &sync.WaitGroup{}
	wg.Add(concurrency)
	del := make(chan *sqs.Message, maxMessages)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for m := range ch {
				if err := h.Apply(*m.Body); err == nil {
					del <- m
				}
			}
		}()
	}

	// close del chan once all workers have completed
	//
	go func() {
		defer close(del)
		wg.Wait()
	}()

	return del
}

func deleteMessages(sqsApi SQS, queryUrl *string, del <-chan *sqs.Message) {
	// delete successfully processed records
	//
	delIn := &sqs.DeleteMessageBatchInput{
		QueueUrl: queryUrl,
		Entries:  []*sqs.DeleteMessageBatchRequestEntry{},
	}
	for v := range del {
		delIn.Entries = append(delIn.Entries, &sqs.DeleteMessageBatchRequestEntry{
			Id:            v.MessageId,
			ReceiptHandle: v.ReceiptHandle,
		})
	}
	sqsApi.DeleteMessageBatch(delIn)
}
