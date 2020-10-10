package reader

import (
	"context"
	"errors"
	"github.com/tidwall/gjson"
	wof_reader "github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	"log"
)

type ReaderFindingAid struct {
	findingaid.FindingAid
	reader wof_reader.Reader
}

func NewReaderFindingAid(ctx context.Context, uri string) (findingaid.FindingAid, error) {
	return nil, errors.New("Not implemented")
}

func (fa *ReaderFindingAid) Index(ctx context.Context, sources ...string) error {
	return errors.New("Not implemented")
}

func (fa *ReaderFindingAid) IndexReader(ctx context.Context, fh io.Reader) error {
	return errors.New("Not implemented")
}

func (fa *ReaderFindingAid) LookupID(ctx context.Context, id int64, i interface{}) error {

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return err
	}

	fh, err := fa.reader.Read(ctx, rel_path)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return err
	}

	repo_rsp := gjson.GetBytes(body, "properties.wof:repo")

	if !repo_rsp.Exists() {
		return errors.New("Invalid WOF record")
	}

	repo := repo_rsp.String()

	fa_rsp := findingaid.FindingAidResponse{
		Id:   id,
		URI:  rel_path,
		Repo: repo,
	}

	log.Println(fa_rsp)
	return nil
}
