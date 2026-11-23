package web

import (
	"bytes"
	"html/template"
	"net/http"
	"time"
)

// builtAt is set by the Docker build linker flags.
var builtAt string

func serveIndex() http.HandlerFunc {
	var build struct {
		Timestamp string
		Label     string
	}
	if t, err := time.Parse(time.RFC3339, builtAt); err == nil {
		build.Timestamp = builtAt
		build.Label = t.In(time.FixedZone("JST", 9*60*60)).Format("2006-01-02 15:04:05 JST")
	}

	page := template.Must(template.ParseFS(assets, "static/index.html"))
	var body bytes.Buffer
	if err := page.Execute(&body, build); err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(body.Bytes()))
	}
}
