package rmfetch

import (
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"os"
	"path/filepath"
	"time"
)

type Doc struct {
	Id   string
	Path string
	Mod  time.Time
}

type RMCloud struct {
	trans *transport.HttpClientCtx
	ctx   api.ApiCtx
}

func (rmc RMCloud) Docs() []Doc {
	tree := rmc.ctx.Filetree()
	refs := make(map[string]*model.Node)
	collectRefs(tree.Root(), refs)
	docs := make([]Doc, 0)
	for id, ref := range refs {
		path, _ := tree.NodeToPath(ref)
		mod, _ := ref.LastModified()
		docs = append(docs, Doc{
			Id:   id,
			Path: path,
			Mod:  mod,
		})
	}
	return docs
}

func (rmc RMCloud) Fetch(doc Doc) ([]byte, error) {
	tmpfile := filepath.Join(os.TempDir(), doc.Id+".zip")
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpfile)
	err := rmc.ctx.FetchDocument(doc.Id, tmpfile)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(tmpfile)
}

func collectRefs(curr *model.Node, docs map[string]*model.Node) {
	for _, node := range curr.Children {
		if node.IsDirectory() {
			collectRefs(node, docs)
		} else {
			docs[node.Id()] = node
		}
	}
}

type Config struct {
	OneTimeCode *string
}

func New(config Config) (*RMCloud, error) {
	trans, err := authHttpCtx(config.OneTimeCode)
	ctx, _, err := api.CreateApiCtx(trans)
	if err != nil {
		return nil, err
	}
	return &RMCloud{
		trans: trans,
		ctx:   ctx,
	}, nil
}
