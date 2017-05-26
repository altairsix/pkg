package epoch

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/pkg/errors"
)

type Millis int64

const (
	Scale = int64(time.Millisecond)
)

var (
	timeFormats = []string{
		time.RFC3339,
		time.RFC1123,
	}
)

func Now() Millis {
	return Time(time.Now())
}

func (em Millis) Add(d time.Duration) Millis {
	return em + Millis(d/time.Millisecond)
}

func (em Millis) Time() time.Time {
	v := int64(em) * Scale
	return time.Unix(v/int64(time.Second), v%int64(time.Second))
}

func (em Millis) Format(layout string) string {
	return em.Time().In(PT).Format(layout)
}

func (em Millis) AttributeValue() *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{
		N: aws.String(strconv.FormatInt(int64(em), 10)),
	}
}

func (em Millis) Value() (driver.Value, error) {
	if em == 0 {
		return nil, nil
	}
	return em.Time(), nil
}

func (em *Millis) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case time.Time:
		if t := Time(v); t > 0 {
			*em = t
		}
		return nil

	default:
		return fmt.Errorf("unhandled type, %#v", v)
	}
}

func (em Millis) Int64() int64 {
	return int64(em)
}

func (em Millis) MarshalJSON() ([]byte, error) {
	if em == 0 {
		return []byte("null"), nil
	}

	t := em.Time().In(PT)
	v := map[string]interface{}{
		"Date":  t.Format("1/2/2006"),
		"Time":  t.Format(time.Kitchen),
		"Value": em.Int64(),
		"date":  t.Format("1/2/2006"),
		"time":  t.Format(time.Kitchen),
		"value": em.Int64(),
		"ago":   (Now() - em).Ago(),
	}

	return json.Marshal(v)
}

type millisModel struct {
	Value int64
}

func (em *Millis) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		v := ""
		err := json.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		for _, layout := range timeFormats {
			t, err := time.Parse(layout, v)
			if err == nil {
				*em = Time(t)
				return nil
			}
		}

		return errors.New("invalid time format")

	} else if data[0] == '{' {
		v := millisModel{}
		err := json.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		*em = Millis(v.Value)

	} else {
		v := int64(0)
		err := json.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		*em = Millis(v)
	}

	return nil
}

func UnixNano(v int64) Millis {
	return Millis(v / Scale)
}

func Time(t time.Time) Millis {
	return UnixNano(t.UnixNano())
}

func (em Millis) Ago() string {
	if em < 0 {
		return "-"
	}

	secs := em / 1000
	if secs < 60 {
		return "just now"
	}

	minutes := secs / 60
	if minutes == 1 {
		return "1 minute ago"
	} else if minutes < 60 {
		return strconv.Itoa(int(minutes)) + " minutes ago"
	}

	hours := minutes / 60
	if hours == 1 {
		return "1 hour ago"
	} else if hours < 24 {
		return strconv.Itoa(int(hours)) + " hours ago"
	}

	days := hours / 24
	if days == 1 {
		return "1 day ago"
	} else if days < 30 {
		return strconv.Itoa(int(days)) + " days ago"
	}

	months := days / 30
	if months == 1 {
		return "1 month ago"
	} else if months < 12 {
		return strconv.Itoa(int(months)) + " months ago"
	}

	years := months / 12
	if years == 1 {
		return "1 year ago"
	}
	return strconv.Itoa(int(years)) + " years ago"
}
