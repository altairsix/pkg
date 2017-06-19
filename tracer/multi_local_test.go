package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiIsLogger(t *testing.T) {
	tracer := Multi()
	span := tracer.StartSpan("op")
	v, ok := span.(logger)
	assert.True(t, ok)
	assert.NotNil(t, v)
}
