package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSigning(t *testing.T) {
	pri, pub, err := GenerateKeypair("test1")
	assert.Nil(t, err)
	assert.Nil(t, SignFile(pri, "auth.go", "test1.sign"))
	assert.Nil(t, VerifyFile(pub, "auth.go", "test1.sign"))
	assert.NotNil(t, VerifyFile(pub, "auth_test.go", "test1.sign"))
}
