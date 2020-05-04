package main

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost/", nil)
	rr := httptest.NewRecorder()

	router := router()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGzipDecompress(t *testing.T) {
	var gzipData bytes.Buffer
	w, _ := gzip.NewWriterLevel(&gzipData, gzip.DefaultCompression)
	runCheckWithCompression(t, "gzip", w, &gzipData)
}

func TestDeflateDecompress(t *testing.T) {
	var compressedData bytes.Buffer
	w, _ := flate.NewWriter(&compressedData, flate.DefaultCompression)
	runCheckWithCompression(t, "deflate", w, &compressedData)
}

func runCheckWithCompression(t *testing.T, compression string, w io.WriteCloser, c io.Reader) {
	uncompressedData := "TEST DATA 123456"
	w.Write([]byte(uncompressedData))
	w.Close()

	req := httptest.NewRequest("GET", "http://localhost", c)

	req.Header.Set("Content-Encoding", compression)

	rr := httptest.NewRecorder()

	http.Handler(
		DecompressHandler(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					var buf bytes.Buffer
					buf.ReadFrom(r.Body)
					assert.Equal(t, uncompressedData, buf.String())
				},
			),
		),
	).ServeHTTP(rr, req)
}
