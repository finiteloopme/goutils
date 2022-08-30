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

# Start a HTTP Handler
1. Update code in `./cmd/main.go`
   - Don't start the gRPC service
     > ```golang
     > // server.RunServer()
     > ```
   - Start the HTTP Handler
     > ```golang
     > server.RunWithHTTPHandler()
     > ```
   - Start the Service: `make run`
2. Test the HTTP handler service:
   - `curl -d "" http://localhost:8090/user.v1alpha1.HelloService/SayHello`
3. Sample Response
   ```json
   {"msg":"SGVsbG8sIFdvcmxk", "respondedAt":null}
   ```
   > The `msg` is a slice (array) of byte.  UTF-8 encoded.  
   > Eg cmd to see the actual message:  
   > ```bash
   > echo SGVsbG8sIFdvcmxk | base64 -d
   > ```