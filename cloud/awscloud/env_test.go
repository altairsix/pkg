package awscloud_test

import (
	"testing"

	"github.com/altairsix/pkg/cloud/awscloud"
	"github.com/stretchr/testify/assert"
)

func TestEnvRegion(t *testing.T) {
	assert.Equal(t, "us-west-2", awscloud.EnvRegion())
}
