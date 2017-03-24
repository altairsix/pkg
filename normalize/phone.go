package normalize

import "regexp"

var rePhone = regexp.MustCompile(`[^0-9]+`)

func Phone(in string) (string, bool) {
	phone := rePhone.ReplaceAllString(in, "")
	switch len(phone) {
	case 10:
		return phone, true
	case 11:
		return phone[1:], true
	default:
		return "", false
	}
}
