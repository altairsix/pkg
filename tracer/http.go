package tracer

import (
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/opentracing/opentracing-go/log"
)

func AWS(client *client.Client) {
	retries := int32(0)

	client.Handlers.Build.PushFront(func(req *request.Request) {
		name := "aws." + req.ClientInfo.ServiceName + "." + req.Operation.Name
		_, ctx := NewSegment(req.Context(), name)
		req.SetContext(ctx)
	})
	client.Handlers.Retry.PushFront(func(req *request.Request) {
		atomic.AddInt32(&retries, 1)
	})
	client.Handlers.Complete.PushFront(func(req *request.Request) {
		segment := SegmentFromContext(req.Context())
		segment.LogFields(log.Int("statusCode", req.HTTPResponse.StatusCode))
		if retries > 0 {
			segment.LogFields(log.Int32("retries", retries))
		}
		segment.Finish()
	})
}
