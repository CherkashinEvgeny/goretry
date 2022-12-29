package retry

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testError = errors.New("test")

func TestGetUnrecoverableCause(t *testing.T) {
	unrecErr := Unrecoverable(testError)
	causeErr := getUnrecoverableErrorCause(unrecErr)
	assert.Equal(t, testError, causeErr)
}
