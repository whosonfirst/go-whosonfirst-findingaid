package repo

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "log"
)

const WOF_ID_PROPERTY string = "wof:id"
const WOF_REPO_PROPERTY string = "wof:repo"

// FindingAidResonse is a struct that contains Who's On First repository information for Who's On First records.
type FindingAidResponse struct {
	// The unique Who's On First ID.
	ID int64 `json:"id"`
	// The name of the Who's On First repository.
	Repo string `json:"repo"`
	// The relative path for a Who's On First ID.
	URI string `json:"uri"`
}

func FindingAidResponseFromReader(ctx context.Context, fh io.Reader) (*FindingAidResponse, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	return FindingAidResponseFromBytes(ctx, body)
}

func FindingAidResponseFromBytes(ctx context.Context, body []byte) (*FindingAidResponse, error) {

	// TO DO: SUPPORT ALT FILES

	path_id := fmt.Sprintf("properties.%s", WOF_ID_PROPERTY)
	path_repo := fmt.Sprintf("properties.%s", WOF_REPO_PROPERTY)

	id_rsp := gjson.GetBytes(body, path_id)

	if !id_rsp.Exists() {
		return nil, fmt.Errorf("Missing '%s' property", path_id)
	}

	repo_rsp := gjson.GetBytes(body, path_repo)

	if !repo_rsp.Exists() {
		return nil, fmt.Errorf("Missing '%s' property", path_repo)
	}

	wof_id := id_rsp.Int()
	wof_repo := repo_rsp.String()

	rel_path, err := uri.Id2RelPath(wof_id)

	if err != nil {
		return nil, err
	}

	rsp := &FindingAidResponse{
		ID:   wof_id,
		Repo: wof_repo,
		URI:  rel_path,
	}

	return rsp, nil
}
