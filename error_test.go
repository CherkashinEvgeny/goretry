package retry

import (
	"github.com/pkg/errors"
	"gotest.tools/assert"
	"testing"
)

var testError = errors.New("test")

func TestUnrecoverableCause(t *testing.T) {
	unrecErr := Unrecoverable(testError)
	causeErr, ok := getUnrecoverableErrorCause(unrecErr)
	assert.Assert(t, ok)
	assert.Equal(t, testError, causeErr)
}
