package trace

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client queries the Jaeger HTTP Query API.
type Client struct {
	BaseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Jaeger API response structs ---

type jaegerResponse struct {
	Data   []jaegerTrace `json:"data"`
	Errors []struct {
		Msg string `json:"msg"`
	} `json:"errors"`
}

type jaegerTrace struct {
	TraceID   string                    `json:"traceID"`
	Spans     []jaegerSpan              `json:"spans"`
	Processes map[string]jaegerProcess  `json:"processes"`
}

type jaegerSpan struct {
	TraceID       string         `json:"traceID"`
	SpanID        string         `json:"spanID"`
	OperationName string         `json:"operationName"`
	References    []jaegerRef    `json:"references"`
	StartTime     int64          `json:"startTime"` // microseconds since epoch
	Duration      int64          `json:"duration"`  // microseconds
	Tags          []jaegerKV     `json:"tags"`
	Logs          []jaegerLog    `json:"logs"`
	ProcessID     string         `json:"processID"`
}

type jaegerRef struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type jaegerKV struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type jaegerLog struct {
	Timestamp int64      `json:"timestamp"`
	Fields    []jaegerKV `json:"fields"`
}

type jaegerProcess struct {
	ServiceName string     `json:"serviceName"`
	Tags        []jaegerKV `json:"tags"`
}

// FetchTraces returns the most recent traces for a service, optionally filtered by operation.
func (c *Client) FetchTraces(service, operation string, limit int) ([]Trace, error) {
	params := url.Values{}
	params.Set("service", service)
	params.Set("limit", fmt.Sprintf("%d", limit))
	if operation != "" {
		params.Set("operation", operation)
	}

	endpoint := fmt.Sprintf("%s/api/traces?%s", c.BaseURL, params.Encode())
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Jaeger at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	var jResp jaegerResponse
	if err := json.NewDecoder(resp.Body).Decode(&jResp); err != nil {
		return nil, fmt.Errorf("failed to parse Jaeger response: %w", err)
	}
	if len(jResp.Errors) > 0 {
		return nil, fmt.Errorf("Jaeger error: %s", jResp.Errors[0].Msg)
	}

	traces := make([]Trace, 0, len(jResp.Data))
	for _, jt := range jResp.Data {
		traces = append(traces, convertTrace(jt))
	}
	return traces, nil
}

// FetchTrace returns a single trace by ID.
func (c *Client) FetchTrace(traceID string) (*Trace, error) {
	endpoint := fmt.Sprintf("%s/api/traces/%s", c.BaseURL, traceID)
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Jaeger: %w", err)
	}
	defer resp.Body.Close()

	var jResp jaegerResponse
	if err := json.NewDecoder(resp.Body).Decode(&jResp); err != nil {
		return nil, fmt.Errorf("failed to parse Jaeger response: %w", err)
	}
	if len(jResp.Data) == 0 {
		return nil, fmt.Errorf("trace %q not found", traceID)
	}

	t := convertTrace(jResp.Data[0])
	return &t, nil
}

func convertTrace(jt jaegerTrace) Trace {
	spans := make([]Span, 0, len(jt.Spans))
	for _, js := range jt.Spans {
		spans = append(spans, convertSpan(js, jt.Processes))
	}
	return Trace{
		TraceID: jt.TraceID,
		Spans:   spans,
	}
}

func convertSpan(js jaegerSpan, processes map[string]jaegerProcess) Span {
	s := Span{
		SpanID:        js.SpanID,
		OperationName: js.OperationName,
		StartTime:     time.UnixMicro(js.StartTime),
		Duration:      time.Duration(js.Duration) * time.Microsecond,
		Tags:          make(map[string]interface{}),
	}

	if proc, ok := processes[js.ProcessID]; ok {
		s.Service = proc.ServiceName
	}

	for _, ref := range js.References {
		if ref.RefType == "CHILD_OF" {
			s.ParentSpanID = ref.SpanID
			break
		}
	}

	for _, tag := range js.Tags {
		s.Tags[tag.Key] = tag.Value
		if tag.Key == "error" {
			if v, ok := tag.Value.(bool); ok && v {
				s.HasError = true
			}
		}
	}

	// extract error message from logs
	for _, log := range js.Logs {
		for _, field := range log.Fields {
			if field.Key == "message" || field.Key == "error" || field.Key == "error.object" {
				if msg, ok := field.Value.(string); ok && msg != "" {
					s.ErrorMessage = msg
				}
			}
		}
	}

	return s
}
