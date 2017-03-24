package awscloud

import "github.com/aws/aws-sdk-go/aws/awserr"

func ErrCode(err error) string {
	switch v := err.(type) {
	case awserr.Error:
		return v.Code()
	default:
		return ""
	}
}
