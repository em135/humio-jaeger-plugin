package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strconv"
	"time"
)

type dependencyReader struct {
	logger hclog.Logger
	client *http.Client
}

type Dependency struct {
	Service  string `json:"service"`
	SpanID   string  `json:"span_id"`
	ParentID string `json:"parent_id"`
}

func (d dependencyReader) GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]model.DependencyLink, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetDependencies")
	defer span.Finish()
	// TODO: noget galt med tiden...
	var oneHourAgo = strconv.FormatInt(time.Now().Unix() - 3600, 10)
	var stringBody = `{"queryString":"groupBy([service, span_id, parent_id])", "start": "` + oneHourAgo + `s", "end": "now"}`
	d.logger.Warn("INFOTAG GetDependencies()1.1 " + stringBody)
	var body = []byte(`{"queryString":"groupBy([service, span_id, parent_id])", "start": "` + oneHourAgo + `s", "end": "now"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

	var parentIdChildCounts = make(map[string]map[string] int)
	var spanIdServices = make(map[string]string)

	for i := range dependencies {
		var dependency = dependencies[i]
		var service = dependency.Service
		var span = dependency.SpanID
		var parent = dependency.ParentID
		spanIdServices[span] = service
		if parent == ""{
			continue
		}
		if serviceCounts, parentExists  := parentIdChildCounts[parent]; parentExists  {
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
	//
	//for k, v := range parentIdChildCounts {
	//	d.logger.Warn("INFOTAG GetDependencies() key " + k)
	//	for j, v := range v {
	//		d.logger.Warn("INFOTAG GetDependencies() value " + j + " count: " + string(v))
	//	}
	//}

	d.logger.Warn("INFOTAG GetDependencies()4")
	var dependencyLinks []model.DependencyLink
	for parent, childCounts := range parentIdChildCounts {
		if service, parentExists := spanIdServices[parent]; parentExists {
			for child, count := range childCounts {
				var dependencyLink = model.DependencyLink{
					Parent:               service,
					Child:                child,
					CallCount: 			  uint64(count),
					Source:               "Humio",
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				}
				dependencyLinks = append(dependencyLinks, dependencyLink)
			}
		}
	}

	//for i := range spanIdServices {
	//	var span = spanIdServices[i]
	//	if childCounts, parentExists := parentIdChildCounts[span]; parentExists {
	//		d.logger.Warn("INFOTAG GetDependencies()4.0")
	//		for child, count := range childCounts {
	//			d.logger.Warn("INFOTAG GetDependencies()4.1")
	//			var dependencyLink = model.DependencyLink{
	//				Parent:               span,
	//				Child:                child,
	//				CallCount: 			  uint64(count),
	//				Source:               "Humio",
	//				XXX_NoUnkeyedLiteral: struct{}{},
	//				XXX_unrecognized:     nil,
	//				XXX_sizecache:        0,
	//			}
	//			dependencyLinks = append(dependencyLinks, dependencyLink)
	//		}
	//
	//	}
	//}
	d.logger.Warn("INFOTAG GetDependencies()5")
	//var m1 = model.DependencyLink{
	//	Parent:               "m1",
	//	Child:                "m2",
	//	CallCount:            1,
	//	Source:               "idk",
	//	XXX_NoUnkeyedLiteral: struct{}{},
	//	XXX_unrecognized:     nil,
	//	XXX_sizecache:        0,
	//}
	//var m2 = model.DependencyLink{
	//	Parent:               "m1",
	//	Child:                "m3",
	//	CallCount:            1,
	//	Source:               "idk",
	//	XXX_NoUnkeyedLiteral: struct{}{},
	//	XXX_unrecognized:     nil,
	//	XXX_sizecache:        0,
	//}
	//
	//
	//var m3 = model.DependencyLink{
	//	Parent:               "",
	//	Child:                "",
	//	CallCount:            1,
	//	Source:               "idk",
	//	XXX_NoUnkeyedLiteral: struct{}{},
	//	XXX_unrecognized:     nil,
	//	XXX_sizecache:        0,
	//}
	//
	//
	//var models []model.DependencyLink
	//models = append(models, m1)
	//models = append(models, m2)
	//models = append(models, m3)
	// TODO dependencyLinks længde er 0 :O
	t := strconv.Itoa(len(dependencyLinks))
	d.logger.Warn("INFOTAG GetDependencies() 6 lenght " + t)
	return dependencyLinks, nil
}

func (h *HumioPlugin) DependencyReader() dependencystore.Reader {
	h.Logger.Warn("INFOTAG DependencyReader()")
	if h.dependencyReader == nil {
		h.Logger.Warn("INFOTAG DependencyReader() is nil")
		dependencyReader := &dependencyReader{logger: h.Logger, client: h.Client}
		return dependencyReader
	}
	return h.dependencyReader
}

