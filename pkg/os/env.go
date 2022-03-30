package pkg

import (
	"errors"
	"os"
	"github.com/finiteloopme/goutils/pkg/log"
)

func ReadEnvVar(key string) string{
	value := os.Getenv(key)
	if value == "" {
		log.Fatal(errors.New("Environment variable " + key + " not set"))
	}
	return value
}
