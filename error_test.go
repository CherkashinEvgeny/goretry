package retry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testError = errors.New("test")

func TestGetUnrecoverableCause(t *testing.T) {
	unrecErr := Unrecoverable(testError)
	causeErr := getUnrecoverableErrorCause(unrecErr)
	assert.Equal(t, testError, causeErr)
}
