package normalizer

import "testing"

func TestNormalizeDocument_ReplacesDynamicSegmentsAndDetectsParams(t *testing.T) {
	t.Parallel()

	doc := APIDocument{
		Endpoints: []Endpoint{
			{
				Method: "get",
				Path:   "/transferencias/enviadas/530c3676-adac-426a-b4b4-94a0a754544d",
			},
			{
				Method: "get",
				Path:   "/clientes/12345",
			},
			{
				Method: "get",
				Path:   "/users/:userId/orders/:orderId",
			},
		},
	}

	out := NormalizeDocument(doc)
	if out.Endpoints[0].Path != "/transferencias/enviadas/{id}" {
		t.Fatalf("path uuid no normalizado: %s", out.Endpoints[0].Path)
	}
	if out.Endpoints[1].Path != "/clientes/{id}" {
		t.Fatalf("path numerico no normalizado: %s", out.Endpoints[1].Path)
	}
	if out.Endpoints[2].Path != "/users/{userid}/orders/{orderid}" {
		t.Fatalf("path params estilo colon no normalizado: %s", out.Endpoints[2].Path)
	}
	if len(out.Endpoints[2].PathParams) != 2 {
		t.Fatalf("path params esperados=2 actual=%d", len(out.Endpoints[2].PathParams))
	}
}

func TestNormalizeDocument_SplitsBasePath(t *testing.T) {
	t.Parallel()

	doc := APIDocument{
		Endpoints: []Endpoint{
			{
				Method: "GET",
				Path:   "/banco/api-cliente-tef/1.0/transferencias",
			},
		},
	}

	out := NormalizeDocument(doc)
	ep := out.Endpoints[0]
	if ep.BasePath != "/banco/api-cliente-tef/1.0" {
		t.Fatalf("basePath inesperado: %s", ep.BasePath)
	}
	if ep.Path != "/transferencias" {
		t.Fatalf("path inesperado: %s", ep.Path)
	}
}
