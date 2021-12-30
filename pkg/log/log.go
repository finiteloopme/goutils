// A simple log utility.
// Supports following two log formats.  Set environment variable `LOG_FORMAT` to the corresponding value:
// 1. json
// 2. text
// Text format is the default
package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Constants for log format
const (
	LogFormatDefault = "text"
	LogFormatJSON    = "json"
)

// Configure log format based on `LOG_FORMAT` variable
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

// Log at Info level
func Info(msg string) {
	log.Info(msg)

}

// Log at WARN level.  Logs the err message first followed by the whole error
func Warn(msg string, err error) {
	log.Warn(msg)
	log.Warn(err)
}

// Log at Debug level
func Debug(msg string) {
	log.Debug(msg)
}

// Log at Fatal level.  Sequence is:
// 1. Log the actual error
// 2. Exit with code 1
func Fatal(err error) {
	log.Error(err)
	os.Exit(1)
}
