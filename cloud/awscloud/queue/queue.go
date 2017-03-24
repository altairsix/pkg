package queue

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/altairsix/pkg/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/savaki/randx"
)

type config struct {
	workers         int
	waitTimeSeconds int64
}

type Option func(*config)

func Workers(n int) Option {
	return func(c *config) {
		c.workers = n
	}
}

type HandleFunc func(k context.Kontext, message *sqs.Message) error

func Start(k context.Kontext, sqsApi *sqs.SQS, queueUrl *string, fn HandleFunc, opts ...Option) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		c := &config{
			workers: 1,
		}

		for _, opt := range opts {
			opt(c)
		}

		del := make(chan *string, 512)                // receipt handles to be deleted
		wait := deleteMessages(sqsApi, queueUrl, del) // go routine to delete threads
		ch := receiveMessages(k, sqsApi, queueUrl)    // go routine to receive messages from sqs

		wg := &sync.WaitGroup{}
		wg.Add(c.workers)
		for i := 0; i < c.workers; i++ {
			go func() {
				defer wg.Done()
				handler(k, fn, ch, del)
			}()
		}
		wg.Wait()
		close(del)

		<-wait
	}()

	return done
}

func handler(k context.Kontext, h HandleFunc, ch <-chan *Envelope, del chan<- *string) {
	defer fmt.Println("closing handler")

	for {
		select {
		case <-k.Context.Done():
			return
		case v := <-ch:
			err := h(k, v.Message)
			if err == nil {
				del <- v.Message.ReceiptHandle
			}
		}
	}
}

// deleteMessages deletes all the messages contained in the chan, del; returns a signal channel that indicates when
// all in flight messages have been deleted
func deleteMessages(sqsApi *sqs.SQS, queueUrl *string, del <-chan *string) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer fmt.Println("closing deleteMessages")

		timer := time.NewTimer(time.Second)
		defer timer.Stop()

		n := 0
		maxN := 10
		entries := make([]*sqs.DeleteMessageBatchRequestEntry, 0, maxN)

		deleteEntries := func() {
			sqsApi.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
				QueueUrl: queueUrl,
				Entries:  entries,
			})
			n = 0
			entries = entries[:0]
		}

		for {
			timer.Reset(time.Second)

			select {
			case <-timer.C:
				if n > 0 {
					deleteEntries()
				}

			case v, ok := <-del:
				if !ok {
					return
				}

				n++
				entries = append(entries, &sqs.DeleteMessageBatchRequestEntry{
					Id:            aws.String(strconv.Itoa(n)),
					ReceiptHandle: v,
				})

				if n == maxN {
					deleteEntries()
				}
			}
		}
	}()

	return done
}

func receiveMessages(k context.Kontext, sqsApi *sqs.SQS, queueUrl *string) <-chan *Envelope {
	ch := make(chan *Envelope)

	go func() {
		defer close(ch)
		defer fmt.Println("closing receiveMessages")

		timeout := time.Second * time.Duration(randx.IntN(20)+10)

		for {
			out, err := sqsApi.ReceiveMessage(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(10),
				QueueUrl:            queueUrl,
				WaitTimeSeconds:     aws.Int64(20),
			})
			if err != nil {
				select {
				case <-k.Context.Done():
					return
				case ch <- &Envelope{Err: err}:
				}

				select {
				case <-k.Context.Done():
					return
				case <-time.After(timeout):
				}
			}

			for _, m := range out.Messages {
				message := m
				select {
				case <-k.Context.Done():
					return
				case ch <- &Envelope{Message: message}:
				}
			}

			select {
			case <-k.Context.Done():
				return
			default:
			}
		}
	}()

	return ch
}

type Envelope struct {
	Message *sqs.Message
	Err     error
}
