package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

type Generator interface {
	Generate(w io.Writer, records []*Record) error
}

type GeneratorFunc func(w io.Writer, records []*Record) error

func (fn GeneratorFunc) Generate(w io.Writer, records []*Record) error {
	return fn(w, records)
}

func NewGenerator() Generator {
	written := map[string]bool{}

	return GeneratorFunc(func(w io.Writer, records []*Record) error {
		if records == nil {
			return nil
		}

		err := WriteSchema(w, "hello world")
		if err != nil {
			return errors.Wrapf(err, "Unable to write Marshal")
		}

		err = WriteMarshal(w, records)
		if err != nil {
			return errors.Wrapf(err, "Unable to write Marshal")
		}

		err = WriteUnmarshal(w, records)
		if err != nil {
			return errors.Wrapf(err, "Unable to write Unmarshal")
		}

		for len(records) > 0 {
			record := records[0]
			records = records[1:]

			if written[record.Name] {
				// don't write records again
				continue
			}

			v, err := WriterRecord(w, record)
			if err != nil {
				return err
			}
			written[record.Name] = true

			if v != nil {
				records = append(records, v...)
			}
		}

		return nil
	})
}

func WriterRecord(w io.Writer, record *Record) ([]*Record, error) {
	if record.Type != "record" {
		return nil, errors.New("Generator only handles record types for now")
	}

	fmt.Fprintf(w, "type %v struct {\n", record.Name)

	results := []*Record{}
	if record.Fields != nil {
		for _, field := range record.Fields {
			v, err := WriteField(w, field)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to render field, %v", field.Name)
			}
			if v != nil {
				results = append(results, v...)
			}
		}
	}
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "")

	return results, nil
}

func WriteField(w io.Writer, f Field) ([]*Record, error) {
	switch string(f.Type) {
	case `"string"`:
		fmt.Fprintf(w, "  %v string\n", f.Name)
	case `"long"`:
		fmt.Fprintf(w, "  %v int64\n", f.Name)
	case `"int"`:
		fmt.Fprintf(w, "  %v int\n", f.Name)
	case `"float"`:
		fmt.Fprintf(w, "  %v float\n", f.Name)
	case `"double"`:
		fmt.Fprintf(w, "  %v double\n", f.Name)
	default:
		data := f.Type
		ptr := ""
		if bytes.HasPrefix(f.Type, []byte(`[`)) {
			ptr = "*"
			v, err := WithoutNull(data)
			if err != nil {
				return nil, err
			}
			data = v
		}

		if !bytes.HasPrefix(data, []byte(`{`)) {
			return nil, fmt.Errorf("unhandled type, %v", string(f.Type))
		}
		records, err := Parse(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		if v := len(records); v == 0 {
			return nil, fmt.Errorf("unable to parse records from field, %v", f.Name)
		} else if v != 1 {
			return nil, fmt.Errorf("Cannot handle fields with union types for field, %v", f.Name)
		}

		fmt.Fprintf(w, "  %v %v%v\n", f.Name, ptr, records[0].Name)
		return records, nil
	}

	return nil, nil
}

func WithoutNull(raw json.RawMessage) (json.RawMessage, error) {
	v := []json.RawMessage{}
	err := json.Unmarshal(raw, &v)
	if err != nil {
		return nil, err
	}
	if len(v) != 2 {
		return nil, errors.New("Only option null handled")
	}

	if string(v[0]) == `null` {
		return v[1], nil
	} else if string(v[1]) == `null` {
		return v[0], nil
	} else {
		return nil, errors.New("neither field was `null`")
	}
}

func WriteSchema(w io.Writer, schema string) error {
	fmt.Fprintf(w, `
const (
	%v = %v%v%v
)

`, "schema", "`", schema, "`")
	return nil
}

func WriteUnmarshal(w io.Writer, records []*Record) error {
	fmt.Fprintln(w, "func Unmarshal(data []byte) (interface{}, error) {")
	fmt.Fprintln(w, `
	payload := make([]byte, 0, len(schema) + len(data))
	payload = append(schema, data...)
	r, err := goavro.NewOCFReader(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	`)
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "")
	return nil
}

func WriteMarshal(w io.Writer, records []*Record) error {
	fmt.Fprintln(w, "func Marshal(v interface{}) ([]byte, error) {")
	fmt.Fprintln(w, "}")
	fmt.Fprintln(w, "")
	return nil
}
