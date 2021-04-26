package main
//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"net/http"
//)
//
//func main() {
//	var body = []byte(`{"queryString":"* | trace_id =~ join({service=carts | groupBy(trace_id, limit=2)}) | groupBy(field=[@rawstring])", "start": "72hours", "end": "now"}`)
//	req, err := http.NewRequest("POST", "https://cloud.humio.com/api/v1/repositories/sockshop-traces/query", bytes.NewBuffer(body))
//	if err != nil {
//		return
//	}
//	req.Header.Set("Authorization", "Bearer luK3y9DcxldlyJrpseqA8T5iF6APdXI839K9uwt0bKXd")
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Set("Accept", "application/json")
//
//	resp, err := http.DefaultClient.Do(req)
//	if err != nil {
//		return
//	}
//
//	defer resp.Body.Close()
//	//if resp.StatusCode == http.StatusOK {
//	//	bodyBytes, err := ioutil.ReadAll(resp.Body)
//	//	if err != nil {
//	//
//	//	}
//	//	bodyString := string(bodyBytes)
//	//	fmt.Println(bodyString)
//	//}
//	//var data map[string]interface{}
//	//reponseBody, err := ioutil.ReadAll(resp.Body)
//	//err = json.Unmarshal(reponseBody, &data)
//	//if err != nil {
//	//	panic(err)
//	//}
//	//fmt.Println(data["trace_id"])
//	//fmt.Println(data["span_id"])
//	var spanElements []SpanElement
//	json.NewDecoder(resp.Body).Decode(&spanElements)
//
//	println(len(spanElements))
//	for i := range spanElements {
//		var span = spanElements[i]
//		var modelSpan Span
//		json.Unmarshal([]byte(span.Rawstring), &modelSpan)
//		fmt.Println(fmt.Sprintf("%s", span.Rawstring))
//		//var v = modelSpan.Attributes["otel.library.name"]
//		//tags := modelSpan.Attributes.(map[string]string)
//		for key, value := range modelSpan.Attributes {
//			println(key)
//			println(value)
//		}
//		//fmt.Printf("span: %s", md["otel.library.name"])
//
//		//start, err := strconv.ParseInt(modelSpan.Start, 10, 64)
//		//if err != nil {
//		//}
//		//fmt.Printf("%d", uint64(start))
//		//end, err := strconv.ParseInt(modelSpan.End, 10, 64)
//		//if err != nil {
//		//}
//		//fmt.Printf("%d", uint64(end))
//	}
//
//
//	//var traces []*model.Trace
//	////var traceIdSpans = make(map[model.TraceID][]*model.Span)
//	//var traceIdSpans = make(map[string][]*model.Span)
//	//
//	//for i := range spanElements {
//	//	var spanElement = spanElements[i]
//	//	println(spanElement.Start)
//	//	println(spanElement.End)
//	//	//we := time.Unix(0, start)
//	//	//du := time.Duration(end-start)
//	//	//
//	//	//startstr := fmt.Sprintf("Elapsed time: %s\n", we)
//	//	//durstr := fmt.Sprintf("Duarati time: %s\n", du)
//	//	//s.plugin.Logger.Warn("INFOTAG start" + startstr)
//	//	//s.plugin.Logger.Warn("INFOTAG duration " + durstr)
//	//	var s = spanElement.Start
//	//	start, err := strconv.ParseInt(s, 10, 64)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//	var e = spanElement.End
//	//	end, err := strconv.ParseInt(e, 10, 64)
//	//	if err != nil {
//	//		panic(err)
//	//	}
//	//	timew := time.Unix(0, start)
//	//	du := time.Duration(end-start)
//	//	fmt.Printf("Elapsed time: %s\n", timew)
//	//	fmt.Printf("Duarati time: %s\n", du)
//	//
//	//	traceId, err := model.TraceIDFromString(spanElement.TraceID)
//	//	if err != nil {
//	//	}
//	//	spanId, err := model.SpanIDFromString(spanElement.SpanID)
//	//	if err != nil {
//	//	}
//	//	parentId, err := model.SpanIDFromString(spanElement.ParentID)
//	//	if err != nil {
//	//	}
//	//	var span = &model.Span{
//	//		TraceID:       traceId,
//	//		SpanID:        spanId,
//	//		OperationName: "",
//	//		References: []model.SpanRef{{
//	//			TraceID:              traceId,
//	//			SpanID:               parentId,
//	//			RefType:              0,
//	//			XXX_NoUnkeyedLiteral: struct{}{},
//	//			XXX_unrecognized:     nil,
//	//			XXX_sizecache:        0,
//	//		}},
//	//		Flags:                0,
//	//		StartTime:            time.Time{},
//	//		Duration:             0,
//	//		Tags:                 nil,
//	//		Logs:                 nil,
//	//		Process:              nil,
//	//		ProcessID:            "",
//	//		Warnings:             nil,
//	//		XXX_NoUnkeyedLiteral: struct{}{},
//	//		XXX_unrecognized:     nil,
//	//		XXX_sizecache:        0,
//	//	}
//	//	if traceIdSpans[traceId.String()] != nil {
//	//		traceIdSpans[traceId.String()] = append(traceIdSpans[traceId.String()], span)
//	//	} else {
//	//		traceIdSpans[traceId.String()] = []*model.Span{span}
//	//	}
//	//}
//	//
//	//for _, value := range traceIdSpans {
//	//	var trace = model.Trace{
//	//		Spans:                value,
//	//		ProcessMap:           nil,
//	//		Warnings:             nil,
//	//		XXX_NoUnkeyedLiteral: struct{}{},
//	//		XXX_unrecognized:     nil,
//	//		XXX_sizecache:        0,
//	//	}
//	//	traces = append(traces, &trace)
//	//}
//}
//
//type SpanElement struct {
//	Rawstring string `json:"@rawstring"`
//	Count     string `json:"_count"`
//	TraceID   string `json:"trace_id"`
//}
//
//type Span struct {
//	TraceID    string     `json:"trace_id"`
//	SpanID     string     `json:"span_id"`
//	ParentID   string     `json:"parent_id"`
//	Kind       string     `json:"kind"`
//	Service    string     `json:"service"`
//	Name       string     `json:"name"`
//	Start      int64    `json:"start"`
//	End        int64    `json:"end"`
//	Attributes map[string]string `json:"attributes"`
//	Status     string     `json:"status"`
//}
