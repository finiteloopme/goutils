package codegen

import (
	"testing"
	"time"
)

var tempFileLocation string = "/tmp"

func setup(t testing.TB) func(t testing.TB) {
	// do init
	tempFileLocation += "/goutils/"
	tempFileLocation += time.Now().Format(time.RFC3339)
	// teardown
	return func(t testing.TB) {

	}
}

func TestNewGoModule(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	moduleName := "new-go-module"
	fullyQualifiedModuleName := "github.com/testrepo/" + moduleName
	outputDir := tempFileLocation + "/simple"
	t.Log("Creating temp project at: " + outputDir)
	projStruct := NewSimpleGoModule(moduleName, fullyQualifiedModuleName, outputDir)

	if (projStruct.Projectname != moduleName) && (projStruct.FullyQualifiedModuleName != fullyQualifiedModuleName) {
		t.Fatalf("Module name expected to be %v, received %v. \n "+
			"Fully qualified name expected to be %v, received %v.", moduleName, projStruct.Projectname, fullyQualifiedModuleName, projStruct.FullyQualifiedModuleName)
	}
}
