GOMOD=vendor

GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

proto:
	protoc -I=./producer/protobuf --go_out=./ ./producer/protobuf/findingaid.proto

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-populate cmd/wof-findingaid-populate/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-sources cmd/wof-findingaid-sources/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-csv2sql cmd/wof-findingaid-csv2sql/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-csv2docstore cmd/wof-findingaid-csv2docstore/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-create-dynamodb-tables cmd/wof-findingaid-create-dynamodb-tables/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-create-dynamodb-import cmd/wof-findingaid-create-dynamodb-import/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/wof-findingaid-resolverd cmd/wof-findingaid-resolverd/main.go


lambda:
	@make lambda-resolverd

lambda-resolverd:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f resolverd.zip; then rm -f resolverd.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="-s -w" -tags lambda.norpc -o bootstrap cmd/wof-findingaid-resolverd/main.go
	zip resolverd.zip bootstrap
	rm -f bootstrap

