package env

import (
	"fmt"
	"os"
)

func MustRead(envVarName string) string {
	envVarValue, isSet := os.LookupEnv(envVarName)
	if !isSet {
		panic(fmt.Sprintf("env variable %s is not set", envVarName))
	}
	return envVarValue
}
