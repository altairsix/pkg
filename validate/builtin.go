package validate

import (
	"regexp"
	"time"

	"github.com/altairsix/pkg/normalize"
	"github.com/altairsix/pkg/types"

	"gopkg.in/asaskevich/govalidator.v4"
)

type Func func(interface{}) bool

var TagMap = map[string]Func{
	"required": IsRequired,
	"email":    IsEmail,
	"phone":    IsPhone,
	"label":    IsLabel,
	"number":   IsNumber,
}

var (
	reNumber = regexp.MustCompile(`^\d+$`)
)

func IsNumber(in interface{}) bool {
	switch v := in.(type) {
	case []byte:
		return reNumber.Match(v)
	case string:
		return reNumber.MatchString(v)
	}

	return false
}

func IsLabel(in interface{}) bool {
	switch v := in.(type) {
	case string:
		return len(v) < 96
	}

	return false
}

func IsRequired(in interface{}) bool {
	switch v := in.(type) {
	case int:
		return v != 0
	case int64:
		return v != 0
	case time.Time:
		return !v.IsZero()
	case string:
		return v != ""
	case map[string]interface{}:
		return len(v) > 0
	case types.StringSet:
		return len(v) > 0
	case types.StringArray:
		return len(v) > 0
	case types.ID:
		return v > 0
	case []types.ID:
		return len(v) > 0
	case types.Key:
		return v != ""
	case []types.Key:
		return len(v) > 0
	case []bool:
		return len(v) > 0
	case []struct{}:
		return len(v) > 0
	case []string:
		return len(v) > 0
	case []int:
		return len(v) > 0
	case []int8:
		return len(v) > 0
	case []int16:
		return len(v) > 0
	case []int32:
		return len(v) > 0
	case []int64:
		return len(v) > 0
	case []uint:
		return len(v) > 0
	case []uint8:
		return len(v) > 0
	case []uint16:
		return len(v) > 0
	case []uint32:
		return len(v) > 0
	case []uint64:
		return len(v) > 0
	case []float32:
		return len(v) > 0
	case []float64:
		return len(v) > 0
	}
	return true
}

func IsEmail(in interface{}) bool {
	switch v := in.(type) {
	case string:
		return govalidator.IsEmail(v)
	default:
		return false
	}
}

func IsPhone(in interface{}) bool {
	switch v := in.(type) {
	case string:
		_, ok := normalize.Phone(v)
		return ok
	default:
		return false
	}
}
