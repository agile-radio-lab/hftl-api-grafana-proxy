package main

import "time"

// AnnotationsReq encodes the information provided by Grafana in its requests.
type AnnotationsReq struct {
	Range      Range      `json:"range"`
	Annotation Annotation `json:"annotation"`
}

// Range specifies the time range the request is valid for.
type Range struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// RangeRaw specifies the time range the request is valid for.
type RangeRaw struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Annotation is the object passed by Grafana when it fetches annotations.
//
// http://docs.grafana.org/plugins/developing/datasources/#annotation-query
type Annotation struct {
	// Name must match in the request and response
	Name string `json:"name"`

	Datasource string `json:"datasource"`
	IconColor  string `json:"iconColor"`
	Enable     bool   `json:"enable"`
	ShowLine   bool   `json:"showLine"`
	Query      string `json:"query"`
}

// AnnotationResponse contains all the information needed to render an
// annotation event.
//
// https://github.com/grafana/simple-json-datasource#annotation-api
type AnnotationResponse struct {
	// The original annotation sent from Grafana.
	Annotation Annotation `json:"annotation"`
	// Time since UNIX Epoch in milliseconds. (required)
	Time int64 `json:"time"`
	// The title for the annotation tooltip. (required)
	Title string `json:"title"`
	// Tags for the annotation. (optional)
	Tags string `json:"tags"`
	// Text for the annotation. (optional)
	Text string `json:"text"`
}

// QueryTarget query targets field
type QueryTarget struct {
	ReferenceID string `json:"refId"`
	Target      string `json:"target"`
	Hide        bool   `json:"hide"`
	Type        string `json:"type"`
}

// Metric is a metric type
type Metric struct {
	Text  interface{} `json:"text"`
	Value interface{} `json:"value"`
}

// TableColumn response table column
type TableColumn struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// QueryResponseTimeserie grafana timeserie query response
type QueryResponseTimeserie struct {
	Target     string          `json:"target"`
	DataPoints [][]interface{} `json:"datapoints"`
}

// QueryResponseTable grafana table query response
type QueryResponseTable struct {
	Type    string          `json:"type"`
	Columns []TableColumn   `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Target  string          `json:"target"`
}

// QueryRequest grafana query request
type QueryRequest struct {
	Timezone      string            `json:"timezone"`
	PanelID       int               `json:"panelId"`
	Range         Range             `json:"range"`
	RangeRaw      RangeRaw          `json:"rangeRaw"`
	Interval      string            `json:"interval"`
	Targets       []QueryTarget     `json:"targets"`
	Format        string            `json:"format"`
	MaxDataPoints int64             `json:"maxDataPoints"`
	IntervalMs    int               `json:"intervalMs"`
	Type          string            `json:"type"`
	ScopedVars    map[string]Metric `json:"scopedVars"`
}

// QueryResponse grafana query response
type QueryResponse struct {
	Data interface{} `json:"data"`
}

// SearchResponse grafana search response
type SearchResponse struct {
	Data []string `json:"data"`
}
