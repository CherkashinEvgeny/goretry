package retry

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testError = errors.New("test")

func TestUnrecoverableCause(t *testing.T) {
	unrecErr := Unrecoverable(testError)
	causeErr, ok := getUnrecoverableErrorCause(unrecErr)
	assert.True(t, ok)
	assert.Equal(t, testError, causeErr)
}
