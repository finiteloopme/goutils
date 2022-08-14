package main

import (
	"flag"
	"os"

	"github.com/finiteloopme/goutils/pkg/codegen"
	"github.com/finiteloopme/goutils/pkg/log"
)

const (
	CREATE_APP string = "create-app"
)

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case CREATE_APP:
		ProcessCreateApp()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usageMessage := "usage:" +
		"goutils create-app --type go-simple --name app-name --fqdn-name module-repo --output app-name"
	log.Info(usageMessage)
	return
}

func ProcessCreateApp() {
	// Example
	// goutils create-app --type go-simple --fqdn-name github.com/finiteloopme/demo/go-simple --output go-simple
	createAppCmd := flag.NewFlagSet(CREATE_APP, flag.ExitOnError)
	appType := createAppCmd.String("type", "go-simple", "Application type.  Only 'go-simple' is supported")
	appName := createAppCmd.String("name", "", "Application name")
	fqdnName := createAppCmd.String("fqdn-name", "", "Fully qualified module name to use with `go mod init`")
	appLocation := createAppCmd.String("output", *appName, "Folder name to host the app")
	createAppCmd.Parse((os.Args[2:]))
	switch *appType {
	case "go-simple":
		codegen.NewSimpleGoModule(*appName, *fqdnName, *appLocation)
	default:
		printUsage()
		os.Exit(1)
	}

	return
}