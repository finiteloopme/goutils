GOFLAGS=-mod=vendor
dummy:

ensure-deps:
	go mod tidy
	go mod vendor

build: ensure-deps
	go build -o bin/goutils ./cmd/main.go
	go build ./pkg/...

install-local: build
	go install ./pkg/...
	
test: 
	go test ./pkg/... -v -cover

release: