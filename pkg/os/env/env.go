// Process configuration for the app
//
// Order of priority:
// 1. Environment config: see github.com/kelseyhightower/envconfig
// 2. Config file: optional. Default: ./config.yaml
package env

import (
	"os"

	"github.com/finiteloopme/goutils/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

func ProcessFileconfig(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Warn("error reading config file. ", err)
	}

	return f, err
}

func ProcessEnvconfig(prefixEnvVar string, structRecord interface{}) error {
	err := envconfig.Process(prefixEnvVar, structRecord)
	if err != nil {
		log.Warn("Error reading environment variables. ", err)
		return err
	}
	return nil
}

func Process(prefixEnvVar string, structRecord interface{}, _configFilename ...string) error {
	var file *os.File
	var err error
	if len(_configFilename) > 0 && _configFilename[0] != "" {
		file, err = ProcessFileconfig(_configFilename[0])
		if err != nil {
			log.Warn("error reading config file. ", err)
			return err
		}
	} else {
		configFilename := "./config.yaml"
		file, err = ProcessFileconfig(configFilename)
		if err != nil {
			log.Warn("error reading default config file. ", err)
		}
		// Don't return error
		// Might be using ENV variables
	}

	if err == nil {
		err = yaml.NewDecoder(file).Decode(structRecord)
		if err != nil {
			log.Warn("Error decoding config file to struct. ", err)
			// check if environment variables are configured
		}
	}
	return ProcessEnvconfig(prefixEnvVar, structRecord)
}
