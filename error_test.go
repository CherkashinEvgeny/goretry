package retry

import (
	"github.com/pkg/errors"
	"gotest.tools/assert"
	"testing"
)

var testError = errors.New("test")

func TestGetUnrecoverableCause(t *testing.T) {
	unrecErr := Unrecoverable(testError)
	causeErr := getUnrecoverableErrorCause(unrecErr)
	assert.Equal(t, testError, causeErr)
}
