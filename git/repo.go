package git

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-index"
	index_utils "github.com/whosonfirst/go-whosonfirst-index/utils"	
	"github.com/whosonfirst/go-cache"		
	_ "github.com/whosonfirst/go-whosonfirst-index-git"
	"io"
	"io/ioutil"	
	_ "log"
	"path/filepath"
	"strings"
	"strconv"
)

type RepoFindingAid struct {
	findingaid.FindingAid
	cache cache.Cache
}

func NewRepoFindingAid(ctx context.Context) (findingaid.FindingAid, error){

	c, err := cache.NewCache(ctx, "gocache://")

	if err != nil {
		return nil, err
	}

	return NewRepoFindingAidWithCache(ctx, c)
}

func NewRepoFindingAidWithCache(ctx context.Context, c cache.Cache) (findingaid.FindingAid, error) {

	fa := &RepoFindingAid{
		cache: c,
	}

	return fa, nil
}

func (fa *RepoFindingAid) Index(ctx context.Context, sources ...string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	remaining := len(sources)

	err_ch := make(chan error)
	done_ch := make(chan bool)
	
	for _, s := range sources {

		go func(ctx context.Context, source string) {

			err := fa.indexSource(ctx, source)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true
			
		}(ctx, s)
	}

	for remaining > 0 {
		select {
		case <- done_ch:
			remaining -= 1
		case err := <- err_ch:
			return err
		default:
			// pass
		}
	}

	return nil
}

func (fa *RepoFindingAid) indexSource(ctx context.Context, source string) error {

	repo_fname := filepath.Base(source)
	repo_ext := filepath.Ext(source)

	repo := strings.TrimRight(repo_fname, repo_ext)
	
	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		ok, err := index_utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
		
		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		feature_fname := filepath.Base(path)
		feature_ext := filepath.Ext(path)

		str_id := strings.TrimRight(feature_fname, feature_ext)
		_, err = strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			return err
		}

		sr := strings.NewReader(repo)
		sr_fh := ioutil.NopCloser(sr)
			
		_, err = fa.cache.Set(ctx, str_id, sr_fh)

		if err != nil {
			return err
		}
		
		return nil
	}

	idx, err := index.NewIndexer("git://", cb)

	if err != nil {
		return err
	}

	return idx.Index(ctx, source)
}

func (fa *RepoFindingAid) LookupID(ctx context.Context, id int64) (string, error) {

	str_id := strconv.FormatInt(id, 10)
	
	fh, err := fa.cache.Get(ctx, str_id)

	if err != nil {
		return "", err
	}

	defer fh.Close()
	
	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return "", err
	}

	return string(body), nil
}
