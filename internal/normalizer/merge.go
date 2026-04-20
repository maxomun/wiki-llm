package normalizer

import (
	"regexp"
	"strings"
)

// MergeDocuments fusiona multiples documentos normalizados en uno solo.
// La prioridad se resuelve por orden de entrada: los primeros documentos
// establecen la base y los siguientes enriquecen campos faltantes.
func MergeDocuments(docs []APIDocument) APIDocument {
	if len(docs) == 0 {
		return APIDocument{}
	}

	normalizedDocs := make([]APIDocument, 0, len(docs))
	for _, doc := range docs {
		normalizedDocs = append(normalizedDocs, NormalizeDocument(doc))
	}

	merged := APIDocument{
		Title:          docs[0].Title,
		Version:        docs[0].Version,
		Description:    docs[0].Description,
		BasePath:       docs[0].BasePath,
		ContractSource: docs[0].ContractSource,
		SourcePath:     docs[0].SourcePath,
		Endpoints:      make([]Endpoint, 0),
	}

	index := make(map[EndpointKey]int)
	for _, doc := range normalizedDocs {
		merged.Title = firstNonEmpty(merged.Title, doc.Title)
		merged.Version = firstNonEmpty(merged.Version, doc.Version)
		merged.Description = firstNonEmpty(merged.Description, doc.Description)
		merged.SourcePath = appendSourcePath(merged.SourcePath, doc.SourcePath)

		for _, endpoint := range doc.Endpoints {
			endpoint = ensureEndpointSource(endpoint, doc.ContractSource)
			key := endpointMergeKey(endpoint)
			if pos, ok := index[key]; ok {
				merged.Endpoints[pos] = mergeEndpointBySourcePriority(merged.Endpoints[pos], endpoint)
				continue
			}
			index[key] = len(merged.Endpoints)
			merged.Endpoints = append(merged.Endpoints, endpoint)
		}
	}

	return ApplyEndpointConfidence(consolidateGlobalMetadata(normalizedDocs, merged))
}

func ensureEndpointSource(endpoint Endpoint, fallback SourceType) Endpoint {
	if len(endpoint.Sources) > 0 || strings.TrimSpace(string(fallback)) == "" {
		return endpoint
	}
	endpoint.Sources = []SourceType{fallback}
	return endpoint
}

var endpointPlaceholderRe = regexp.MustCompile(`\{[^/]+\}`)

func endpointMergeKey(endpoint Endpoint) EndpointKey {
	return EndpointKey{
		Method: strings.ToUpper(strings.TrimSpace(endpoint.Method)),
		Path:   endpointPathSignature(endpoint.Path),
	}
}

func endpointPathSignature(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return "/"
	}
	return endpointPlaceholderRe.ReplaceAllString(p, "{}")
}

func mergeEndpointBySourcePriority(existing, incoming Endpoint) Endpoint {
	existingHasOpenAPI := containsSource(existing.Sources, SourceOpenAPI)
	incomingHasOpenAPI := containsSource(incoming.Sources, SourceOpenAPI)

	base := existing
	enrichment := incoming

	// OpenAPI define la estructura base incluso si llega despues.
	if !existingHasOpenAPI && incomingHasOpenAPI {
		base = incoming
		enrichment = existing
	}
	return mergeEndpoint(base, enrichment)
}

func mergeEndpoint(base, incoming Endpoint) Endpoint {
	incomingHasPostman := containsSource(incoming.Sources, SourcePostman)

	// Contract fields: base already representa la fuente prioritaria.
	base.BasePath = firstNonEmpty(base.BasePath, incoming.BasePath)
	base.Path = firstNonEmpty(base.Path, incoming.Path)
	base.Method = firstNonEmpty(base.Method, incoming.Method)
	base.OperationID = firstNonEmpty(base.OperationID, incoming.OperationID)

	// Text fields: OpenAPI como base, Postman solo complementa.
	base.Summary = firstNonEmpty(base.Summary, incoming.Summary)
	base.Description = firstNonEmpty(base.Description, incoming.Description)
	base.Tags = dedupeStrings(append(base.Tags, incoming.Tags...))
	base.SecurityRefs = dedupeStrings(append(base.SecurityRefs, incoming.SecurityRefs...))
	base.Sources = dedupeSourceTypes(append(base.Sources, incoming.Sources...))
	base.Deprecated = base.Deprecated || incoming.Deprecated

	base.PathParams = mergeParameters(base.PathParams, incoming.PathParams)
	base.Parameters = mergeParameters(base.Parameters, incoming.Parameters)
	base.RequestBody = mergeRequestBody(base.RequestBody, incoming.RequestBody, incomingHasPostman)
	base.Responses = mergeResponses(base.Responses, incoming.Responses, incomingHasPostman)

	// Re-normalizar despues del merge evita duplicados de params path
	// cuando las fuentes usan placeholders distintos (ej: {id} vs {id_cliente}).
	return normalizeEndpoint(base)
}

func mergeRequestBody(base, incoming *RequestBody, incomingHasPostman bool) *RequestBody {
	if base == nil {
		return incoming
	}
	if incoming == nil {
		return base
	}

	base.Required = base.Required || incoming.Required
	base.Description = firstNonEmpty(base.Description, incoming.Description)
	base.SchemaRef = firstNonEmpty(base.SchemaRef, incoming.SchemaRef)
	base.ContentTypes = dedupeStrings(append(base.ContentTypes, incoming.ContentTypes...))

	// Ejemplos desde Postman enriquecen request body sin redefinir estructura.
	if incomingHasPostman {
		base.Example = firstNonEmpty(base.Example, incoming.Example)
	}
	return base
}

func mergeParameters(base, incoming []Parameter) []Parameter {
	out := make([]Parameter, len(base))
	copy(out, base)
	index := make(map[string]int, len(out))
	for i, p := range out {
		index[strings.ToLower(p.In)+"|"+p.Name] = i
	}
	for _, p := range incoming {
		key := strings.ToLower(p.In) + "|" + p.Name
		if pos, ok := index[key]; ok {
			out[pos].Required = out[pos].Required || p.Required
			out[pos].Description = firstNonEmpty(out[pos].Description, p.Description)
			out[pos].SchemaRef = firstNonEmpty(out[pos].SchemaRef, p.SchemaRef)
			out[pos].Type = firstNonEmpty(out[pos].Type, p.Type)
			out[pos].Format = firstNonEmpty(out[pos].Format, p.Format)
			out[pos].Example = firstNonEmpty(out[pos].Example, p.Example)
			continue
		}
		index[key] = len(out)
		out = append(out, p)
	}
	return out
}

func containsSource(values []SourceType, wanted SourceType) bool {
	for _, v := range values {
		if v == wanted {
			return true
		}
	}
	return false
}

func mergeResponses(base, incoming []Response, incomingHasPostman bool) []Response {
	out := make([]Response, len(base))
	copy(out, base)
	index := make(map[string]int, len(out))
	for i, r := range out {
		index[r.StatusCode] = i
	}
	for _, r := range incoming {
		if pos, ok := index[r.StatusCode]; ok {
			out[pos].Description = firstNonEmpty(out[pos].Description, r.Description)
			out[pos].SchemaRef = firstNonEmpty(out[pos].SchemaRef, r.SchemaRef)
			out[pos].ContentTypes = dedupeStrings(append(out[pos].ContentTypes, r.ContentTypes...))
			if incomingHasPostman {
				out[pos].Example = firstNonEmpty(out[pos].Example, r.Example)
			}
			continue
		}
		index[r.StatusCode] = len(out)
		out = append(out, r)
	}
	return out
}

func appendSourcePath(base, incoming string) string {
	incoming = strings.TrimSpace(incoming)
	if incoming == "" {
		return base
	}
	if strings.TrimSpace(base) == "" {
		return incoming
	}
	parts := strings.Split(base, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == incoming {
			return base
		}
	}
	return base + ", " + incoming
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		n := strings.TrimSpace(v)
		if n == "" {
			continue
		}
		if _, ok := set[n]; ok {
			continue
		}
		set[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func firstNonEmpty(base, incoming string) string {
	if strings.TrimSpace(base) != "" {
		return base
	}
	return strings.TrimSpace(incoming)
}

func consolidateGlobalMetadata(sources []APIDocument, merged APIDocument) APIDocument {
	openapiDocs := filterDocumentsBySource(sources, SourceOpenAPI)
	postmanDocs := filterDocumentsBySource(sources, SourcePostman)

	merged.Title = pickMetadataValue(
		collectValues(openapiDocs, func(d APIDocument) string { return d.Title }),
		collectValues(postmanDocs, func(d APIDocument) string { return d.Title }),
		[]string{merged.Title},
		[]string{"api"},
	)

	merged.Version = pickMetadataValue(
		collectValues(openapiDocs, func(d APIDocument) string { return d.Version }),
		collectValues(postmanDocs, func(d APIDocument) string { return d.Version }),
		[]string{merged.Version},
		[]string{"unknown"},
	)

	merged.BasePath = consolidateBasePath(openapiDocs, merged.Endpoints)
	merged.ContractSource = pickContractSource(openapiDocs, postmanDocs)

	return merged
}

func filterDocumentsBySource(docs []APIDocument, source SourceType) []APIDocument {
	out := make([]APIDocument, 0, len(docs))
	for _, doc := range docs {
		if doc.ContractSource == source {
			out = append(out, doc)
			continue
		}
		if source == SourceOpenAPI && containsAnyEndpointSource(doc.Endpoints, SourceOpenAPI) {
			out = append(out, doc)
		}
		if source == SourcePostman && containsAnyEndpointSource(doc.Endpoints, SourcePostman) {
			out = append(out, doc)
		}
	}
	return out
}

func containsAnyEndpointSource(endpoints []Endpoint, source SourceType) bool {
	for _, ep := range endpoints {
		if containsSource(ep.Sources, source) {
			return true
		}
	}
	return false
}

func collectValues(docs []APIDocument, getter func(APIDocument) string) []string {
	out := make([]string, 0, len(docs))
	for _, doc := range docs {
		out = append(out, getter(doc))
	}
	return out
}

func pickMetadataValue(groups ...[]string) string {
	for _, group := range groups {
		for _, v := range group {
			clean := strings.TrimSpace(v)
			if clean == "" {
				continue
			}
			if strings.EqualFold(clean, "postman") {
				continue
			}
			return clean
		}
	}
	return ""
}

func consolidateBasePath(openapiDocs []APIDocument, endpoints []Endpoint) string {
	for _, doc := range openapiDocs {
		if strings.TrimSpace(doc.BasePath) != "" {
			return strings.TrimSpace(doc.BasePath)
		}
	}

	counts := make(map[string]int)
	topPath := ""
	topCount := 0
	for _, endpoint := range endpoints {
		base := strings.TrimSpace(endpoint.BasePath)
		if base == "" {
			continue
		}
		counts[base]++
		if counts[base] > topCount {
			topCount = counts[base]
			topPath = base
		}
	}
	if topPath != "" {
		return topPath
	}
	return "/"
}

func pickContractSource(openapiDocs, postmanDocs []APIDocument) SourceType {
	if len(openapiDocs) > 0 {
		return SourceOpenAPI
	}
	if len(postmanDocs) > 0 {
		return SourcePostman
	}
	return SourceCode
}
