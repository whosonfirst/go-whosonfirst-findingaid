package http

import (
	"encoding/json"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "log"
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

		/*

> make debug
go run -mod vendor cmd/lookupd/main.go
2020/10/11 13:00:41 Listening on http://localhost:8080
2020/10/11 13:00:46 http: panic serving 127.0.0.1:62495: reflect: call of reflect.Value.FieldByName on interface Value
goroutine 50 [running]:
net/http.(*conn).serve.func1(0xc0000a8000)
	/usr/local/go/src/net/http/server.go:1800 +0x139
panic(0x142a200, 0xc0003de2a0)
	/usr/local/go/src/runtime/panic.go:975 +0x3e3
reflect.flag.mustBe(...)
	/usr/local/go/src/reflect/value.go:208
reflect.Value.FieldByName(0x14252e0, 0xc0000b0060, 0x194, 0x14a82fe, 0x2, 0x194, 0xc0004ca990, 0x0)
	/usr/local/go/src/reflect/value.go:887 +0x1ed
github.com/whosonfirst/go-whosonfirst-findingaid/repo.(*RepoFindingAid).LookupID(0xc0001e0150, 0x1575600, 0xc0000ae0c0, 0x51f1317, 0x13faee0, 0xc0000b0060, 0x0, 0x14a9b4f)
	/Users/asc/whosonfirst/go-whosonfirst-findingaid/repo/repo.go:249 +0x24b
github.com/whosonfirst/go-whosonfirst-findingaid/http.LookupHandler.func1(0x1574680, 0xc0000cc000, 0xc0000c2000)
	/Users/asc/whosonfirst/go-whosonfirst-findingaid/http/lookup.go:44 +0x172
net/http.HandlerFunc.ServeHTTP(0xc0001e4040, 0x1574680, 0xc0000cc000, 0xc0000c2000)

		*/
		
		fa_rsp, err := fa.Result(ctx)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusBadRequest)
			return
		}

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
