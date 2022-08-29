MOD_FQ_NAME={{ .FullyQualifiedModuleName}}
OUTPUT_FOLDER={{ .Out.Foldername}}
BINARY_NAME={{ .Out.Binaryname}}
CMD_NAME={{ .Cmd.Foldername}}

create-module:
	go mod init ${MOD_FQ_NAME}

fmt-deps:
	# go vet
	# go fmt
	go mod tidy

build: fmt-deps
	go build -o ${OUTPUT_FOLDER}/${BINARY_NAME} ./${CMD_NAME}/...
	# go build ./pkg/...

run: 
	go run ${CMD_NAME}/main.go

test:
	go test ./... -v -cover