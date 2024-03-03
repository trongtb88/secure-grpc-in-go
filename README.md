# secure-grpc-in-go
Secure gRPC in Golang

## Setup development environment

- Install `protoc`:

```bash
brew install protobuf
```

- Install `protoc-gen-go` and `protoc-gen-go-grpc`

```bash
go get google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

- Install `protoc-gen-grpc-gateway` and `protoc-gen-openapiv2`

```bash
go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
```
```
## Generate gRPC code
```bash
$ make gen
```

## Generate SSL Certificates
```bash
$ go run generate_cert.go
```

## Run the server
```bash
$ go run server.go
```
