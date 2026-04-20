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

func TestNormalizeDocument_DeduplicatesEquivalentPathParams(t *testing.T) {
	t.Parallel()

	doc := APIDocument{
		Endpoints: []Endpoint{
			{
				Method: "GET",
				Path:   "/clientes/{id}",
				Parameters: []Parameter{
					{Name: "id", In: "path", Required: true, Type: "string"},
					{Name: "id_cliente", In: "path", Required: true, Description: "ID del cliente"},
				},
			},
		},
	}

	out := NormalizeDocument(doc)
	params := out.Endpoints[0].Parameters
	if len(params) != 1 {
		t.Fatalf("se esperaba 1 parametro path deduplicado, actual=%d", len(params))
	}
	if params[0].Name != "id" {
		t.Fatalf("nombre path param inesperado: %s", params[0].Name)
	}
	if params[0].Description == "" {
		t.Fatalf("se esperaba conservar descripcion al deduplicar")
	}
}

func TestNormalizeDocument_CanonicalizesPathParamsByPathStructure(t *testing.T) {
	t.Parallel()

	doc := APIDocument{
		Endpoints: []Endpoint{
			{
				Method: "GET",
				Path:   "/clientes/{clienteid}/cuentas/{cuentaid}",
				Parameters: []Parameter{
					{Name: "id_cliente", In: "path", Required: true},
					{Name: "id_cuenta", In: "path", Required: true},
				},
			},
		},
	}

	out := NormalizeDocument(doc)
	params := out.Endpoints[0].Parameters
	if len(params) != 2 {
		t.Fatalf("se esperaban 2 params path, actual=%d", len(params))
	}
	if params[0].Name != "clienteid" || params[1].Name != "cuentaid" {
		t.Fatalf("parametros no alineados al path consolidado: %+v", params)
	}
}
