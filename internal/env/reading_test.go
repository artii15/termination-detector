package env_test

import (
	"os"
	"testing"

	"github.com/artii15/termination-detector/internal/env"
	"github.com/stretchr/testify/assert"
)

func TestMustRead(t *testing.T) {
	testEnvVarName := "ENVS_READING_TEST"
	testEneVarValue := "dummy"
	err := os.Setenv(testEnvVarName, testEneVarValue)
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv(testEnvVarName)
		assert.NoError(t, err)
	}()

	assert.Equal(t, testEneVarValue, env.MustRead(testEnvVarName))
	assert.Panics(t, func() {
		env.MustRead("NOT_EXISTING_ENV_VAR")
	})
}
