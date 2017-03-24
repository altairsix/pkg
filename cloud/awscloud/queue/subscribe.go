package queue

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

func Subscribe(w io.Writer, snsApi *sns.SNS, sqsApi *sqs.SQS, topicName, queueName string) (*string, error) {
	fmt.Fprintf(w, "Creating SNS topic, %v ... ", topicName)
	topicOut, err := snsApi.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		fmt.Fprintf(w, "failed - %v\n", err)
		return nil, errors.Wrapf(err, "queue:subscribe:err:create_topic %v", topicName)
	}
	fmt.Fprintln(w, "ok")

	fmt.Fprintf(w, "Creating SQS queue, %v ... ", queueName)
	qOut, err := sqsApi.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		fmt.Fprintf(w, "failed - %v\n", err)
		return nil, errors.Wrapf(err, "queue:subscribe:err:create_queue %v", queueName)
	}
	fmt.Fprintln(w, "ok")

	qARN, err := findQueueARN(sqsApi, qOut.QueueUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "queue:subscribe:err:find_queue_arn %v", queueName)
	}

	fmt.Fprintf(w, "Subscribing queue, %v, to topic, %v ... ", queueName, topicName)
	_, err = snsApi.Subscribe(&sns.SubscribeInput{
		Endpoint: qARN,
		Protocol: aws.String("sqs"),
		TopicArn: topicOut.TopicArn,
	})
	if err != nil {
		fmt.Fprintf(w, "failed - %v\n", err)
		return nil, errors.Wrapf(err, "queue:subscribe:err:subscribe %v -> %v", topicName, queueName)
	}
	fmt.Fprintln(w, "ok")

	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "Queue_Policy",
  "Statement":
    {
       "Sid":"Queue_AnonymousAccess_ReceiveMessage",
       "Effect": "Allow",
       "Principal": "*",
       "Action": "sqs:SendMessage",
       "Resource": "%v"
    }
}`, *qARN)

	_, err = sqsApi.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		QueueUrl: qOut.QueueUrl,
		Attributes: map[string]*string{
			sqs.QueueAttributeNamePolicy: aws.String(policy),
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "queue:subscribe:err:set_queue_attributes %v -> %v", topicName, queueName)
	}

	return qOut.QueueUrl, nil
}

func findQueueARN(sqsApi *sqs.SQS, queueUrl *string) (*string, error) {
	attr, err := sqsApi.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: queueUrl,
		AttributeNames: []*string{
			aws.String("QueueArn"),
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "subscribe:queue_arn:err:get_queue_attributes %v", *queueUrl)
	}

	return attr.Attributes["QueueArn"], nil
}
