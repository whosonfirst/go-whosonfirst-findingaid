proto:
	protoc -I=./producer/protobuf --go_out=./ ./producer/protobuf/findingaid.proto

cli:
	go build -mod vendor -o bin/populate cmd/populate/main.go
	go build -mod vendor -o bin/sources cmd/sources/main.go
	go build -mod vendor -o bin/csv2sql cmd/csv2sql/main.go
	go build -mod vendor -o bin/create-dynamodb-tables cmd/create-dynamodb-tables/main.go
