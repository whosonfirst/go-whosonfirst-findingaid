package main

import (
	"context"
	"flag"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
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

	cors_handler := cors.New(cors.Options{})

	lookup_handler, err := http.LookupHandler(r)

	if err != nil {
		log.Fatalf("Failed to create lookup handler, %v", err)
	}

	lookup_handler = cors_handler.Handler(lookup_handler)

	mux := go_http.NewServeMux()

	mux.Handle("/", lookup_handler)

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
