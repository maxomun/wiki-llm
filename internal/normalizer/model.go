package normalizer

// APIDocument representa el modelo interno estable para una API completa.
type APIDocument struct {
	Title       string
	Version     string
	Description string
	SourcePath  string
	Endpoints   []Endpoint
}

// SourceType representa el origen de datos de un endpoint.
type SourceType string

const (
	SourceOpenAPI SourceType = "openapi"
	SourcePostman SourceType = "postman"
)

// EndpointKey define la identidad unica de un endpoint.
type EndpointKey struct {
	Method string
	Path   string
}

// Endpoint representa una operacion HTTP de la API.
type Endpoint struct {
	BasePath     string
	Path         string
	Method       string
	OperationID  string
	Summary      string
	Description  string
	Tags         []string
	PathParams   []Parameter
	Parameters   []Parameter
	RequestBody  *RequestBody
	Responses    []Response
	Deprecated   bool
	SecurityRefs []string
	Sources      []SourceType
}

// Parameter representa un parametro de entrada de una operacion.
type Parameter struct {
	Name        string
	In          string
	Required    bool
	Description string
	SchemaRef   string
	Type        string
	Format      string
	Example     string
}

// RequestBody representa el cuerpo de entrada esperado por una operacion.
type RequestBody struct {
	Required     bool
	Description  string
	ContentTypes []string
	SchemaRef    string
}

// Response representa una salida posible de una operacion.
type Response struct {
	StatusCode   string
	Description  string
	ContentTypes []string
	SchemaRef    string
}
