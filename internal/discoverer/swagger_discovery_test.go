package discoverer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max/wiki-llm/internal/extractor"
)

func TestFindSwaggerJSON_Priority(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docsPath := filepath.Join(root, "docs")
	nestedDocs := filepath.Join(root, "module", "docs")
	if err := os.MkdirAll(docsPath, 0o755); err != nil {
		t.Fatalf("crear docs: %v", err)
	}
	if err := os.MkdirAll(nestedDocs, 0o755); err != nil {
		t.Fatalf("crear nested docs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "swagger.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("crear swagger raiz: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDocs, "swagger.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("crear swagger nested: %v", err)
	}
	expected := filepath.Join(docsPath, "swagger.json")
	if err := os.WriteFile(expected, []byte("{}"), 0o644); err != nil {
		t.Fatalf("crear swagger docs: %v", err)
	}

	got, err := FindSwaggerJSON(root)
	if err != nil {
		t.Fatalf("FindSwaggerJSON error: %v", err)
	}
	if got != expected {
		t.Fatalf("swagger elegido inesperado: got=%s want=%s", got, expected)
	}
}

func TestResolveGenerateAPISources_UsesCodeWhenNoOpenAPIInSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	docsPath := filepath.Join(root, "docs")
	if err := os.MkdirAll(docsPath, 0o755); err != nil {
		t.Fatalf("crear docs: %v", err)
	}
	swaggerPath := filepath.Join(docsPath, "swagger.json")
	if err := os.WriteFile(swaggerPath, []byte(`{"swagger":"2.0","info":{"title":"x"},"paths":{}}`), 0o644); err != nil {
		t.Fatalf("crear swagger: %v", err)
	}
	postmanPath := filepath.Join(root, "collection.json")
	if err := os.WriteFile(postmanPath, []byte(`{"info":{"name":"x","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[]}`), 0o644); err != nil {
		t.Fatalf("crear postman: %v", err)
	}

	sources, logs, err := ResolveGenerateAPISources(root, []string{postmanPath}, extractor.SourceTypeAuto)
	if err != nil {
		t.Fatalf("ResolveGenerateAPISources error: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("sources esperados=2 actual=%d", len(sources))
	}
	if sources[0].SourceType != extractor.SourceTypeOpenAPI {
		t.Fatalf("primer source debe ser openapi: %+v", sources[0])
	}
	if sources[1].SourceType != extractor.SourceTypePostman {
		t.Fatalf("segundo source debe ser postman: %+v", sources[1])
	}
	if len(logs) == 0 {
		t.Fatalf("se esperaban logs de descubrimiento")
	}
}
