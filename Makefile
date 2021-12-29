GOFLAGS=-mod=vendor
dummy:

ensure-deps:
	go mod tidy
	go mod vendor

build: ensure-deps
	go build ./pkg/...

test: 
	go test ./pkg/... ./cmd/...

release: