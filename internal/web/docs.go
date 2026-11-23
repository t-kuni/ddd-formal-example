package web

import (
	"bytes"
	"compress/gzip"
	"embed"
	"io"
	"net/http"
	"strings"
	"time"

	"wholesale.local/orbit/api"
)

//go:embed docs/*
var docs embed.FS

func registerDocs(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", api.OpenAPI)
	mux.HandleFunc("GET /docs", serveDocsFile("index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /docs/reference.js", serveDocsFile("reference.js", "text/javascript; charset=utf-8"))
	mux.HandleFunc("GET /docs/scalar-1.83.0.js", serveScalarBundle)
}

func serveDocsFile(name, contentType string) http.HandlerFunc {
	body, err := docs.ReadFile("docs/" + name)
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		if name == "index.html" {
			// Scalar injects its component styles at runtime.
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; worker-src 'self' blob:; object-src 'none'; frame-ancestors 'none'")
		}
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
	}
}

func serveScalarBundle(w http.ResponseWriter, r *http.Request) {
	body, err := docs.ReadFile("docs/scalar-1.83.0.js.gz")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")
	for _, coding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(coding), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			http.ServeContent(w, r, "scalar-1.83.0.js", time.Time{}, bytes.NewReader(body))
			return
		}
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Scalar bundle unavailable", http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	_, _ = io.Copy(w, reader)
}
