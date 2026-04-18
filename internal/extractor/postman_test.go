package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPostmanCollection_Success(t *testing.T) {
	t.Parallel()

	collection := `{
  "info": {
    "name": "Postman Demo API",
    "description": "Coleccion demo",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Folder A",
      "item": [
        {
          "name": "Get Item",
          "request": {
            "method": "GET",
            "description": "Consulta item",
            "url": {
              "path": ["api", "items", ":id"],
              "query": [{"key": "q", "description": "Busqueda"}],
              "variable": [{"key": "id", "description": "identificador"}]
            }
          },
          "response": []
        }
      ]
    }
  ]
}`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "collection.json")
	if err := os.WriteFile(sourcePath, []byte(collection), 0o644); err != nil {
		t.Fatalf("escribir collection temporal: %v", err)
	}

	doc, err := ExtractPostmanCollection(sourcePath)
	if err != nil {
		t.Fatalf("ExtractPostmanCollection devolvio error: %v", err)
	}

	if doc.Title != "Postman Demo API" {
		t.Fatalf("title inesperado: %q", doc.Title)
	}
	if len(doc.Endpoints) != 1 {
		t.Fatalf("endpoints esperados=1 actual=%d", len(doc.Endpoints))
	}
	ep := doc.Endpoints[0]
	if ep.Method != "GET" {
		t.Fatalf("metodo inesperado: %s", ep.Method)
	}
	if ep.Path != "/api/items/{id}" {
		t.Fatalf("path inesperado: %s", ep.Path)
	}
	if len(ep.Tags) != 1 || ep.Tags[0] != "Folder A" {
		t.Fatalf("tags inesperados: %+v", ep.Tags)
	}
	if len(ep.Parameters) != 2 {
		t.Fatalf("parametros esperados=2 actual=%d", len(ep.Parameters))
	}
}
