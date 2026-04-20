package normalizer

func ApplyEndpointConfidence(doc APIDocument) APIDocument {
	out := doc
	out.Endpoints = make([]Endpoint, 0, len(doc.Endpoints))
	for _, endpoint := range doc.Endpoints {
		endpoint.Confidence = calculateEndpointConfidence(endpoint)
		out.Endpoints = append(out.Endpoints, endpoint)
	}
	return out
}

func calculateEndpointConfidence(endpoint Endpoint) ConfidenceLevel {
	score := 0
	if containsSource(endpoint.Sources, SourceOpenAPI) {
		score += 2
	}
	if containsSource(endpoint.Sources, SourcePostman) {
		score++
	}
	if containsSource(endpoint.Sources, SourceCode) || endpoint.Implementation != nil {
		score++
	}

	switch {
	case score >= 3:
		return ConfidenceHigh
	case score >= 2:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}
