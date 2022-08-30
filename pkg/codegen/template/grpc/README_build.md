# Steps to build the package
```sh
# Initiatlise basic Golang module
make create-module
# Generate gRPC code
make grpc-codegen
# Tidy up all the dependencies
make fmt-deps
# Build a binary
make build
```

# Test gRPC
```zsh
make test
```

# Start the gRPC server
```sh
make run
```