// Implementation of io.Writer interface
package io

import (
	"os"
	"strings"

	log "github.com/finiteloopme/goutils/pkg/log"
)

// Filename to be used, including the path
type FileWriter struct {
	Filename string
}

// Writes the contents to the file.
// Appends the contents to the file, if file exists
// Else a file is created.  Including the required directories.
func (fw *FileWriter) Write(p []byte) (int, error) {
	// Get the path from the filename
	lastSlash := strings.LastIndex(fw.Filename, "/")
	path := fw.Filename[0:lastSlash]
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist
		os.MkdirAll(string(path), os.ModePerm)
	}
	// Open the file in Append mode, if it exists.
	// Else open the file in append mode
	f, err := os.OpenFile(fw.Filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	// Write the contents to the file
	n, err := f.Write(p)
	if err != nil {
		log.Fatal(err)
	}
	return n, nil
}
