cli:
	go build -mod vendor -o bin/lookupd cmd/lookupd/main.go

debug:
	go run -mod vendor cmd/lookupd/main.go

lambda-handlers:
	@make lambda-server

lambda-server:	
	if test -f main; then rm -f main; fi
	if test -f lookupd.zip; then rm -f lookupd.zip; fi
	GOOS=linux go build -mod vendor -o main cmd/lookupd/main.go
	zip lookupd.zip main
	rm -f main
