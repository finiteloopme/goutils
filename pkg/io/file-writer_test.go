package io

import (
	"os"
	"testing"
	"time"
)

var test_file_name string = "/tmp/goutils/"

func setup(t testing.TB) func(testing.TB) {
	// do init
	test_file_name += time.Now().Format(time.RFC3339) + "/io-writer-test.txt"
	// teardown
	return func(t testing.TB) {

	}
}

func TestWrite(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)
	expectedContent := "This is file is used for unit testing of io.Writer implementaiton in goutils"

	fw := FileWriter{Filename: test_file_name}
	fw.Write([]byte(expectedContent))

	readContents, err := os.ReadFile(test_file_name)
	if err != nil {
		t.Fatalf("Error reading file [%v]", test_file_name)
	}

	if string(readContents) != expectedContent {
		t.Fatalf("Expected contents (%v), received (%v)", expectedContent, string(readContents))
	}
}
