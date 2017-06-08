package app_test

import (
	"strings"
	"testing"

	"github.com/altairsix/pkg/cmd/avrogen/app"
	"github.com/stretchr/testify/assert"
)

func TestParseHandlesSingleObject(t *testing.T) {
	schema := `{
	"type": "record",
	"name": "Weather",
	"namespace": "test",
	"doc": "A weather reading.",
	"fields": [
		{
			"name": "person",
			"type": {
				"name": "Person",
				"type": "record",
				"fields": [
					{
						"name": "first",
						"type": "string"
					}
				]
			}
		},
		{
			"name": "time",
			"type": "long"
		},
		{
			"name": "temp",
			"type": "int"
		}
	]
}`

	records, err := app.Parse(strings.NewReader(schema))
	assert.Nil(t, err)
	assert.Len(t, records, 1)
}

func TestParseHandlesUnion(t *testing.T) {
	schema := `[
	{
	"type": "record",
	"name": "Weather",
	"namespace": "test",
	"doc": "A weather reading.",
	"fields": [
		{
			"name": "person",
			"type": {
				"name": "Person",
				"type": "record",
				"fields": [
					{
						"name": "first",
						"type": "string"
					}
				]
			}
		},
		{
			"name": "time",
			"type": "long"
		},
		{
			"name": "temp",
			"type": "int"
		}
	]
}
]`

	records, err := app.Parse(strings.NewReader(schema))
	assert.Nil(t, err)
	assert.Len(t, records, 1)
}
