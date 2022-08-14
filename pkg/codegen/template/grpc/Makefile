APP_NAME={{ .Projectname}}
MOD_FQ_NAME={{ .FullyQualifiedModuleName}}
OUTPUT_FOLDER={{ .Out.Foldername}}
BINARY_NAME={{ .Out.Binaryname}}
CMD_NAME={{ .Cmd.Foldername}}
GCP_REGION=us-central1
GCP_ZONE=${GCP_REGION}-a

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

grpc-mod-update:
	cd {{ .Api.Parentfolder}}; buf mod update

grpc-mod-lint: grpc-mod-update
	cd {{ .Api.Parentfolder}}; buf lint

grpc-codegen: grpc-mod-update
	buf generate

# Ensure `create-module` target has been executed once
cloud-run:
	gcloud run deploy ${APP_NAME} --platform=managed --allow-unauthenticated --region=${GCP_REGION} --source=.

undeploy-cloud-run:
	gcloud run services delete new-cr-go-module --platform=managed --region=us-central1 --quiet