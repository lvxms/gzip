package gzip

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type gzipHandler struct {
	*Options
	gzPool           sync.Pool
	MinContentLength int //mdw 增加最小压缩限制
}

//mdw 增加最小压缩限制
func newGzipHandler(level int, minContentLength int, options ...Option) *gzipHandler {
	handler := &gzipHandler{
		Options:          DefaultOptions,
		MinContentLength: minContentLength, //mdw 最小压缩
		gzPool: sync.Pool{
			New: func() interface{} {
				gz, err := gzip.NewWriterLevel(ioutil.Discard, level)
				if err != nil {
					panic(err)
				}
				return gz
			},
		},
	}
	for _, setter := range options {
		setter(handler.Options)
	}
	return handler
}

func (g *gzipHandler) Handle(c *gin.Context) {
	if fn := g.DecompressFn; fn != nil && c.Request.Header.Get("Content-Encoding") == "gzip" {
		fn(c)
	}

	if !g.shouldCompress(c.Request) {
		return
	}

	gz := g.gzPool.Get().(*gzip.Writer)
	defer g.gzPool.Put(gz)
	defer gz.Reset(ioutil.Discard)
	gz.Reset(c.Writer)

	c.Header("Content-Encoding", "gzip")
	c.Header("Vary", "Accept-Encoding")
	c.Writer = &gzipWriter{c.Writer, gz, g.MinContentLength, 0, false} //mdw 增加最小压缩限制
	defer func() {
		var (
			size   int
			IsGzip bool
		)

		writer := (c.Writer).(*gzipWriter)

		if size, IsGzip = (writer.IsCompress()); !IsGzip {
			c.Writer = (c.Writer).(*gzipWriter).ResponseWriter
			writer.writer = nil
			c.Header("Content-Length", fmt.Sprint(size))
			return
		}

		gz.Close()
		//c.Header("Content-Length", fmt.Sprint(c.Writer.Size()))
		c.Header("Content-Length", fmt.Sprint(size))
	}()
	c.Next()
}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(req.URL.Path) {
		return false
	}
	if g.ExcludedPathesRegexs.Contains(req.URL.Path) {
		return false
	}

	return true
}
