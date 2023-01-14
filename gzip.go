package gzip

import (
	"compress/gzip"
	"fmt"

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
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	//mdw 增加最小压缩限制
	g.ContentLength = len(s)
	if g.ContentLength < g.MinContentLength {
		g.Header().Del("Content-Encoding")
		g.Header().Del("Vary")
		return g.ResponseWriter.Write([]byte(s))
	}

	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	//mdw 增加最小压缩限制
	g.ContentLength = len(data)
	if g.ContentLength < g.MinContentLength {
		g.Header().Del("Content-Encoding")
		g.Header().Del("Vary")
		return g.ResponseWriter.Write(data)
	}

	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

//mdw 增加最小压缩限制
func (g *gzipWriter) Size() int {
	if g.ContentLength < g.MinContentLength {
		g.Header().Set("Content-Length", fmt.Sprint(g.ContentLength))
		return g.ContentLength
	}

	return g.ResponseWriter.Size()
}

// Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}
