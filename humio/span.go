package humio

import (
	"encoding/json"
)

type SpanResponse struct {
	RawString string `json:"@rawstring"`

	humioSpan *Span
}

func (s *SpanResponse) Payload() *Span {
	if s.humioSpan == nil {
		var span Span
		json.Unmarshal([]byte(s.RawString), &span)
		s.humioSpan = &span
	}
	return s.humioSpan
}

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
