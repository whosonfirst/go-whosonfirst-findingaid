package main

import (
	"context"
	"github.com/aaronland/go-http-server"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
	"github.com/whosonfirst/go-reader"		
	"flag"
	"log"
	go_http "net/http"
)

func main() {

	server_uri := flag.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")
	reader_uri := flag.String("reader-uri", "", "A valid whosonfirst/go-reader URI")
	
	flag.Parse()

	ctx := context.Background()

	r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}
	
	handler, err := http.LookupHandler(r)

	mux := go_http.NewServeMux()

	mux.Handle("/", handler)

	s, err := server.NewServer(ctx, *server_uri)

	if err != nil {
		log.Fatalf("Failed to create server, %v", err)
	}

	log.Printf("Listening on %s", s.Address())
	
	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}

