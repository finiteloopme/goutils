# goutils
Collection of golang utilities

# Usage
1. Simple utility to start HTTP Server
2. Log handler
3. Environment variable reader
4. Generates application scafold
5. Random number generator

# Build
## Create a Golang microservice
```zsh
# example usage
goutils create-app \
    --type=go-grpc \
    --name=hello \
    --fqdn-name=github.com/finiteloopme/demo/hello \
    --output=hello
```

## Configure basic build
```zsh
cd hello
# Initiatlise basic Golang module
make create-module
# Generate gRPC code
make grpc-codegen
# Tidy up all the dependencies
make fmt-deps
```
