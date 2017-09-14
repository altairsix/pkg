package tracer_test

import (
	"testing"

	"github.com/altairsix/pkg/local"
	"github.com/altairsix/pkg/tracer"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestDynamodb(t *testing.T) {
	opentracing.SetGlobalTracer(tracer.DefaultTracer)
	tracer.AWS(local.DynamoDB.Client)

	_, err := local.DynamoDB.ListTables(&dynamodb.ListTablesInput{})
	assert.Nil(t, err)
}
