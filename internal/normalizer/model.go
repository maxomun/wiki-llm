package normalizer

// APIDocument representa el modelo interno estable para una API completa.
type APIDocument struct {
	Title       string
	Version     string
	Description string
	SourcePath  string
	Endpoints   []Endpoint
}

// Endpoint representa una operacion HTTP de la API.
type Endpoint struct {
	Path         string
	Method       string
	OperationID  string
	Summary      string
	Description  string
	Tags         []string
	Parameters   []Parameter
	RequestBody  *RequestBody
	Responses    []Response
	Deprecated   bool
	SecurityRefs []string
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
