package normalizer

import "testing"

func TestApplyEndpointConfidence_AssignsExpectedLevels(t *testing.T) {
	t.Parallel()

	doc := APIDocument{
		Endpoints: []Endpoint{
			{
				Method:     "GET",
				Path:       "/high",
				Sources:    []SourceType{SourceOpenAPI, SourcePostman, SourceCode},
				Confidence: "",
			},
			{
				Method:  "GET",
				Path:    "/medium",
				Sources: []SourceType{SourceOpenAPI},
			},
			{
				Method:  "GET",
				Path:    "/low",
				Sources: []SourceType{SourcePostman},
			},
		},
	}

	out := ApplyEndpointConfidence(doc)
	if out.Endpoints[0].Confidence != ConfidenceHigh {
		t.Fatalf("confidence high esperada, actual=%q", out.Endpoints[0].Confidence)
	}
	if out.Endpoints[1].Confidence != ConfidenceMedium {
		t.Fatalf("confidence medium esperada, actual=%q", out.Endpoints[1].Confidence)
	}
	if out.Endpoints[2].Confidence != ConfidenceLow {
		t.Fatalf("confidence low esperada, actual=%q", out.Endpoints[2].Confidence)
	}
}

func TestCalculateEndpointConfidence_UsesImplementationAsCodeSignal(t *testing.T) {
	t.Parallel()

	endpoint := Endpoint{
		Method: "GET",
		Path:   "/impl",
		Sources: []SourceType{
			SourcePostman,
		},
		Implementation: &ImplementationInfo{HandlerName: "Get"},
	}

	got := calculateEndpointConfidence(endpoint)
	if got != ConfidenceMedium {
		t.Fatalf("confidence inesperada con implementation detectada: %q", got)
	}
}
