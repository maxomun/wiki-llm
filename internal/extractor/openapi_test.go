package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractOpenAPI_Success(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.0.3
info:
  title: Test API
  version: 1.2.3
  description: API de prueba
paths:
  /items/{id}:
    parameters:
      - name: traceId
        in: header
        required: false
        schema:
          type: string
    get:
      summary: Obtener item
      description: Retorna un item por id
      parameters:
        - name: id
          in: path
          schema:
            type: integer
        - name: q
          in: query
          required: false
          schema:
            type: string
        - name: traceId
          in: header
          required: true
          description: Correlacion
          schema:
            type: string
      requestBody:
        required: true
        description: cuerpo
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Req'
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "openapi.yaml")
	if err := os.WriteFile(sourcePath, []byte(spec), 0o644); err != nil {
		t.Fatalf("escribir spec temporal: %v", err)
	}

	doc, err := ExtractOpenAPI(sourcePath)
	if err != nil {
		t.Fatalf("ExtractOpenAPI devolvio error: %v", err)
	}

	if doc.Title != "Test API" {
		t.Fatalf("title inesperado: %q", doc.Title)
	}
	if doc.Version != "1.2.3" {
		t.Fatalf("version inesperada: %q", doc.Version)
	}
	if len(doc.Endpoints) != 1 {
		t.Fatalf("endpoints esperados=1 actual=%d", len(doc.Endpoints))
	}

	ep := doc.Endpoints[0]
	if ep.Method != "GET" || ep.Path != "/items/{id}" {
		t.Fatalf("endpoint inesperado: %s %s", ep.Method, ep.Path)
	}
	if ep.OperationID != "GET_items_id" {
		t.Fatalf("operationId fallback inesperado: %q", ep.OperationID)
	}
	if ep.RequestBody == nil || ep.RequestBody.SchemaRef != "#/components/schemas/Req" {
		t.Fatalf("requestBody no mapeado correctamente: %+v", ep.RequestBody)
	}
	if len(ep.Responses) != 1 || ep.Responses[0].SchemaRef != "#/components/schemas/Item" {
		t.Fatalf("responses no mapeadas correctamente: %+v", ep.Responses)
	}

	params := make(map[string]bool)
	for _, p := range ep.Parameters {
		params[p.Name] = p.Required
	}
	if len(params) != 3 {
		t.Fatalf("parametros esperados=3 actual=%d", len(params))
	}
	if !params["id"] {
		t.Fatalf("parametro path id debe ser required=true")
	}
	if !params["traceId"] {
		t.Fatalf("parametro traceId debe ser override required=true")
	}
	if params["q"] {
		t.Fatalf("parametro query q debe ser required=false")
	}
}

func TestExtractOpenAPI_MissingOpenAPIVersion(t *testing.T) {
	t.Parallel()

	spec := `info:
  title: Invalid
paths: {}`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(sourcePath, []byte(spec), 0o644); err != nil {
		t.Fatalf("escribir spec temporal: %v", err)
	}

	_, err := ExtractOpenAPI(sourcePath)
	if err == nil {
		t.Fatalf("se esperaba error por falta de campo openapi/swagger")
	}
}

func TestExtractOpenAPI_Swagger20JSON(t *testing.T) {
	t.Parallel()

	spec := `{
  "swagger": "2.0",
  "info": {"title":"Swagger API","version":"1.0"},
  "basePath": "/banco/api-cif/1.0",
  "paths": {
    "/clientes/{id}": {
      "get": {
        "operationId":"GetCliente",
        "responses": {"200":{"description":"OK"}}
      }
    }
  }
}`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "swagger.json")
	if err := os.WriteFile(sourcePath, []byte(spec), 0o644); err != nil {
		t.Fatalf("escribir swagger temporal: %v", err)
	}

	doc, err := ExtractOpenAPI(sourcePath)
	if err != nil {
		t.Fatalf("ExtractOpenAPI swagger error: %v", err)
	}
	if len(doc.Endpoints) != 1 {
		t.Fatalf("se esperaba 1 endpoint en swagger, actual=%d", len(doc.Endpoints))
	}
	if doc.Endpoints[0].BasePath != "/banco/api-cif/1.0" {
		t.Fatalf("basePath swagger inesperado: %s", doc.Endpoints[0].BasePath)
	}
}
