package normalizer

import "strings"

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
		Title:       docs[0].Title,
		Version:     docs[0].Version,
		Description: docs[0].Description,
		SourcePath:  docs[0].SourcePath,
		Endpoints:   make([]Endpoint, 0),
	}

	index := make(map[EndpointKey]int)
	for _, doc := range normalizedDocs {
		merged.Title = firstNonEmpty(merged.Title, doc.Title)
		merged.Version = firstNonEmpty(merged.Version, doc.Version)
		merged.Description = firstNonEmpty(merged.Description, doc.Description)
		merged.SourcePath = appendSourcePath(merged.SourcePath, doc.SourcePath)

		for _, endpoint := range doc.Endpoints {
			key := EndpointKey{Method: strings.ToUpper(endpoint.Method), Path: endpoint.Path}
			if pos, ok := index[key]; ok {
				merged.Endpoints[pos] = mergeEndpointBySourcePriority(merged.Endpoints[pos], endpoint)
				continue
			}
			index[key] = len(merged.Endpoints)
			merged.Endpoints = append(merged.Endpoints, endpoint)
		}
	}

	return merged
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
	base.BasePath = firstNonEmpty(base.BasePath, incoming.BasePath)
	base.Path = firstNonEmpty(base.Path, incoming.Path)
	base.Method = firstNonEmpty(base.Method, incoming.Method)
	base.OperationID = firstNonEmpty(base.OperationID, incoming.OperationID)
	base.Summary = firstNonEmpty(base.Summary, incoming.Summary)
	base.Description = firstNonEmpty(base.Description, incoming.Description)
	base.Tags = dedupeStrings(append(base.Tags, incoming.Tags...))
	base.SecurityRefs = dedupeStrings(append(base.SecurityRefs, incoming.SecurityRefs...))
	base.Sources = dedupeSourceTypes(append(base.Sources, incoming.Sources...))
	base.Deprecated = base.Deprecated || incoming.Deprecated

	base.PathParams = mergeParameters(base.PathParams, incoming.PathParams)
	base.Parameters = mergeParameters(base.Parameters, incoming.Parameters)
	base.Responses = mergeResponses(base.Responses, incoming.Responses)
	if base.RequestBody == nil {
		base.RequestBody = incoming.RequestBody
	} else if incoming.RequestBody != nil {
		base.RequestBody.Required = base.RequestBody.Required || incoming.RequestBody.Required
		base.RequestBody.Description = firstNonEmpty(base.RequestBody.Description, incoming.RequestBody.Description)
		base.RequestBody.SchemaRef = firstNonEmpty(base.RequestBody.SchemaRef, incoming.RequestBody.SchemaRef)
		base.RequestBody.ContentTypes = dedupeStrings(append(base.RequestBody.ContentTypes, incoming.RequestBody.ContentTypes...))
	}

	// Si Postman trae body real, enriquece descripcion/content-types aunque exista body en OpenAPI.
	if containsSource(incoming.Sources, SourcePostman) && incoming.RequestBody != nil {
		if base.RequestBody == nil {
			base.RequestBody = incoming.RequestBody
		} else {
			base.RequestBody.Description = firstNonEmpty(incoming.RequestBody.Description, base.RequestBody.Description)
			base.RequestBody.ContentTypes = dedupeStrings(append(base.RequestBody.ContentTypes, incoming.RequestBody.ContentTypes...))
		}
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

func mergeResponses(base, incoming []Response) []Response {
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
