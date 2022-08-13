// Implementation of io.Writer interface
package io

import (
	"fmt"
	"os"
	"strings"
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
	if err := CreateDir(path); err != nil {
		return 0, fmt.Errorf("Error creating directory (%v). Error: (%v)", path, err)
	}
	// Open the file in Append mode, if it exists.
	// Else open the file in append mode
	f, err := os.OpenFile(fw.Filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, fmt.Errorf("Error opening file (%v). Error: (%v)", fw.Filename, err)
	}
	defer f.Close()
	// Write the contents to the file
	n, err := f.Write(p)
	if err != nil {
		return 0, fmt.Errorf("Error writing to file (%v). Error: (%v)", fw.Filename, err)
	}
	return n, nil
}
