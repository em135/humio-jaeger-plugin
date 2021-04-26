package plugin

import (
	"context"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/opentracing/opentracing-go"
	"time"
)

type dependencyReader struct {
	plugin *HumioPlugin
}

func (d dependencyReader) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetDependencies")
	defer span.Finish()
	d.plugin.Logger.Warn("INFOTAG GetDependencies()")

	return nil, nil
}

func (h *HumioPlugin) DependencyReader() dependencystore.Reader {
	h.Logger.Warn("INFOTAG DependencyReader()")
	if h.dependencyReader == nil {
		h.Logger.Warn("INFOTAG DependencyReader() is nil")
		dependencyReader := &dependencyReader{plugin: h}
		return dependencyReader
	}
	return h.dependencyReader
}

