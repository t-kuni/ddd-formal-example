// Package api contains the HTTP contract and its code generation entry point.
package api

import (
	"bytes"
	_ "embed"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:generate go tool oapi-codegen -config oapi-codegen.yaml -o ../internal/web/contract/api.gen.go openapi.yaml

//go:embed openapi.yaml
var spec []byte

func Load() (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromData(spec)
}

// OpenAPI serves the same document used for validation and code generation.
func OpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeContent(w, r, "openapi.yaml", time.Time{}, bytes.NewReader(spec))
}
