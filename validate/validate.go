package validate

import (
	"reflect"
	"strings"
)

var (
	TagName = "validate"
)

func Struct(in interface{}) error {
	value := reflect.ValueOf(in)
	typ := value.Type()
	if typ.Kind() == reflect.Ptr {
		value = value.Elem()
		typ = typ.Elem()
	}

	errs := FieldErrs{}
	for i := value.NumField() - 1; i >= 0; i-- {
		field := typ.Field(i)
		v := value.Field(i).Interface()

		validations := field.Tag.Get(TagName)
		if validations == "" {
			continue
		}

		fieldName := Name(field)
		for _, name := range strings.Split(validations, ",") {
			fn, ok := TagMap[name]
			if !ok {
				continue
			}

			if ok = fn(v); !ok {
				fieldErr, ok := errs[fieldName]
				if !ok {
					fieldErr = FieldErr{}
				}

				fieldErr = append(fieldErr, name)
				errs[fieldName] = fieldErr
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func Name(in reflect.StructField) string {
	tags := in.Tag.Get("json")
	if tags == "" {
		return in.Name
	}

	if strings.HasPrefix(tags, ",") {
		return in.Name
	}

	segments := strings.Split(tags, ",")
	return segments[0]
}
