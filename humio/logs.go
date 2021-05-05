package humio

type Log struct {
	Event     string `json:"@rawstring"`
	Level     string `json:"level"`
	Timestamp int64  `json:"@timestamp"`
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
	Logger    string `json:"logger"`
	Thread    string `json:"thread"`
	Throwable string `json:"throwable,omitempty"`
}
