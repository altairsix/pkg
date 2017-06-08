package app

import "encoding/json"

type Field struct {
	Name string
	Type json.RawMessage
}

type Record struct {
	Type      string
	Name      string
	Namespace string
	Doc       string
	Fields    []Field
}
