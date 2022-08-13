package io

import (
	"os"
	"testing"
	"time"
)

var test_dir_name string = "/tmp/goutils/"

func setupDir(t testing.TB) func(testing.TB) {
	// do init
	test_dir_name += time.Now().Format(time.RFC3339) + "/io-create-dir"
	// teardown
	return func(t testing.TB) {

	}
}

func TestCreateDir(t *testing.T) {
	teardown := setupDir(t)
	defer teardown(t)

	// Create a dir
	if err := CreateDir(test_dir_name); err != nil {
		t.Fatalf("Error creating director (%v). Error: (%v)", test_dir_name, err)
	}
	// Check the dir exists
	if f, err := os.Stat(test_dir_name); err != nil || !f.IsDir() {
		t.Fatalf("Expected (%v) to be a folder.  Error: (%v)", test_dir_name, err)
	}
	// Try creating it again
	if err := CreateDir(test_dir_name); err != nil {
		t.Fatalf("Error creating director (%v). Error: (%v)", test_dir_name, err)
	}
}
