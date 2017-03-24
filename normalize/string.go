package normalize

import "strings"

func String(v string) string {
	return strings.TrimSpace(v)
}

func StringLower(v string) string {
	return strings.ToLower(String(v))
}

func Email(v string) string {
	return StringLower(v)
}
