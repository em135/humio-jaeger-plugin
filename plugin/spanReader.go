package plugin

import (
	"HumioJaegerStoragePlugin/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strconv"
	"time"
)

const (
	url = "https://cloud.humio.com/api/v1/repositories/sockshop-traces/query"
)

type spanReader struct {
	logger hclog.Logger
	client *http.Client
}

func (h *HumioPlugin) SpanReader() spanstore.Reader {
	h.Logger.Warn("INFOTAG SpanReader()")
	if h.spanReader == nil {
		h.Logger.Warn("INFOTAG SpanReader() is nil")
		reader := &spanReader{logger: h.Logger, client: h.Client}
		h.spanReader = reader
		return reader
	}
	return h.spanReader
}

func (s *spanReader) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	s.logger.Warn("INFOTAG GetTrace()")
	str := fmt.Sprintf("%v", ctx)
	s.logger.Warn("INFOTAG " + str)

	var beginningOfTime = strconv.FormatInt(time.Time.Unix(time.Now()), 10)
	var body = []byte(`{"queryString":"* | trace_id = ` + traceID.String() + ` | groupBy(field=[@rawstring])", "start": "` + beginningOfTime + `s", "end": "now"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {

		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var spanElements []models.SpanElement
	json.NewDecoder(resp.Body).Decode(&spanElements)

	var spans []*model.Span
	for i := range spanElements {
		var spanElement = spanElements[i]
		span, err := createSpan(spanElement)
		if err != nil {
			return nil, err
		}
		spans = append(spans, span)
	}
	var trace = model.Trace{
		Spans:                spans,
		ProcessMap:           []model.Trace_ProcessMapping{},
		Warnings:             []string{},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}

	return &trace, nil
}

func (s *spanReader) GetServices(ctx context.Context) ([]string, error) {
	s.logger.Warn("INFOTAG GetServices() 0")
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetServices")
	defer span.Finish()
	var beginningOfTime = strconv.FormatInt(time.Time.Unix(time.Now()), 10)
	var body = []byte(`{"queryString":"groupBy(service)", "start": "` + beginningOfTime + `s", "end": "now"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		s.logger.Error("INFOTAG GetServices() error " + err.Error())
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("INFOTAG GetServices() error " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var services []models.Service
	errdecode := json.NewDecoder(resp.Body).Decode(&services)
	if errdecode != nil {
		s.logger.Warn("INFOTAG GetServices() decode error " + errdecode.Error())
		return nil, errdecode
	}
	var serviceNames []string
	for _s := range services {
		serviceNames = append(serviceNames, services[_s].Service)
	}
	return serviceNames, nil
}

// TODO beggningOfTime might not be a good idea, make a system property that the image is run with
func (s *spanReader) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	s.logger.Warn("INFOTAG GetOperations()")
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetOperations")
	defer span.Finish()

	var queryFields string
	var service = query.ServiceName
	var kind = query.SpanKind
	if service != "" {
		queryFields += "service=" + service + "|"
	}
	if kind != "" {
		queryFields += "kind=" + kind + "|"
	}
	var beginningOfTime = strconv.FormatInt(time.Time.Unix(time.Now()), 10)
	var body = []byte(`{"queryString":"` + queryFields + `groupBy(field=[name, kind])", "start": "` + beginningOfTime + `s", "end": "now"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		s.logger.Error("INFOTAG GetOperations() error: " + err.Error())
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("INFOTAG GetOperations() error " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var operationsDecoded []models.Operation
	errdecode := json.NewDecoder(resp.Body).Decode(&operationsDecoded)
	if errdecode != nil {
		s.logger.Error("INFOTAG GetOperations() decode error " + errdecode.Error())
		return nil, errdecode
	}

	var operations []spanstore.Operation
	for i := range operationsDecoded {
		var operation = operationsDecoded[i]
		var spanKind = getJaegerSpanKind(operation.Kind)
		operations = append(operations, spanstore.Operation{Name: operation.Name, SpanKind: spanKind})
	}
	return operations, nil
}

func (s *spanReader) FindTraces(ctx context.Context, query *spanstore.TraceQueryParameters) ([]*model.Trace, error) {
	s.logger.Warn("INFOTAG FindTraces()")
	span1, ctx := opentracing.StartSpanFromContext(ctx, "GetOperations")
	defer span1.Finish()

	var service = query.ServiceName
	var numOfTraces = strconv.Itoa(query.NumTraces)
	var currentTime = time.Now().Unix()
	var startTime = currentTime-query.StartTimeMin.Unix()
	var endTime = currentTime-query.StartTimeMax.Unix()
	if endTime < 0 {
		endTime = 0
	}
	var minDuration = query.DurationMin.Nanoseconds()
	var maxDuration = query.DurationMax.Nanoseconds()
	var startTimeString = strconv.FormatInt(startTime, 10)
	var endTimeString = strconv.FormatInt(endTime, 10)
	var minDurationString = strconv.FormatInt(minDuration, 10)
	var maxDurationString = strconv.FormatInt(maxDuration, 10)



	//var operation = query.OperationName
	//var operation = query.OperationName
	//query.Tags



	var body []byte
	if minDuration == 0 && maxDuration == 0 {
		var testString = `{"queryString":"* | trace_id =~ join({service=` + service + ` | groupBy(trace_id, limit=` + numOfTraces + `)}) | groupBy(field=[@rawstring])", "start": "` + startTimeString + `s", "end": "` + endTimeString + `s"}`
		s.logger.Warn("INFOTAG query " + testString)
		body = []byte(testString)
	} else {
		// TODO bug, spans are limited, not trace ids
		var testString = `{"queryString":"* | trace_id =~ join({service=` + service + ` | duration:=end-start | groupBy(trace_id, function=sum(duration, as=trace_duration)) | test(trace_duration >= ` + minDurationString + `) | test(trace_duration <= ` + maxDurationString + `)}) | groupBy(field=[@rawstring], limit=` + numOfTraces + `)", "start": "` + startTimeString + `s", "end": "` + endTimeString + `s"}`
		s.logger.Warn("INFOTAG query " + testString)
		body = []byte(testString)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		s.logger.Error("INFOTAG FindTraces() error " + err.Error())
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("INFOTAG FindTraces() error " + err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	var spanElements []models.SpanElement
	json.NewDecoder(resp.Body).Decode(&spanElements)

	var traceIdSpans = make(map[string][]*model.Span)

	for i := range spanElements {
		var spanElement = spanElements[i]
		span, err := createSpan(spanElement)
		if err != nil {
			s.logger.Error("INFOTAG FindTraces() error " + err.Error())
			return nil, err
		}
		var traceId = span.TraceID.String()
		if traceIdSpans[traceId] != nil {
			traceIdSpans[traceId] = append(traceIdSpans[traceId], span)
		} else {
			traceIdSpans[traceId] = []*model.Span{span}
		}
	}

	var traces []*model.Trace
	for _, value := range traceIdSpans {
		var traceDuration int64
		for i := range value {
			var span = value[i]
			traceDuration += span.Duration.Nanoseconds()
		}
		var trace = model.Trace{
			Spans:                value,
			ProcessMap:           []model.Trace_ProcessMapping{},
			Warnings:             []string{},
			XXX_NoUnkeyedLiteral: struct{}{},
			XXX_unrecognized:     []byte{},
			XXX_sizecache:        0,
		}
		traces = append(traces, &trace)
	}
	return traces, nil
}

// TODO implement me
func (s *spanReader) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	s.logger.Warn("INFOTAG FindTraceIDs()")
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetOperations")
	defer span.Finish()
	return nil, nil
}

func createSpan(spanElement models.SpanElement) (*model.Span, error) {
	var modelSpan models.Span
	json.Unmarshal([]byte(spanElement.Rawstring), &modelSpan)
	traceId, err := model.TraceIDFromString(modelSpan.TraceID)
	if err != nil {
		return nil, err
	}
	spanId, err := model.SpanIDFromString(modelSpan.SpanID)
	if err != nil {
		return nil, err
	}
	parentId, _ := model.SpanIDFromString(modelSpan.ParentID)
	spanTags := createSpanTags(modelSpan)

	var references []model.SpanRef
	var reference = model.SpanRef{
		TraceID:              traceId,
		SpanID:               parentId,
		RefType:              0,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}
	references = append(references, reference)

	process := model.Process{
		ServiceName:          modelSpan.Service,
		Tags:                 []model.KeyValue{},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}
	var span = &model.Span{
		TraceID:              traceId,
		SpanID:               spanId,
		OperationName:        modelSpan.Name,
		References:           references,
		Flags:                0,
		StartTime:            time.Unix(0, modelSpan.Start),
		Duration:             time.Duration(modelSpan.End - modelSpan.Start),
		Tags:                 spanTags,
		Logs:                 []model.Log{},
		Process:              &process,
		ProcessID:            "",
		Warnings:             []string{},
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     []byte{},
		XXX_sizecache:        0,
	}
	return span, nil
}

func createSpanTags(modelSpan models.Span) []model.KeyValue {
	var spanTags []model.KeyValue
	for key, value := range modelSpan.Attributes {
		spanTags = append(spanTags, model.KeyValue{Key: key, VStr: string(value)})
	}
	var spanKind = getJaegerSpanKind(modelSpan.Kind)
	if spanKind != "" {
		spanTags = append(spanTags, model.KeyValue{Key: "span.kind", VStr: spanKind})
	}
	var status = getStatus(modelSpan.Status)
	if status != "" {
		spanTags = append(spanTags, model.KeyValue{Key: "span.status", VStr: status})
		if status == "ERROR" {
			spanTags = append(spanTags, model.KeyValue{Key: "error", VStr: "true"})
		}
	}
	return spanTags
}

func getJaegerSpanKind(input string) string {
	switch kind := input; kind {
	case "SPAN_KIND_CLIENT":
		return "client"
	case "SPAN_KIND_SERVER":
		return "server"
	case "SPAN_KIND_CONSUMER":
		return "consumer"
	case "SPAN_KIND_PRODUCER":
		return "producer"
	}
	return ""
}

func getStatus(input string) string {
	switch status := input; status {
	case "STATUS_CODE_OK":
		return "OK"
	case "STATUS_CODE_ERROR":
		return "ERROR"
	}
	return ""
}
