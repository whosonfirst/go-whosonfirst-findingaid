proto:
	protoc -I=./producer/protobuf --go_out=./ ./producer/protobuf/findingaid.proto

cli:
	go build -mod vendor -o bin/populate cmd/populate/main.go
	go build -mod vendor -o bin/sources cmd/sources/main.go
	go build -mod vendor -o bin/csv2sql cmd/csv2sql/main.go
	go build -mod vendor -o bin/csv2docstore cmd/csv2docstore/main.go
	go build -mod vendor -o bin/create-dynamodb-tables cmd/create-dynamodb-tables/main.go
	go build -mod vendor -o bin/resolverd cmd/resolverd/main.go


lambda:
	@make lambda-resolverd

lambda-resolverd:
	if test -f main; then rm -f main; fi
	if test -f resolverd.zip; then rm -f resolverd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/resolverd/main.go
	zip resolverd.zip main
	rm -f main

