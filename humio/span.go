package humio

import "encoding/json"

type Span struct {
	TraceID    string                     `json:"trace_id"`
	SpanID     string                     `json:"span_id"`
	ParentID   string                     `json:"parent_id"`
	Kind       string                     `json:"kind"`
	Service    string                     `json:"service"`
	Name       string                     `json:"name"`
	Start      int64                      `json:"start"`
	End        int64                      `json:"end"`
	Attributes map[string]json.RawMessage `json:"attributes"`
	Status     string                     `json:"status"`
}
