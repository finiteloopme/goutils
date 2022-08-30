# goutils
Collection of golang utilities

# Create a Golang microservice
```zsh
# example usage
goutils create-app \
    --type=go-grpc \
    --name=hello \
    --fqdn-name=github.com/finiteloopme/demo/hello \
    --output=hello
```

# Configure basic build
```zsh
cd hello
# Initiatlise basic Golang module
make create-module
# Generate gRPC code
make grpc-codegen
# Tidy up all the dependencies
make fmt-deps
```
