package trigger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v, err := op.Apply([]byte(`{"source":"abc"}`))
		assert.Nil(t, err)
		assert.Equal(t, []byte(`"abc"`), v)
	})

	t.Run("not exists", func(t *testing.T) {
		_, err := op.Apply([]byte(`{"nope":"abc"}`))
		assert.NotNil(t, err)
	})
}
