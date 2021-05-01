package plugin

import (
	"context"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
)

type spanWriter struct {
	plugin *HumioPlugin
}

func (s spanWriter) WriteSpan(ctx context.Context, span *model.Span) error {
	s.plugin.Logger.Warn("INFOTAG WriteSpan() trace id " + span.TraceID.String())
	s.plugin.Logger.Warn("INFOTAG WriteSpan() span id  " + span.SpanID.String())
	return nil
}

func (h *HumioPlugin) SpanWriter() spanstore.Writer {
	h.Logger.Warn("INFOTAG SpanWriter()")
	if h.spanWriter == nil {
		h.Logger.Warn("INFOTAG SpanWriter() is nil")
		writer := &spanWriter{plugin: h}
		h.spanWriter = writer
		return writer
	}
	return h.spanWriter
}
