package checkpoint_test

import (
	"context"
	"testing"

	"github.com/altairsix/pkg/checkpoint"
	"github.com/altairsix/pkg/local"
	"github.com/savaki/randx"
	"github.com/stretchr/testify/assert"
)

func TestLifecycle(t *testing.T) {
	err := checkpoint.CreateTable(local.DynamoDB, local.Env, 5, 5)
	assert.Nil(t, err)

	ctx := context.Background()
	cp := checkpoint.New(local.Env, local.DynamoDB)
	key := randx.AlphaN(12)

	offset := randx.Int63()
	err = cp.Save(ctx, key, offset)
	assert.Nil(t, err)

	actual, err := cp.Load(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, actual, offset)

	err = cp.Save(ctx, key, offset-1)
	assert.NotNil(t, err, "offsets cannot decrease")

	err = cp.Save(ctx, key, offset)
	assert.Nil(t, err, "save should be idempotent")

	err = cp.Save(ctx, key, offset+1)
	assert.Nil(t, err, "save should accept incrementing values")

	actual, err = cp.Load(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, actual, offset+1)
}
