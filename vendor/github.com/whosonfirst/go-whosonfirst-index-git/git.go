package git

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-index"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"net/url"
	"os"
	"path/filepath"
)

func init() {
	dr := NewGitDriver()
	index.Register("git", dr)
}

type GitDriver struct {
	index.Driver
	target   string
	preserve bool
}

func NewGitDriver() index.Driver {
	dr := &GitDriver{}
	return dr
}

func (d *GitDriver) Open(uri string) error {

	u, err := url.Parse(uri)

	if err != nil {
		return err
	}

	d.target = u.Path

	q := u.Query()

	if q.Get("preserve") == "1" {
		d.preserve = true
	}

	return nil
}

func (d *GitDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	var repo *gogit.Repository

	opts := &gogit.CloneOptions{
		URL: uri,
	}

	switch d.target {
	case "":

		r, err := gogit.Clone(memory.NewStorage(), nil, opts)

		if err != nil {
			return err
		}

		repo = r
	default:

		fname := filepath.Base(uri)
		path := filepath.Join(d.target, fname)

		r, err := gogit.PlainClone(path, false, opts)

		if err != nil {
			return err
		}

		if !d.preserve {
			defer os.RemoveAll(path)
		}

		repo = r
	}

	ref, err := repo.Head()

	if err != nil {
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())

	if err != nil {
		return err
	}

	tree, err := commit.Tree()

	if err != nil {
		return err
	}

	err = tree.Files().ForEach(func(f *object.File) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		switch filepath.Ext(f.Name) {
		case ".geojson":
			// continue
		default:
			return nil
		}

		fh, err := f.Reader()

		if err != nil {
			return err
		}

		defer fh.Close()

		ctx := index.AssignPathContext(ctx, f.Name)
		return index_cb(ctx, fh)
	})

	if err != nil {
		return err
	}

	return nil
}
