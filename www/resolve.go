package www

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	_ "log"
	"net/http"
)

func ResolveHandler(fa findingaid.Resolver) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		path := req.URL.Path

		if path == "/" {

			err := usage(rsp, req)

			if err != nil {
				http.Error(rsp, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		fa_rsp, err := fa.ResolveURI(ctx, path)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(fa_rsp)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}

func usage(rsp http.ResponseWriter, req *http.Request) error {

	rsp.Header().Set("Content-type", "text/html")

	rsp.Write([]byte(`<!DOCTYPE html>
<html>
 <head>
   <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
   <title>Who's On First Finding Aid</title>
 </head>
 <body>
 <h1>Who's On First Finding Aid</h1>
 <p>For example:</p>
 <pre>
$> curl -s <a href="https://data.whosonfirst.org/findingaid/85633041">https://data.whosonfirst.org/findingaid/85633041</a> | jq
{
  "id": 85633041,
  "uri": "856/330/41/85633041.geojson",
  "repo": "whosonfirst-data-admin-ca"
}
 </pre>
 <p style="font-style:italic;">If you're reading this there may still be some missing data that is in the process of being updated.</p>
</body>`))

	return nil
}
