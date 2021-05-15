package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/opentracing/opentracing-go"
)

type dependencyReader struct {
	logger hclog.Logger
	client *http.Client
	url    string
}

type Dependency struct {
	Service  string `json:"service"`
	SpanID   string `json:"span_id"`
	ParentID string `json:"parent_id"`
}

func (d dependencyReader) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetDependencies")
	defer span.Finish()

	var oneDayAgo = "86400s"
	var stringBody = `{"queryString":"groupBy([service, span_id, parent_id])", "start": "` + oneDayAgo + `", "end": "now"}`
	d.logger.Warn("GetDependencies(): " + stringBody)
	var body = []byte(`{"queryString":"groupBy([service, span_id, parent_id])", "start": "` + oneDayAgo + `", "end": "now"}`)

	req, err := http.NewRequest("POST", d.url, bytes.NewBuffer(body))
	if err != nil {

		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var dependencies []Dependency
	json.NewDecoder(resp.Body).Decode(&dependencies)

	var parentIdChildCounts = make(map[string]map[string]int)
	var spanIdServices = make(map[string]string)

	for i := range dependencies {
		var dependency = dependencies[i]
		var service = dependency.Service
		var span = dependency.SpanID
		var parent = dependency.ParentID
		spanIdServices[span] = service
		if parent == "" {
			continue
		}
		if serviceCounts, parentExists := parentIdChildCounts[parent]; parentExists {
			if count, serviceExists := serviceCounts[service]; serviceExists {
				serviceCounts[service] = count + 1
			} else {
				serviceCounts[service] = 1
			}
			parentIdChildCounts[parent] = serviceCounts
		} else {
			serviceCounts := map[string]int{service: 1}
			parentIdChildCounts[parent] = serviceCounts
		}
	}

	var dependencyLinks []model.DependencyLink
	for parent, childCounts := range parentIdChildCounts {
		if service, parentExists := spanIdServices[parent]; parentExists {
			for child, count := range childCounts {
				var dependencyLink = model.DependencyLink{
					Parent:               service,
					Child:                child,
					CallCount:            uint64(count),
					Source:               "Humio",
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				}
				dependencyLinks = append(dependencyLinks, dependencyLink)
			}
		}
	}

	return dependencyLinks, nil
}

func (h *HumioPlugin) DependencyReader() dependencystore.Reader {
	if h.dependencyReader == nil {
		dependencyReader := &dependencyReader{logger: h.logger, client: h.client, url: h.url}
		return dependencyReader
	}
	return h.dependencyReader
}
