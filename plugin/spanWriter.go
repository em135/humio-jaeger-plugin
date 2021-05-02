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
	s.plugin.logger.Warn("Span write request ignored: This storage plugin only supports reading spans")
	return nil
}

func (h *HumioPlugin) SpanWriter() spanstore.Writer {
	if h.spanWriter == nil {
		writer := &spanWriter{plugin: h}
		h.spanWriter = writer
		return writer
	}
	return h.spanWriter
}
