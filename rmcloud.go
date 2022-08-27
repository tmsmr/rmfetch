package rmfetch

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/model"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrMissingCode = errors.New("missing required one-time code")
	ErrPDFGen      = errors.New("failed to generate the PDF")
)

type RMDoc struct {
	Id   string    `json:"id"`
	Path string    `json:"path"`
	Mod  time.Time `json:"mod"`
}

type RMCloud struct {
	ctx api.ApiCtx
}

func New() (*RMCloud, error) {
	if _, err := os.Stat(config.ConfigPath()); err != nil {
		if code, ok := os.LookupEnv("RMAPI_DEVICE_CODE"); !ok || len(code) != 8 {
			return nil, ErrMissingCode
		}
	}
	trans := api.AuthHttpCtx(false, true)
	ctx, _, err := api.CreateApiCtx(trans)
	if err != nil {
		return nil, err
	}
	return &RMCloud{
		ctx: ctx,
	}, nil
}

func (rmc RMCloud) Docs() []RMDoc {
	tree := rmc.ctx.Filetree()
	refs := make(map[string]*model.Node)
	collectRefs(tree.Root(), refs)
	docs := make([]RMDoc, 0)
	for id, ref := range refs {
		path, _ := tree.NodeToPath(ref)
		mod, _ := ref.LastModified()
		docs = append(docs, RMDoc{
			Id:   id,
			Path: path,
			Mod:  mod,
		})
	}
	return docs
}

func (rmc RMCloud) Fetch(doc RMDoc) ([]byte, error) {
	tmp := filepath.Join(os.TempDir(), doc.Id+".zip")
	defer func(name string) {
		_ = os.Remove(name)
	}(tmp)
	err := rmc.ctx.FetchDocument(doc.Id, tmp)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(tmp)
}

func (rmc RMCloud) GenPDF(doc RMDoc, rmrlAas string) ([]byte, error) {
	zip, err := rmc.Fetch(doc)
	var post bytes.Buffer
	mpw := multipart.NewWriter(&post)
	ff, err := mpw.CreateFormFile("file", doc.Id+".zip")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(ff, bytes.NewReader(zip)); err != nil {
		return nil, err
	}
	if err = mpw.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/render", rmrlAas),
		&post)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mpw.FormDataContentType())
	client := http.Client{
		Timeout: time.Minute,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, ErrPDFGen
	}
	var pdf bytes.Buffer
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	_, err = pdf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}
	return pdf.Bytes(), nil
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
