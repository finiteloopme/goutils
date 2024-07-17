package env_test

import (
	"os"
	"testing"

	"github.com/finiteloopme/goutils/pkg/os/env"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func setupProcessTest(filename string, record interface{}) func() {
	if filename != "" {
		file, _ := os.Create(filename)
		yaml.NewEncoder(file).Encode(record)
	}
	return func() {
		os.Remove(filename)
	}
}
func TestProcess(t *testing.T) {
	t.Run("Fail on no config file", func(t1 *testing.T) {
		var temp struct {
			Temp  string
			Temp2 string
		}
		err := env.Process("", &temp, "file-does-not-exist.test")
		assert.Error(t1, err, "Expected file not found error")

	})

	t.Run("Success reading config file", func(t1 *testing.T) {
		type recordType struct {
			Temp  string
			Temp2 string
		}
		var expected, actual recordType

		tearDownTest := setupProcessTest("./test-config.yaml", &expected)
		defer tearDownTest()
		err := env.Process("", &actual)
		assert.NoError(t1, err, "Unexpected error. ", err)
	})
}
