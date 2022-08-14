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

type ApiType struct {
	Parentfolder string `default:"api"`
	// name of the API
	Name string `default:"user"`
	// API version
	Version string `default:"v1alpha1"`
	// Location for the generated code. Relative to parent
	CodeGenLocation string `default:"gen/proto/go"`
}

type OutPutStructure struct {
	Foldername string `default:"bin"`
	Binaryname string
}

type CmdStructure struct {
	Foldername string `default:"cmd"`
}

type PkgStructure struct {
	Foldername string `default:"pkg"`
}

type InternalStructure struct {
	Foldername string `default:"internal"`
}

type MakefileStructure struct {
	Filename string `default:"Makefile"`
}

type ReadmeStructure struct {
	Filename string `default:"README.md"`
}

type DockerStructure struct {
	Filename string `default:"Dockerfile"`
}

type BufYamlStructure struct {
	Filename string `default:"buf.yaml"`
}
type BufGenYamlStructure struct {
	Filename string `default:"buf.gen.yaml"`
}

type BufWorkYamlStructure struct {
	Filename string `default:"buf.work.yaml"`
}

type ServerGoStructure struct {
	Filename string `default:"server.go"`
}

type UserGoStructure struct {
	Filename string `default:"user.proto"`
}

//go:embed template
var templatesFS embed.FS

// TODO: buf.build
// 1. buf.yaml
// 2. buf.gen.yaml
// 3. buf.work.yaml
// 4. .gitignore

type ProjectStructure struct {
	Projectname              string
	FullyQualifiedModuleName string
	Api                      ApiType
	BufYaml                  BufYamlStructure
	User                     UserGoStructure
	Out                      OutPutStructure
	Cmd                      CmdStructure
	Pkg                      PkgStructure
	Internal                 InternalStructure
	ServerGo                 ServerGoStructure
	Make                     MakefileStructure
	ReadMe                   ReadmeStructure
	Dockerfile               DockerStructure
	BufGenYaml               BufGenYamlStructure
	BufWorkYaml              BufWorkYamlStructure
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

	// Dockerfile
	// Ignored

	return projStruct
}

func NewCloudRunGoModule(moduleName string, fullyQualifiedModuleName string, outputDir string) ProjectStructure {
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
	projStruct.parseTemplate("template/cloudrun/main.go_template", outputDir+"/"+projStruct.Cmd.Foldername+"/main.go")

	// Pkg
	io.CreateDir(outputDir + "/" + projStruct.Pkg.Foldername)

	// Internal
	io.CreateDir(outputDir + "/" + projStruct.Internal.Foldername)

	// Makefile
	projStruct.parseTemplate("template/cloudrun/Makefile", outputDir+"/"+projStruct.Make.Filename)

	// Dockerfile
	projStruct.parseTemplate("template/cloudrun/Dockerfile_template", outputDir+"/"+projStruct.Dockerfile.Filename)

	return projStruct
}

func NewGRPCGoModule(moduleName string, fullyQualifiedModuleName string, outputDir string) ProjectStructure {
	var projStruct ProjectStructure
	// var err error
	envconfig.Process("", &projStruct)

	// Projectname
	projStruct.Projectname = moduleName

	// FullyQualifiedModuleName
	projStruct.FullyQualifiedModuleName = fullyQualifiedModuleName

	// Api
	io.CreateDir(outputDir + "/" + projStruct.Api.Parentfolder + "/" + projStruct.Api.Name + "/" + projStruct.Api.Version)
	projStruct.parseTemplate("template/grpc/buf.yaml_template",
		outputDir+"/"+projStruct.Api.Parentfolder+"/"+projStruct.BufYaml.Filename)
	projStruct.parseTemplate("template/grpc/user.proto_template",
		outputDir+"/"+projStruct.Api.Parentfolder+"/"+projStruct.Api.Name+"/"+projStruct.Api.Version+"/"+projStruct.User.Filename)

	// Out
	projStruct.Out.Binaryname = moduleName
	io.CreateDir(outputDir + "/" + projStruct.Out.Foldername)

	// Cmd
	io.CreateDir(outputDir + "/" + projStruct.Cmd.Foldername)
	projStruct.parseTemplate("template/grpc/main.go_template", outputDir+"/"+projStruct.Cmd.Foldername+"/main.go")

	// Pkg
	io.CreateDir(outputDir + "/" + projStruct.Pkg.Foldername)

	// Internal
	io.CreateDir(outputDir + "/" + projStruct.Internal.Foldername)
	projStruct.parseTemplate("template/grpc/server.go_template",
		outputDir+"/"+projStruct.Internal.Foldername+"/"+projStruct.ServerGo.Filename)

	// Makefile
	projStruct.parseTemplate("template/grpc/Makefile", outputDir+"/"+projStruct.Make.Filename)

	// Dockerfile
	projStruct.parseTemplate("template/grpc/Dockerfile_template", outputDir+"/"+projStruct.Dockerfile.Filename)

	// buf.gen.yaml
	projStruct.parseTemplate("template/grpc/buf.gen.yaml_template",
		outputDir+"/"+projStruct.BufGenYaml.Filename)
	// buf.work.yaml
	projStruct.parseTemplate("template/grpc/buf.work.yaml_template",
		outputDir+"/"+projStruct.BufWorkYaml.Filename)

	return projStruct
}

func (p ProjectStructure) parseTemplate(templateFileName string, outputFileName string) {
	tMake, err := template.ParseFS(templatesFS, templateFileName)
	if err != nil {
		log.Fatal(err)
	}
	if err := tMake.Execute(&io.FileWriter{Filename: outputFileName}, p); err != nil {
		log.Fatal(err)
	}
}
