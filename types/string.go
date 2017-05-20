package types

import (
	"database/sql/driver"
	"encoding/json"
	"sort"

	"github.com/pkg/errors"
)

var (
	errInvalidType = errors.New("type:err:invalid_type")
)

type StringSet map[string]struct{}

func (i StringSet) Add(ids ...string) {
	for _, id := range ids {
		i[id] = struct{}{}
	}
}

func (i StringSet) Overlaps(that StringSet) bool {
	for key := range that {
		if _, found := i[key]; found {
			return true
		}
	}

	return false
}

func (i StringSet) Len() int {
	return len(i)
}

func (i StringSet) Array() StringArray {
	if len(i) == 0 {
		return nil
	}

	items := make(StringArray, 0, len(i))

	for id := range i {
		items = append(items, id)
	}

	sort.Strings(items)
	return items
}

func (i StringSet) Contains(v string) bool {
	_, found := i[v]
	return found
}

func (i StringSet) IsZero() bool {
	return len(i) == 0
}

func (i StringSet) IsPresent() bool {
	return !i.IsZero()
}

type StringArray []string

// Unique returns a new []string with all the empty duplicate removed
func (arr StringArray) Unique() StringArray {
	results := StringSet{}

	if arr != nil {
		for _, item := range arr {
			if item != "" {
				results.Add(item)
			}
		}
	}

	return results.Array()
}

func (arr StringArray) Map(fn func(string) string) StringArray {
	results := make(StringArray, 0, len(arr))

	if arr != nil {
		for _, item := range arr {
			if v := fn(item); v != "" {
				results = append(results, v)
			}
		}
	}

	return results
}

func (arr StringArray) Take(n int) StringArray {
	l := len(arr)
	if l == 0 {
		return StringArray{}
	}

	if l < n {
		n = l
	}
	return arr[0:n]
}

func (arr StringArray) Append(items ...string) StringArray {
	v := arr
	if v == nil {
		v = StringArray{}
	}
	return append(v, items...)
}

func (arr StringArray) Value() (driver.Value, error) {
	data, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}

	return string(data), nil
}

func (arr *StringArray) Scan(in interface{}) error {
	if in == nil {
		*arr = StringArray{}
		return nil
	}

	data, ok := in.([]byte)
	if !ok {
		return errInvalidType
	}

	item := StringArray{}
	err := json.Unmarshal(data, &item)
	if err != nil {
		return errors.Wrap(err, "questions:type:err:string_array")
	}

	*arr = item
	return nil
}
