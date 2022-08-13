// Create three types of Go modules
// 1. Simple Go Module with a main
// 2. Go module with scafolding to deploy to Cloud Run
// 3. A Go module with API (gRPC) support
package codegen

import (
	"embed"
	"text/template"

	"github.com/finiteloopme/goutils/pkg/io"
	log "github.com/finiteloopme/goutils/pkg/log"
	"github.com/kelseyhightower/envconfig"
)

type FSType string

const (
	File   FSType = "FILE"
	Folder        = "FOLDER"
)

type ApiStructure struct {
	// name of the API
	Name string
	// API version
	Version string `default:"v1"`
	// Location for the generated code. Relative to parent
	CodeGenLocation string `default:"gen/proto/go"`
	Type            FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type ApiType struct {
	Parentfolder string `default:"api"`
	Apis         []ApiStructure
	Type         FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type OutPutStructure struct {
	Foldername string `default:"bin"`
	Binaryname string
	Type       FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type CmdStructure struct {
	Foldername string `default:"cmd"`
	Type       FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type PkgStructure struct {
	Foldername string `default:"pkg"`
	Type       FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type InternalStructure struct {
	Foldername string `default:"internal"`
	Type       FSType `default:"FOLDER"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type MakefileStructure struct {
	Filename string `default:"Makefile"`
	Type     FSType `default:"FILE"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

type ReadmeStructure struct {
	Filename string `default:"README.md"`
	Type     FSType `default:"FILE"`
	// Ignore for internal use only
	ignore bool `default:"false"`
}

//go:embed template/simple
var simpleTemplateFS embed.FS

// // go:embed template/simple/main.go_template
// var simpleMainTemplateFS embed.FS

// TODO: buf.build
// 1. buf.yaml
// 2. buf.gen.yaml
// 3. buf.work.yaml
// 4. .gitignore

type ProjectStructure struct {
	Projectname              string
	FullyQualifiedModuleName string
	Api                      ApiType
	Out                      OutPutStructure
	Cmd                      CmdStructure
	Pkg                      PkgStructure
	Internal                 InternalStructure
	Type                     FSType `default:"FOLDER"`
	Make                     MakefileStructure
	ReadMe                   ReadmeStructure
	// Ignore for internal use only
	ignore bool `default:"false"`
}

// Create a go module with the given name
// moduleName: the name of the go module
// fullyQualifiedModuleName: fully qualified name for the module
// outputDir: Output folder
func NewSimpleGoModule(moduleName string, fullyQualifiedModuleName string, outputDir string) ProjectStructure {
	var projStruct ProjectStructure
	// var err error
	envconfig.Process("", &projStruct)

	// Projectname
	projStruct.Projectname = moduleName

	// FullyQualifiedModuleName
	projStruct.FullyQualifiedModuleName = fullyQualifiedModuleName

	// Api
	// Ignored

	// Out
	projStruct.Out.Binaryname = moduleName
	io.CreateDir(outputDir + "/" + projStruct.Out.Foldername)

	// Cmd
	io.CreateDir(outputDir + "/" + projStruct.Cmd.Foldername)
	projStruct.parseTemplate("template/simple/main.go_template", outputDir+"/"+projStruct.Cmd.Foldername+"/main.go")

	// Pkg
	io.CreateDir(outputDir + "/" + projStruct.Pkg.Foldername)

	// Internal
	io.CreateDir(outputDir + "/" + projStruct.Internal.Foldername)

	// Makefile
	projStruct.parseTemplate("template/simple/Makefile", outputDir+"/"+projStruct.Make.Filename)

	return projStruct
}

func (p ProjectStructure) parseTemplate(templateFileName string, outputFileName string) {
	tMake, err := template.ParseFS(simpleTemplateFS, templateFileName)
	if err != nil {
		log.Fatal(err)
	}
	if err := tMake.Execute(&io.FileWriter{Filename: outputFileName}, p); err != nil {
		log.Fatal(err)
	}
}
