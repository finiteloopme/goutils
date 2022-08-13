// Create directory if one doesn't exist
package io

import (
	"fmt"
	"os"
)

func CreateDir(dirname string) error {

	// Check if the file exists
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		// File doesn't exist
		return os.MkdirAll(string(dirname), os.ModePerm)
	} else if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}
