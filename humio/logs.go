package humio

type Log struct {
	Event     string `json:"@rawstring"`
	Level     string `json:"level"`
	Timestamp int64  `json:"@timestamp"`
	TraceID   string `json:"traceId,omitempty"`
	SpanID    string `json:"spanId,omitempty"`
	Logger    string `json:"logger"`
	Thread    string `json:"thread"`
	Throwable string `json:"throwable,omitempty"`
}
