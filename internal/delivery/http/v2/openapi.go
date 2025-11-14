package v2

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	openAPISpec     []byte
	openAPISpecOnce sync.Once
	openAPISpecErr  error
)

func loadOpenAPISpec() ([]byte, error) {
	openAPISpecOnce.Do(func() {
		dir, err := os.Getwd()
		if err != nil {
			openAPISpecErr = err
			return
		}
		path := filepath.Join(dir, "docs", "openapi.yaml")
		openAPISpec, openAPISpecErr = os.ReadFile(path)
	})
	return openAPISpec, openAPISpecErr
}

func (h *Handler) serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec, err := loadOpenAPISpec()
	if err != nil {
		http.Error(w, "failed to load OpenAPI spec: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}
