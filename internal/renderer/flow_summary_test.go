package renderer

import (
	"strings"
	"testing"

	"github.com/max/wiki-llm/internal/normalizer"
)

func TestSummarizeFlow_UsesDetectedSignals(t *testing.T) {
	t.Parallel()

	endpoint := normalizer.Endpoint{
		Method: "GET",
		Implementation: &normalizer.ImplementationInfo{
			HandlerName:         "BuscarPorId",
			ServiceCalls:        []string{"service.BuscarPorId"},
			RepositoryCalls:     []string{"repo.BuscarPorId"},
			UsesDatabase:        true,
			DatabaseTypes:       []string{"sql"},
			DatabaseTables:      []string{"cif.tb_cliente"},
			DatabaseQueries:     []string{"SELECT * FROM cif.tb_cliente WHERE id = @id"},
			ExternalAPICalls:    []string{"client.Do"},
			UsesMessaging:       true,
			DatabaseCollections: []string{"clientes"},
		},
	}

	steps := summarizeFlow(endpoint)
	joined := strings.Join(steps, "\n")

	assertContains(t, joined, "Enruta la solicitud al handler: BuscarPorId")
	assertContains(t, joined, "Ejecuta una operacion de consulta sobre entidades:")
	assertContains(t, joined, "Orquesta logica de aplicacion via servicios: service.BuscarPorId")
	assertContains(t, joined, "Consulta repositorios: repo.BuscarPorId")
	assertContains(t, joined, "Ejecuta operaciones de base de datos")
	assertContains(t, joined, "tipo(s): sql")
	assertContains(t, joined, "tablas: cif.tb_cliente")
	assertContains(t, joined, "Aplica consultas SQL detectadas (consulta, 1 query(s)):")
	assertContains(t, joined, "Consume APIs externas: client.Do")
	assertContains(t, joined, "Publica eventos o mensajes")
	assertContains(t, joined, "Retorna respuesta al cliente")
}

func TestJoinLimited_TruncatesWithRemainder(t *testing.T) {
	t.Parallel()

	got := joinLimited([]string{"a", "b", "c", "d"}, 2)
	if got != "a, b (+2 mas)" {
		t.Fatalf("joinLimited inesperado: %q", got)
	}
}

func TestInferOperationType_UsesHTTPMethodFallback(t *testing.T) {
	t.Parallel()

	if got := inferOperationType("PATCH", nil); got != "actualizacion" {
		t.Fatalf("inferOperationType inesperado: %q", got)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("no contiene %q en %q", want, got)
	}
}
