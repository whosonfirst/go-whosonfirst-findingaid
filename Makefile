proto:
	protoc -I=./protobuf --go_out=./ ./protobuf/findingaid.proto

cli:
	go build -mod vendor -o bin/populate cmd/populate/main.go
	go build -mod vendor -o bin/sources cmd/sources/main.go

