package http

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"log"
	go_http "net/http"
)

func LookupHandler(fa findingaid.FindingAid) (go_http.Handler, error) {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {

		ctx := req.Context()
		path := req.URL.Path

		if path == "/" {

			err := usage(rsp, req)

			if err != nil {
				go_http.Error(rsp, err.Error(), go_http.StatusInternalServerError)
				return
			}

			return
		}

		id, _, err := uri.ParseURI(path)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusBadRequest)
			return
		}

		var fa_rsp repo.FindingAidResponse

		log.Println("LOOKUP", id)

		err = fa.LookupID(ctx, id, &fa_rsp)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusBadRequest)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(fa_rsp)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusInternalServerError)
			return
		}

		return
	}

	return go_http.HandlerFunc(fn), nil
}

func usage(rsp go_http.ResponseWriter, req *go_http.Request) error {

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
