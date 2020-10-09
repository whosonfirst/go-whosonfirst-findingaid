package http

import (
	go_http "net/http"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/tidwall/gjson"
	"io/ioutil"
)

func LookupHandler(r reader.Reader) (go_http.Handler, error) {

	fn := func(rsp go_http.ResponseWriter, req *go_http.Request) {

		ctx := req.Context()
		path := req.URL.Path

		id, _, err := uri.ParseURI(path)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusBadRequest)
			return			
		}

		rel_path, err := uri.Id2RelPath(id)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusBadRequest)
			return			
		}

		fh, err := r.Read(ctx, rel_path)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusInternalServerError)			
			return			
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			go_http.Error(rsp, err.Error(), go_http.StatusInternalServerError)			
			return			
		}

		repo_rsp := gjson.GetBytes(body, "properties.wof:repo")

		if !repo_rsp.Exists(){
			go_http.Error(rsp, "Invalid WOF record", go_http.StatusInternalServerError)			
			return			
		}

		repo := repo_rsp.String()
		
		rsp.Write([]byte(repo))
		return
	}

	return go_http.HandlerFunc(fn), nil	
}
