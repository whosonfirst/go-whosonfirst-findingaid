// resolverd provides an HTTP server endpoint for resolving Who's On First URIs to their corresponding repository name
// using a go-whosonfirst-findingaid/resolver.Resolver instance.
package main

/*

$> java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb

$> ./bin/resolverd -resolver-uri 'awsdynamodb:///findingaid?region=local&endpoint=http://localhost:8000&credentials=static:local:local:local&partition_key=id'
2021/11/06 16:37:48 Listening for requests on http://localhost:8080

$> curl http://localhost:8080/1678780019
sfomuseum-data-flights-2018

$> ./bin/read -reader-uri 'findingaid://http/localhost:8080?template=https://raw.githubusercontent.com/sfomuseum-data/{repo}/main/data/' 85922583 | jq '.["properties"]["wof:name"]'
"San Francisco"

*/

import (
	"context"
	"github.com/aaronland/go-http-server"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/resolver"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"log"
	"net/http"
)

func handler(r resolver.Resolver) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		path := req.URL.Path

		id, _, err := uri.ParseURI(path)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		repo, err := r.GetRepo(ctx, id)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-Type", "text/plain")
		rsp.Write([]byte(repo))
	}

	h := http.HandlerFunc(fn)
	return h, nil
}

func main() {

	fs := flagset.NewFlagSet("resolver")

	server_uri := fs.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")
	resolver_uri := fs.String("resolver-uri", "", "...")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "RESOLVERD")

	if err != nil {
		log.Fatalf("Failed to set flags from environment variables, %v", err)
	}

	ctx := context.Background()

	r, err := resolver.NewResolver(ctx, *resolver_uri)

	if err != nil {
		log.Fatalf("Failed to create new resolver, %v", err)
	}

	h, err := handler(r)

	if err != nil {
		log.Fatalf("Failed to create new handler, %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	s, err := server.NewServer(ctx, *server_uri)

	if err != nil {
		log.Fatalf("Failed to create new server, %v", err)
	}

	log.Printf("Listening for requests on %s\n", s.Address())
	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		log.Fatalf("Failed to serve requests, %v", err)
	}

}
