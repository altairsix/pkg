package fq_test

import (
	"testing"

	"github.com/altairsix/pkg/fq"
	"github.com/stretchr/testify/assert"
)

func TestTableName(t *testing.T) {
	assert.Equal(t, "a-b-c", fq.DynamoDBTableName("a", "b", "c"))
}
