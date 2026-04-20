package sourcecode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max/wiki-llm/internal/normalizer"
)

func TestEnrichDocument_MapsEndpointToHandler(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	routerFile := filepath.Join(root, "router.go")
	handlerFile := filepath.Join(root, "cliente_handler.go")

	routerContent := `package api
func setup(router *Router, h *ClienteHandler) {
	router.GET("/clientes/:id", h.BuscarPorId)
}`
	handlerContent := `package api
func (h *ClienteHandler) BuscarPorId() {
	h.service.BuscarPorId()
	h.repo.GetByID()
	qry := ` + "`UPDATE cif.tb_telefono SET telefono = 'x' WHERE id = $1`" + `
	h.db.Query("select * from clientes where id = $1")
	h.db.Exec(qry)
	h.db.Exec("execute sp_actualiza_cliente @id = 1")
	h.mongo.Collection("clientes").FindOne(nil, nil)
}`

	if err := os.WriteFile(routerFile, []byte(routerContent), 0o644); err != nil {
		t.Fatalf("escribir router.go: %v", err)
	}
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0o644); err != nil {
		t.Fatalf("escribir handler.go: %v", err)
	}

	doc := normalizer.APIDocument{
		Title: "API",
		Endpoints: []normalizer.Endpoint{
			{Method: "GET", BasePath: "/banco/api-cif/1.0", Path: "/clientes/{id}"},
		},
	}

	out, err := EnrichDocument(doc, root)
	if err != nil {
		t.Fatalf("EnrichDocument error: %v", err)
	}
	if len(out.Endpoints) != 1 {
		t.Fatalf("endpoints esperados=1 actual=%d", len(out.Endpoints))
	}
	impl := out.Endpoints[0].Implementation
	if impl == nil {
		t.Fatalf("se esperaba implementation metadata")
	}
	if impl.HandlerName != "BuscarPorId" {
		t.Fatalf("handler inesperado: %s", impl.HandlerName)
	}
	if !impl.UsesDatabase {
		t.Fatalf("se esperaba deteccion de uso de base de datos")
	}
	if len(impl.ServiceCalls) == 0 || len(impl.RepositoryCalls) == 0 {
		t.Fatalf("se esperaban llamadas a servicio y repositorio")
	}
	if len(impl.DatabaseTypes) == 0 {
		t.Fatalf("se esperaban tipos de base de datos detectados")
	}
	if len(impl.DatabaseTables) == 0 || !contains(impl.DatabaseTables, "clientes") {
		t.Fatalf("tabla no detectada correctamente: %+v", impl.DatabaseTables)
	}
	if len(impl.DatabaseSPs) == 0 {
		t.Fatalf("stored procedure no detectado")
	}
	if len(impl.DatabaseCollections) == 0 || impl.DatabaseCollections[0] != "clientes" {
		t.Fatalf("collection no detectada correctamente: %+v", impl.DatabaseCollections)
	}
	if len(impl.DatabaseQueries) == 0 {
		t.Fatalf("queries no detectadas")
	}
}

func TestEnrichDocument_InspectsRepositoryLayer(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	routerFile := filepath.Join(root, "router.go")
	handlerFile := filepath.Join(root, "cliente_handler.go")
	repoFile := filepath.Join(root, "cliente_repo.go")

	routerContent := `package api
func setup(router *Router, h *ClienteHandler) {
	router.GET("/clientes/:id", h.BuscarPorId)
}`
	handlerContent := `package api
func (h *ClienteHandler) BuscarPorId() {
	h.repo.GetByID()
}`
	repoContent := `package api
func (r *ClienteRepo) GetByID() {
	r.db.Query("select * from clientes where id = $1")
}`

	if err := os.WriteFile(routerFile, []byte(routerContent), 0o644); err != nil {
		t.Fatalf("escribir router.go: %v", err)
	}
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0o644); err != nil {
		t.Fatalf("escribir handler.go: %v", err)
	}
	if err := os.WriteFile(repoFile, []byte(repoContent), 0o644); err != nil {
		t.Fatalf("escribir repo.go: %v", err)
	}

	doc := normalizer.APIDocument{
		Endpoints: []normalizer.Endpoint{
			{Method: "GET", Path: "/clientes/{id}"},
		},
	}
	out, err := EnrichDocument(doc, root)
	if err != nil {
		t.Fatalf("EnrichDocument error: %v", err)
	}
	impl := out.Endpoints[0].Implementation
	if impl == nil {
		t.Fatalf("se esperaba implementation metadata")
	}
	if !impl.UsesDatabase {
		t.Fatalf("se esperaba deteccion de base de datos desde capa repository")
	}
	if len(impl.DatabaseTables) == 0 || !contains(impl.DatabaseTables, "clientes") {
		t.Fatalf("tabla esperada no detectada: %+v", impl.DatabaseTables)
	}
}

func TestEnrichDocument_MatchesRouteByStructuralPlaceholder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	routerFile := filepath.Join(root, "router.go")
	handlerFile := filepath.Join(root, "cliente_handler.go")

	routerContent := `package api
func setup(router *Router, h *ClienteHandler) {
	router.GET("/clientes/:id", h.BuscarPorId)
}`
	handlerContent := `package api
func (h *ClienteHandler) BuscarPorId() {
	h.repo.GetByID()
}
func (r *ClienteRepo) GetByID() {
	r.db.Query("select * from clientes where id = $1")
}`

	if err := os.WriteFile(routerFile, []byte(routerContent), 0o644); err != nil {
		t.Fatalf("escribir router.go: %v", err)
	}
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0o644); err != nil {
		t.Fatalf("escribir handler.go: %v", err)
	}

	doc := normalizer.APIDocument{
		Endpoints: []normalizer.Endpoint{
			{Method: "GET", Path: "/clientes/{id_cliente}"},
		},
	}
	out, err := EnrichDocument(doc, root)
	if err != nil {
		t.Fatalf("EnrichDocument error: %v", err)
	}
	impl := out.Endpoints[0].Implementation
	if impl == nil {
		t.Fatalf("se esperaba implementation metadata")
	}
	if impl.HandlerName != "BuscarPorId" {
		t.Fatalf("handler no detectado via matching estructural: %q", impl.HandlerName)
	}
	if !impl.UsesDatabase {
		t.Fatalf("se esperaba detectar base de datos de forma indirecta")
	}
}

func TestEnrichDocument_TracesServiceRepositoryAndHelperCalls(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	routerFile := filepath.Join(root, "router.go")
	handlerFile := filepath.Join(root, "cliente_handler.go")
	serviceFile := filepath.Join(root, "cliente_service.go")
	repoFile := filepath.Join(root, "cliente_repo.go")

	routerContent := `package api
func setup(router *Router, h *ClienteHandler) {
	router.GET("/clientes/:id", h.BuscarPorId)
}`
	handlerContent := `package api
func (h *ClienteHandler) BuscarPorId() {
	h.service.BuscarPorId()
}`
	serviceContent := `package api
func (s *ClienteService) BuscarPorId() {
	enviarMetricas()
	s.repo.GetByID()
}
func enviarMetricas() {
	publisher.Publish("cliente.consultado")
}`
	repoContent := `package api
func (r *ClienteRepo) GetByID() {
	client.Do(req)
	r.db.Query("select * from clientes where id = $1")
}`

	if err := os.WriteFile(routerFile, []byte(routerContent), 0o644); err != nil {
		t.Fatalf("escribir router.go: %v", err)
	}
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0o644); err != nil {
		t.Fatalf("escribir handler.go: %v", err)
	}
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0o644); err != nil {
		t.Fatalf("escribir service.go: %v", err)
	}
	if err := os.WriteFile(repoFile, []byte(repoContent), 0o644); err != nil {
		t.Fatalf("escribir repo.go: %v", err)
	}

	doc := normalizer.APIDocument{
		Endpoints: []normalizer.Endpoint{
			{Method: "GET", Path: "/clientes/{id}"},
		},
	}
	out, err := EnrichDocument(doc, root)
	if err != nil {
		t.Fatalf("EnrichDocument error: %v", err)
	}
	impl := out.Endpoints[0].Implementation
	if impl == nil {
		t.Fatalf("se esperaba implementation metadata")
	}
	if !contains(impl.ServiceCalls, "service.BuscarPorId") {
		t.Fatalf("no se detecto llamada a servicio: %+v", impl.ServiceCalls)
	}
	if !contains(impl.RepositoryCalls, "repo.GetByID") {
		t.Fatalf("no se detecto llamada a repositorio: %+v", impl.RepositoryCalls)
	}
	if !impl.UsesDatabase {
		t.Fatalf("se esperaba detectar base de datos en cadena indirecta")
	}
	if !impl.UsesMessaging {
		t.Fatalf("se esperaba detectar mensajeria en helper invocado")
	}
	if len(impl.ExternalAPICalls) == 0 || !contains(impl.ExternalAPICalls, "client.Do") {
		t.Fatalf("se esperaba detectar API externa en cadena indirecta: %+v", impl.ExternalAPICalls)
	}
}

func TestEnrichDocument_DetectsDatabaseWithQueryRowAndConcatenatedRoute(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	routerFile := filepath.Join(root, "router.go")
	handlerFile := filepath.Join(root, "handler.go")
	repoFile := filepath.Join(root, "repo.go")

	routerContent := `package api
func setup(router *Router, h *ClienteHandler) {
	router.GET(options.BaseURL+"/clientes/:id", h.BuscarPorId)
}`
	handlerContent := `package api
func (h *ClienteHandler) BuscarPorId() {
	h.repo.BuscarPorId()
}`
	repoContent := `package api
func (r *Repo) BuscarPorId() {
	qryCliente := ` + "`SELECT c.id FROM cif.tb_cliente c WHERE c.id = @id`" + `
	r.db.QueryRow(qryCliente, sql.Named("id", 1))
}`

	if err := os.WriteFile(routerFile, []byte(routerContent), 0o644); err != nil {
		t.Fatalf("escribir router.go: %v", err)
	}
	if err := os.WriteFile(handlerFile, []byte(handlerContent), 0o644); err != nil {
		t.Fatalf("escribir handler.go: %v", err)
	}
	if err := os.WriteFile(repoFile, []byte(repoContent), 0o644); err != nil {
		t.Fatalf("escribir repo.go: %v", err)
	}

	doc := normalizer.APIDocument{
		Endpoints: []normalizer.Endpoint{
			{Method: "GET", Path: "/clientes/{id}"},
		},
	}
	out, err := EnrichDocument(doc, root)
	if err != nil {
		t.Fatalf("EnrichDocument error: %v", err)
	}

	impl := out.Endpoints[0].Implementation
	if impl == nil {
		t.Fatalf("se esperaba implementation metadata")
	}
	if !impl.UsesDatabase {
		t.Fatalf("se esperaba deteccion de base de datos via QueryRow")
	}
	if len(impl.DatabaseTypes) == 0 || !contains(impl.DatabaseTypes, "sql") {
		t.Fatalf("tipo de base de datos no detectado: %+v", impl.DatabaseTypes)
	}
	if len(impl.DatabaseTables) == 0 || !contains(impl.DatabaseTables, "cif.tb_cliente") {
		t.Fatalf("tabla no detectada correctamente: %+v", impl.DatabaseTables)
	}
}

func contains(values []string, wanted string) bool {
	for _, v := range values {
		if v == wanted {
			return true
		}
	}
	return false
}
