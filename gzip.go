package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

func Gzip(level int, options ...Option) gin.HandlerFunc {
	//mdw 增加最小压缩限制
	return newGzipHandler(level, 0, options...).Handle
}

//mdw 增加最小压缩限制
func NewGzip(level int, minContentLength int, options ...Option) gin.HandlerFunc {
	return newGzipHandler(level, minContentLength, options...).Handle
}

type gzipWriter struct {
	gin.ResponseWriter
	writer           *gzip.Writer
	MinContentLength int //mdw 增加最小压缩限制
	ContentLength    int
	IsGzip           bool
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	//mdw 增加最小压缩限制
	g.ContentLength = len(s)
	if strings.ToLower(g.Header().Get("Transfer-Encoding")) == "chunked" || g.ContentLength < g.MinContentLength {
		g.IsGzip = false
		g.Header().Del("Content-Length")
		g.Header().Del("Content-Encoding")
		g.Header().Del("Vary")
		return g.ResponseWriter.Write([]byte(s))
	}

	g.IsGzip = true
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	//mdw 增加最小压缩限制
	g.ContentLength = len(data)
	if strings.ToLower(g.Header().Get("Transfer-Encoding")) == "chunked" || g.ContentLength < g.MinContentLength {
		g.IsGzip = false
		g.Header().Del("Content-Length")
		g.Header().Del("Content-Encoding")
		g.Header().Del("Vary")
		return g.ResponseWriter.Write(data)
	}

	g.IsGzip = true
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

//mdw 增加最小压缩限制
func (g *gzipWriter) Size() int {
	if g.ContentLength < g.MinContentLength {
		return g.ContentLength
	}

	return g.ResponseWriter.Size()
}

//mdw add
func (g *gzipWriter) IsCompress() (int, bool) {
	var size int

	if g.ContentLength < g.MinContentLength {
		size = g.ContentLength
	}

	size = g.ResponseWriter.Size()

	return size, g.IsGzip
}

//mdw add
func (g *gzipWriter) Header() http.Header {
	return g.ResponseWriter.Header()
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}
