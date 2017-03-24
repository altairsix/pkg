package validate

import "strings"

type FieldErr []string

func (f FieldErr) Format(name string) string {
	if f == nil {
		return name
	}

	return name + ":" + strings.Join(f, ", ")
}

func NewFieldErr(name, errName string) FieldErrs {
	return FieldErrs{
		name: FieldErr{errName},
	}
}

type FieldErrs map[string]FieldErr

func (v FieldErrs) Error() string {
	segments := make([]string, 0, len(v))
	for name, field := range v {
		segments = append(segments, field.Format(name))
	}
	return strings.Join(segments, "; ")
}
