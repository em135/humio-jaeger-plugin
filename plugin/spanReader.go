package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"humio-jaeger-plugin/humio"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/opentracing/opentracing-go"
)

type spanReader struct {
	logger hclog.Logger
	client *http.Client
}

func (h *HumioPlugin) SpanReader() spanstore.Reader {
	if h.spanReader == nil {
		reader := &spanReader{logger: h.logger, client: h.client}
		h.spanReader = reader
		return reader
	}
	return h.spanReader
}

func (s *spanReader) GetTrace(ctx context.Context, traceID model.TraceID) (*model.Trace, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetTrace")
	defer span.Finish()

	var beginningOfTime = strconv.FormatInt(time.Time.Unix(time.Now()), 10)
	var body = []byte(`{"queryString":"#type = traces | trace_id = ` + traceID.String() + ` | select(@rawstring)", "start": "` + beginningOfTime + `s", "end": "now"}`)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var spanResponse []humio.SpanResponse
	json.NewDecoder(resp.Body).Decode(&spanResponse)

	logs, err := s.findLogs([]string{traceID.String()}, beginningOfTime, "now", ctx)
	if err != nil {
		return nil, err
	}

	var spans = make([]*model.Span, 0, len(spanResponse))
	for _, humioSpan := range spanResponse {
		log := []*humio.Log{}
		if traceLogs, ok := logs[humioSpan.Payload().TraceID]; ok {
			if spanLogs, ok := traceLogs[humioSpan.Payload().SpanID]; ok {
				log = spanLogs
			}
		}

		span, err := createSpan(humioSpan.Payload(), log)
		if err != nil {
			return nil, err
		}
		spans = append(spans, span)
	}
	var trace = model.Trace{
		Spans: spans,
	}
	return &trace, nil
}

// TODO beggningOfTime might not be a good idea, maybe make a system property that the image is run with?
func (s *spanReader) GetServices(ctx context.Context) ([]string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetServices")
	defer span.Finish()

	var beginningOfTime = strconv.FormatInt(time.Time.Unix(time.Now()), 10)
	var body = []byte(`{"queryString":"#type = traces | groupBy(service)", "start": "` + beginningOfTime + `s", "end": "now"}`)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var services []humio.Service
	errdecode := json.NewDecoder(resp.Body).Decode(&services)
	if errdecode != nil {
		return nil, errdecode
	}
	var serviceNames = make([]string, 0, len(services))
	for _s := range services {
		serviceNames = append(serviceNames, services[_s].Service)
	}
	return serviceNames, nil
}

// TODO beggningOfTime might not be a good idea, maybe make a system property that the image is run with?
func (s *spanReader) GetOperations(ctx context.Context, query spanstore.OperationQueryParameters) ([]spanstore.Operation, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetOperations")
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
	var body = []byte(`{"queryString":"#type = traces | ` + queryFields + `groupBy(field=[name, kind])", "start": "` + beginningOfTime + `s", "end": "now"}`)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var operationsDecoded []humio.Operation
	errdecode := json.NewDecoder(resp.Body).Decode(&operationsDecoded)
	if errdecode != nil {
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
	span1, _ := opentracing.StartSpanFromContext(ctx, "GetOperations")
	defer span1.Finish()

	var service = query.ServiceName
	var numOfTraces = strconv.Itoa(query.NumTraces)
	var currentTime = time.Now().Unix()
	var startTime = currentTime - query.StartTimeMin.Unix()
	var endTime = currentTime - query.StartTimeMax.Unix()
	if endTime < 0 {
		endTime = 0
	}
	var minDuration = query.DurationMin.Nanoseconds()
	var maxDuration = query.DurationMax.Nanoseconds()
	var startTimeString = strconv.FormatInt(startTime, 10)
	var endTimeString = strconv.FormatInt(endTime, 10)
	var minDurationString = strconv.FormatInt(minDuration, 10)
	var maxDurationString = strconv.FormatInt(maxDuration, 10)
	var operation = query.OperationName
	var tags = query.Tags

	var queryFields strings.Builder
	if service != "" {
		queryFields.WriteString("service=" + service + "|")
	}
	if operation != "" {
		queryFields.WriteString("name=" + operation + "|")
	}

	for key := range tags {
		var value = tags[key]
		if !strings.EqualFold(key, strings.TrimSpace("error")) {
			queryFields.WriteString("attributes." + key + "=" + value + "|")
		} else {
			if strings.EqualFold(value, strings.TrimSpace("true")) {
				queryFields.WriteString("status=STATUS_CODE_ERROR |")
			} else if strings.EqualFold(value, strings.TrimSpace("false")) {
				queryFields.WriteString("status=STATUS_CODE_UNSET |")
			}
		}
	}

	var body []byte
	if minDuration == 0 && maxDuration == 0 {
		var testString = `{"queryString":"#type = traces | trace_id =~ join({` + queryFields.String() + ` groupBy(trace_id, limit=` + numOfTraces + `)}) | select(@rawstring)", "start": "` + startTimeString + `s", "end": "` + endTimeString + `s"}`
		s.logger.Debug("Query: " + testString)
		body = []byte(testString)
	} else {
		// TODO: While max() of span durations is not guaranteed to be equal to trace duration, it is a very good approximation
		// The sum of span durations is much larger than trace duration due to overlap!
		// TODO Bug: This only considers the duration of spans matching the tags, not the duration of the entire trace itself!
		var testString = `{"queryString":"#type = traces | trace_id =~ join({` + queryFields.String() + ` duration:=end-start | groupBy(trace_id, function=max(duration, as=trace_duration)) | test(trace_duration >= ` + minDurationString + `) | test(trace_duration <= ` + maxDurationString + `) | tail(` + numOfTraces + `)}) | select(@rawstring)", "start": "` + startTimeString + `s", "end": "` + endTimeString + `s"}`
		s.logger.Debug("Query: " + testString)
		body = []byte(testString)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var spanResponse []humio.SpanResponse
	json.NewDecoder(resp.Body).Decode(&spanResponse)

	// Get all trace ids to query for logs
	// Go does not have sets, so this is a hacky alternative
	traceIdSet := map[string]struct{}{}
	for _, span := range spanResponse {
		traceIdSet[span.Payload().TraceID] = struct{}{}
	}
	traceIds := make([]string, 0, len(traceIdSet))
	for id := range traceIdSet {
		traceIds = append(traceIds, id)
	}
	logs, err := s.findLogs(traceIds, startTimeString, endTimeString, ctx)
	if err != nil {
		return nil, err
	}

	var traceIdSpans = make(map[string][]*model.Span)

	for i := range spanResponse {
		var humioSpan = spanResponse[i].Payload()
		log := []*humio.Log{}
		if traceLogs, ok := logs[humioSpan.TraceID]; ok {
			if spanLogs, ok := traceLogs[humioSpan.SpanID]; ok {
				log = spanLogs
			}
		}

		span, err := createSpan(humioSpan, log)
		if err != nil {
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
			Spans: value,
		}
		traces = append(traces, &trace)
	}
	return traces, nil
}

// This method is not used
func (s *spanReader) FindTraceIDs(ctx context.Context, query *spanstore.TraceQueryParameters) ([]model.TraceID, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "FindTraceIDs")
	defer span.Finish()

	return nil, nil
}

func (s *spanReader) findLogs(traceIds []string, start string, end string, ctx context.Context) (map[string]map[string][]*humio.Log, error) {
	if len(traceIds) == 0 {
		return map[string]map[string][]*humio.Log{}, nil
	}

	var queryString = `{"queryString":"#type = elastic_input | in(traceId, values=[` + strings.Join(traceIds, ",") + `])", "start": "` + start + `s", "end": "` + end + `s"}`
	s.logger.Debug("Query: " + queryString)
	body := []byte(queryString)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var rawLogs []*humio.Log
	err = json.NewDecoder(resp.Body).Decode(&rawLogs)
	if err != nil {
		return nil, err
	}

	// Return logs ordered by trace id and span id (double nested map of slices)
	logs := map[string]map[string][]*humio.Log{}
	for _, log := range rawLogs {
		if traceLogs, ok := logs[log.TraceID]; ok {
			traceLogs[log.SpanID] = append(traceLogs[log.SpanID], log)
		} else {
			logs[log.TraceID] = map[string][]*humio.Log{
				log.SpanID: {log},
			}
		}
	}

	return logs, nil
}

func createSpan(humioSpan *humio.Span, humioLogs []*humio.Log) (*model.Span, error) {
	traceId, err := model.TraceIDFromString(humioSpan.TraceID)
	if err != nil {
		return nil, err
	}
	spanId, err := model.SpanIDFromString(humioSpan.SpanID)
	if err != nil {
		return nil, err
	}
	parentId, _ := model.SpanIDFromString(humioSpan.ParentID)
	spanTags, processTags := createSpanTags(humioSpan)

	var references []model.SpanRef
	var reference = model.SpanRef{
		TraceID: traceId,
		SpanID:  parentId,
		RefType: 0,
	}
	references = append(references, reference)

	process := model.Process{
		ServiceName: humioSpan.Service,
		Tags:        processTags,
	}

	// TODO: Do errors look right?
	logs := make([]model.Log, 0, len(humioLogs))
	for _, log := range humioLogs {
		logs = append(logs, model.Log{
			Timestamp: time.Unix(0, log.Timestamp*int64(time.Millisecond)),
			Fields: []model.KeyValue{
				model.String("event", log.Event),
			},
		})
	}

	var span = &model.Span{
		TraceID:       traceId,
		SpanID:        spanId,
		OperationName: humioSpan.Name,
		References:    references,
		Flags:         0,
		StartTime:     time.Unix(0, humioSpan.Start),
		Duration:      time.Duration(humioSpan.End - humioSpan.Start),
		Tags:          spanTags,
		Logs:          logs,
		Process:       &process,
	}
	return span, nil
}

func createSpanTags(modelSpan *humio.Span) ([]model.KeyValue, []model.KeyValue) {
	var spanTags []model.KeyValue
	var processTags []model.KeyValue

	for key, value := range modelSpan.Attributes {
		if strings.HasPrefix(key, "process.") {
			key = strings.Replace(key, "process.", "", 1)
			processTags = append(processTags, model.KeyValue{Key: key, VStr: string(value)})
		} else {
			spanTags = append(spanTags, model.KeyValue{Key: key, VStr: string(value)})
		}
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

	return spanTags, processTags
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
