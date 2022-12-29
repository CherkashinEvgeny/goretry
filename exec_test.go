package retry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSuccessExec(t *testing.T) {
	counter := 0
	err := Exec(func(retryNumber int) (err error) {
		counter++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, counter)
}

func TestSecondRetrySuccessExec(t *testing.T) {
	counter := 0
	err := Exec(func(retryNumber int) (err error) {
		counter++
		if retryNumber == 0 {
			return testError
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, counter)
}

func TestUnrecoverableFailureExec(t *testing.T) {
	counter := 0
	err := Exec(func(retryNumber int) (err error) {
		counter++
		return Unrecoverable(testError)
	})
	assert.Equal(t, err, testError)
	assert.Equal(t, 1, counter)
}
