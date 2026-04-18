package normalizer

import "testing"

func TestMergeDocuments_MergesEndpointsAndMetadata(t *testing.T) {
	t.Parallel()

	openapiDoc := APIDocument{
		Title:       "API Base",
		Version:     "1.0.0",
		Description: "Desde OpenAPI",
		SourcePath:  "openapi.yaml",
		Endpoints: []Endpoint{
			{
				Path:        "/health",
				Method:      "GET",
				OperationID: "Health",
				Summary:     "Health",
				Responses:   []Response{{StatusCode: "200", Description: "OK"}},
			},
		},
	}

	postmanDoc := APIDocument{
		Title:      "Postman API",
		Version:    "postman",
		SourcePath: "collection.json",
		Endpoints: []Endpoint{
			{
				Path:        "/health",
				Method:      "GET",
				Description: "Endpoint de salud",
				Tags:        []string{"Postman"},
				Responses:   []Response{{StatusCode: "200", ContentTypes: []string{"application/json"}}},
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
	if len(health.Responses) != 1 || len(health.Responses[0].ContentTypes) != 1 {
		t.Fatalf("response enriquecida no fue fusionada correctamente")
	}
}
