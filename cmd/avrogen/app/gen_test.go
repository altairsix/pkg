package app_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/altairsix/pkg/cmd/avrogen/app"
	"github.com/stretchr/testify/assert"
)

func TestGen(t *testing.T) {
	person := json.RawMessage(`{
					"name": "Person",
					"type": "record",
					"fields": [
						{
							"name": "Name",
							"type": "string"
						}
					]
				}`)

	records := []*app.Record{
		{
			Type: "record",
			Name: "A",
			Fields: []app.Field{
				{
					Name: "First",
					Type: json.RawMessage(`"string"`),
				},
				{
					Name: "Person",
					Type: person,
				},
			},
		},
		{
			Type: "record",
			Name: "B",
			Fields: []app.Field{
				{
					Name: "First",
					Type: json.RawMessage(`"string"`),
				},
				{
					Name: "Person",
					Type: person,
				},
			},
		},
	}

	err := app.NewGenerator().Generate(os.Stdout, records)
	assert.Nil(t, err)
}

func TestGenOptional(t *testing.T) {
	person := json.RawMessage(`[{
					"name": "Person",
					"type": "record",
					"fields": [
						{
							"name": "Name",
							"type": "string"
						}
					]
				}, null]`)

	records := []*app.Record{
		{
			Type: "record",
			Name: "A",
			Fields: []app.Field{
				{
					Name: "First",
					Type: json.RawMessage(`"string"`),
				},
				{
					Name: "Person",
					Type: person,
				},
			},
		},
		{
			Type: "record",
			Name: "B",
			Fields: []app.Field{
				{
					Name: "First",
					Type: json.RawMessage(`"string"`),
				},
				{
					Name: "Person",
					Type: person,
				},
			},
		},
	}

	err := app.NewGenerator().Generate(os.Stdout, records)
	assert.Nil(t, err)
}
