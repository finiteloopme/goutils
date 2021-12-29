package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	LogFormatDefault = "text"
	LogFormatJSON    = "json"
)

func init() {
	// Read the LOG_FORMAT env variable for log formatting
	var logFormat = os.Getenv("LOG_FORMAT")

	if logFormat == LogFormatJSON {
		//Set Log format to JSON
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// Set Log format to Text
		log.SetFormatter(&log.TextFormatter{})
	}
}

func Info(msg string) {
	log.Info(msg)

}

func Warn(msg string, err error) {
	log.Warn(msg)
	log.Warn(err)
}

func Debug(msg string) {
	log.Debug(msg)
}
