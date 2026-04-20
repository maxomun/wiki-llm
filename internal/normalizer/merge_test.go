package normalizer

import "testing"

func TestMergeDocuments_MergesEndpointsAndMetadata(t *testing.T) {
	t.Parallel()

	openapiDoc := APIDocument{
		Title:          "API Base",
		Version:        "1.0.0",
		Description:    "Desde OpenAPI",
		BasePath:       "/banco/api-cif",
		ContractSource: SourceOpenAPI,
		SourcePath:     "openapi.yaml",
		Endpoints: []Endpoint{
			{
				Path:        "/health",
				Method:      "GET",
				OperationID: "Health",
				Summary:     "Health from OpenAPI",
				Description: "Descripcion OpenAPI",
				RequestBody: &RequestBody{
					Required:     true,
					Description:  "Body OpenAPI",
					SchemaRef:    "#/components/schemas/HealthReq",
					ContentTypes: []string{"application/json"},
				},
				Responses: []Response{{StatusCode: "200", Description: "OK"}},
			},
		},
	}

	postmanDoc := APIDocument{
		Title:          "Postman API",
		Version:        "",
		ContractSource: SourcePostman,
		SourcePath:     "collection.json",
		Endpoints: []Endpoint{
			{
				Path:        "/health",
				Method:      "GET",
				Summary:     "Health from Postman",
				Description: "Endpoint de salud",
				Tags:        []string{"Postman"},
				Sources:     []SourceType{SourcePostman},
				RequestBody: &RequestBody{
					Example: `{"ping":"pong"}`,
				},
				Responses: []Response{{StatusCode: "200", ContentTypes: []string{"application/json"}, Example: `{"ok":true}`}},
			},
			{
				Path:        "/ready",
				Method:      "GET",
				OperationID: "Ready",
			},
		},
	}

	merged := MergeDocuments([]APIDocument{openapiDoc, postmanDoc})
	if merged.Title != "API Base" {
		t.Fatalf("title inesperado: %q", merged.Title)
	}
	if merged.Version != "1.0.0" {
		t.Fatalf("version inesperada: %q", merged.Version)
	}
	if merged.BasePath != "/banco/api-cif" {
		t.Fatalf("basePath inesperado: %q", merged.BasePath)
	}
	if merged.ContractSource != SourceOpenAPI {
		t.Fatalf("contract source inesperado: %q", merged.ContractSource)
	}
	if len(merged.Endpoints) != 2 {
		t.Fatalf("endpoints esperados=2 actual=%d", len(merged.Endpoints))
	}

	var health Endpoint
	for _, ep := range merged.Endpoints {
		if ep.Method == "GET" && ep.Path == "/health" {
			health = ep
			break
		}
	}
	if health.Description == "" {
		t.Fatalf("descripcion enriquecida no fue fusionada")
	}
	if health.Summary != "Health from OpenAPI" {
		t.Fatalf("summary no debe ser sobrescrito por postman: %q", health.Summary)
	}
	if len(health.Responses) != 1 || len(health.Responses[0].ContentTypes) != 1 {
		t.Fatalf("response enriquecida no fue fusionada correctamente")
	}
	if health.RequestBody == nil || health.RequestBody.SchemaRef != "#/components/schemas/HealthReq" {
		t.Fatalf("request body estructural debe mantenerse desde openapi: %+v", health.RequestBody)
	}
	if health.RequestBody.Example == "" {
		t.Fatalf("request body example desde postman no fue aplicado")
	}
	if health.Responses[0].Example == "" {
		t.Fatalf("response example desde postman no fue aplicado")
	}
	if health.Confidence != ConfidenceHigh {
		t.Fatalf("confidence inesperado para endpoint consolidado: %q", health.Confidence)
	}
}

func TestMergeDocuments_MergesEndpointByStructuralPath(t *testing.T) {
	t.Parallel()

	openapiDoc := APIDocument{
		ContractSource: SourceOpenAPI,
		Endpoints: []Endpoint{
			{
				Path:    "/clientes/{id}",
				Method:  "GET",
				Summary: "Desde OpenAPI",
				Parameters: []Parameter{
					{Name: "id", In: "path", Required: true},
				},
			},
		},
	}

	postmanDoc := APIDocument{
		ContractSource: SourcePostman,
		Endpoints: []Endpoint{
			{
				Path:    "/clientes/{id_cliente}",
				Method:  "GET",
				Summary: "Desde Postman",
				Parameters: []Parameter{
					{Name: "id_cliente", In: "path", Required: true},
				},
				Sources: []SourceType{SourcePostman},
			},
		},
	}

	merged := MergeDocuments([]APIDocument{openapiDoc, postmanDoc})
	if len(merged.Endpoints) != 1 {
		t.Fatalf("se esperaba un endpoint consolidado, actual=%d", len(merged.Endpoints))
	}
	if merged.Endpoints[0].Path != "/clientes/{id}" {
		t.Fatalf("path consolidado inesperado: %q", merged.Endpoints[0].Path)
	}
	if len(merged.Endpoints[0].Parameters) != 1 || merged.Endpoints[0].Parameters[0].Name != "id" {
		t.Fatalf("parametros path no consolidados estructuralmente: %+v", merged.Endpoints[0].Parameters)
	}
}
