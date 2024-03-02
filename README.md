# secure-grpc-in-go
Secure gRPC in Golang

## Install gRPC
```bash
$ go get -u google.golang.org/grpc
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
