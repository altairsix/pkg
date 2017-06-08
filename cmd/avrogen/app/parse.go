package app

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

func Parse(r io.Reader) ([]*Record, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Handle Single Object
	record := &Record{}
	err = json.Unmarshal(data, record)
	if err == nil {
		return []*Record{record}, err
	}
	if v, ok := err.(*json.UnmarshalTypeError); !ok || v.Value != "array" {
		return nil, err
	}

	// Handle Union of Objects
	records := []*Record{}
	err = json.Unmarshal(data, &records)
	return records, err
}
