package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/tmsmr/rmfetch"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
)

type EnvSpec struct {
	RMApiDeviceCode string `envconfig:"RMAPI_DEVICE_CODE" required:"true"`
	RMRLAasBaseUrl  string `envconfig:"RMRL_BASE_URL"`
}

var env EnvSpec
var rmc *rmfetch.RMCloud

func init() {
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal(err)
	}
}

func refreshDocs(c *gin.Context) error {
	var err error
	rmc, err = rmfetch.New()
	if err != nil {
		rmc = nil
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return err
	}
	return nil
}

func listDocs(c *gin.Context) {
	if err := refreshDocs(c); err != nil {
		return
	}
	c.JSON(200, rmc.Docs())
}

func lookupDoc(c *gin.Context, id string) (*rmfetch.RMDoc, error) {
	docs := rmc.Docs()
	idx := slices.IndexFunc(docs, func(d rmfetch.RMDoc) bool { return d.Id == id })
	if idx == -1 {
		c.AbortWithStatus(http.StatusNotFound)
		return nil, errors.New("unknown doc")
	}
	return &docs[idx], nil
}

func fetchDoc(c *gin.Context) {
	if rmc == nil {
		if err := refreshDocs(c); err != nil {
			return
		}
	}
	id := c.Param("id")
	doc, err := lookupDoc(c, id)
	if err != nil {
		return
	}
	zip, err := rmc.Fetch(*doc)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+id+".zip")
	_, err = c.Writer.Write(zip)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
	}
}

func genPDF(c *gin.Context) {
	if env.RMRLAasBaseUrl == "" {
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	if rmc == nil {
		if err := refreshDocs(c); err != nil {
			return
		}
	}
	id := c.Param("id")
	doc, err := lookupDoc(c, id)
	if err != nil {
		return
	}
	pdf, err := rmc.GenPDF(*doc, env.RMRLAasBaseUrl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+id+".pdf")
	_, err = c.Writer.Write(pdf)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
	}
}

func main() {
	r := gin.Default()

	//r.StaticFile("/index.html", "./static/index.html")

	r.GET("/docs", listDocs)
	r.GET("/docs/:id/zip", fetchDoc)
	r.GET("/docs/:id/pdf", genPDF)

	_ = r.Run()
}
