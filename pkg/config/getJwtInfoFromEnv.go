package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetJwtSecretFromEnv returns a secret to be used with JWT based on the content of the env variable
// JWT_SECRET : should contain a string with your secret
func GetJwtSecretFromEnv() (string, error) {
	var err error
	val, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		return "", &Error{
			err: err,
			msg: "ERROR: CONFIG ENV JWT_SECRET should contain your JWT secret.",
		}
	}
	return fmt.Sprintf("%s", val), nil
}

// GetJwtDurationFromEnv returns a number  string based on the values of environment variable :
// JWT_DURATION_MINUTES : int value between 1 and 14400 minutes, 10 days seems an extreme max value
// the parameter defaultJwtDuration will be used if this env variable is not defined
// in case the ENV variable JWT_DURATION_MINUTES exists and contains an invalid integer the functions returns 0 and an error
func GetJwtDurationFromEnv(defaultJwtDuration int) (int, error) {
	JwtDuration := defaultJwtDuration

	var err error
	val, exist := os.LookupEnv("JWT_DURATION_MINUTES")
	if exist {
		JwtDuration, err = strconv.Atoi(val)
		if err != nil {
			return 0, &Error{
				err: err,
				msg: "ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain a valid integer.",
			}
		}
		if JwtDuration < 1 || JwtDuration > 14400 {
			return 0, &Error{
				err: err,
				msg: "ERROR: CONFIG ENV JWT_DURATION_MINUTES should contain an integer between 1 and 14400",
			}
		}
	}
	return JwtDuration, nil
}
