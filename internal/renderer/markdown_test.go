package renderer

import (
	"strings"
	"testing"

	"github.com/max/wiki-llm/internal/normalizer"
)

func TestRenderAPI_GeneratesExpectedFiles(t *testing.T) {
	t.Parallel()

	doc := normalizer.APIDocument{
		Title:       "My API",
		Version:     "0.1.0",
		Description: "Descripcion demo",
		SourcePath:  "./docs/openapi.yaml",
		Endpoints: []normalizer.Endpoint{
			{
				Path:        "/items/{id}",
				Method:      "GET",
				OperationID: "GetItem",
				Summary:     "Obtiene item",
				Description: "Busca item por id",
				Tags:        []string{"Items"},
				Parameters: []normalizer.Parameter{
					{Name: "id", In: "path", Required: true, Type: "integer", Description: "Id"},
				},
				RequestBody: &normalizer.RequestBody{
					Required:     false,
					ContentTypes: []string{"application/json"},
					SchemaRef:    "#/components/schemas/Req",
				},
				Responses: []normalizer.Response{
					{StatusCode: "200", Description: "OK", ContentTypes: []string{"application/json"}, SchemaRef: "#/components/schemas/Item"},
				},
			},
		},
	}

	files := RenderAPI(doc)
	if len(files) != 2 {
		t.Fatalf("cantidad de archivos inesperada: %d", len(files))
	}

	index, ok := files["index.md"]
	if !ok {
		t.Fatalf("falta index.md")
	}
	if !strings.Contains(index, "| Metodo | BasePath | Path | OperationId | Sources |") {
		t.Fatalf("index.md no contiene tabla de endpoints")
	}

	api, ok := files["apis/my-api.md"]
	if !ok {
		t.Fatalf("falta archivo de api renderizada")
	}
	requiredSnippets := []string{
		"## Tabla de contenidos",
		"### Parametros",
		"### Request Body",
		"### Responses",
		"| Status | Descripcion | Content Types | Schema |",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(api, snippet) {
			t.Fatalf("archivo renderizado no contiene fragmento requerido: %q", snippet)
		}
	}
}
