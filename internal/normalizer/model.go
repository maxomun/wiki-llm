package normalizer

// APIDocument representa el modelo interno estable para una API completa.
type APIDocument struct {
	Title          string
	Version        string
	Description    string
	BasePath       string
	ContractSource SourceType
	SourcePath     string
	Endpoints      []Endpoint
}

// SourceType representa el origen de datos de un endpoint.
type SourceType string

const (
	SourceOpenAPI SourceType = "openapi"
	SourcePostman SourceType = "postman"
	SourceCode    SourceType = "code"
)

// ConfidenceLevel representa el nivel de certeza de un endpoint consolidado.
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

// EndpointKey define la identidad unica de un endpoint.
type EndpointKey struct {
	Method string
	Path   string
}

// Endpoint representa una operacion HTTP de la API.
type Endpoint struct {
	BasePath       string
	Path           string
	Method         string
	OperationID    string
	Summary        string
	Description    string
	Tags           []string
	PathParams     []Parameter
	Parameters     []Parameter
	RequestBody    *RequestBody
	Responses      []Response
	Deprecated     bool
	SecurityRefs   []string
	Sources        []SourceType
	Confidence     ConfidenceLevel
	Implementation *ImplementationInfo
}

// ImplementationInfo representa metadata de implementacion detectada en codigo fuente.
type ImplementationInfo struct {
	HandlerName         string
	HandlerFile         string
	ServiceCalls        []string
	RepositoryCalls     []string
	ExternalAPICalls    []string
	UsesDatabase        bool
	DatabaseTypes       []string
	DatabaseTables      []string
	DatabaseSPs         []string
	DatabaseCollections []string
	DatabaseQueries     []string
	UsesMessaging       bool
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
	Example      string
}

// Response representa una salida posible de una operacion.
type Response struct {
	StatusCode   string
	Description  string
	ContentTypes []string
	SchemaRef    string
	Example      string
}
