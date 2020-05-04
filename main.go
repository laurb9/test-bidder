package main

import (
	"compress/flate"
	"compress/gzip"
	"log"
	"net/http"
	"strings"

	"github.com/laurb9/test-bidder/bidder"
	"github.com/laurb9/test-bidder/template"

	gorilla "github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
)

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// FIXME: docs
	w.WriteHeader(http.StatusOK)
}

// DecompressHandler is a wrapper to decompress gzip/deflate compressed requests
func DecompressHandler(h http.Handler) http.Handler {
	// FIXME: r.ContentLength will be incorrect unless we decompress all of it here.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origBody := r.Body
		switch strings.TrimSpace(r.Header.Get("Content-Encoding")) {
		case "gzip":
			r.Body, _ = gzip.NewReader(r.Body)
			defer func() {
				// FIXME: this does not work, Close() must have been called already
				if err := r.Body.Close(); err != nil {
					log.Printf("gzip decompress error: %s", err.Error())
				}
			}()
		case "deflate":
			r.Body = flate.NewReader(r.Body)
		}
		h.ServeHTTP(w, r)
		origBody.Close()
	})
}

func router() http.Handler {
	router := httprouter.New()
	router.GET("/", index)
	router.HandlerFunc("POST", "/b/openrtb/2.5", bidder.Handler)
	router.HandlerFunc("POST", "/t/openrtb/2.5", template.Handler)

	gzipHandler := DecompressHandler(
		gorilla.ContentTypeHandler(gorilla.CompressHandler(router), "application/json"),
	)

	return gzipHandler
}

func main() {
	port := ":8086"
	log.Printf("Starting server on %s", port)
	log.Fatal(http.ListenAndServe(port, router()))
}
