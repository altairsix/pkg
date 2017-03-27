package queue_test

import (
	"io/ioutil"
	"testing"

	"github.com/altairsix/pkg/cloud/awscloud/queue"
	"github.com/altairsix/pkg/local"
	"github.com/stretchr/testify/assert"
)

func TestSubscribe(t *testing.T) {
	u1, err := queue.Subscribe(ioutil.Discard, local.SNS, local.SQS, "debug", "debug-matt")
	assert.Nil(t, err)
	assert.NotNil(t, u1)

	u2, err := queue.Subscribe(ioutil.Discard, local.SNS, local.SQS, "debug", "debug-matt")
	assert.Nil(t, err)
	assert.NotNil(t, u2)

	assert.Equal(t, *u1, *u2)
}
