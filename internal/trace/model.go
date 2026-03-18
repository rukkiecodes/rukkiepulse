package trace

import "time"

type Trace struct {
	TraceID string
	Spans   []Span
}

type Span struct {
	SpanID        string
	ParentSpanID  string
	OperationName string
	Service       string
	StartTime     time.Time
	Duration      time.Duration
	HasError      bool
	ErrorMessage  string
	Tags          map[string]interface{}
}

// RootSpan returns the span with no parent (or the earliest span).
func (t *Trace) RootSpan() *Span {
	spanIDs := make(map[string]bool, len(t.Spans))
	for _, s := range t.Spans {
		spanIDs[s.SpanID] = true
	}
	for i := range t.Spans {
		if t.Spans[i].ParentSpanID == "" || !spanIDs[t.Spans[i].ParentSpanID] {
			return &t.Spans[i]
		}
	}
	if len(t.Spans) > 0 {
		return &t.Spans[0]
	}
	return nil
}

// TotalDuration returns the duration from first span start to last span end.
func (t *Trace) TotalDuration() time.Duration {
	if len(t.Spans) == 0 {
		return 0
	}
	root := t.RootSpan()
	if root != nil {
		return root.Duration
	}
	return t.Spans[0].Duration
}

// HasErrors returns true if any span has an error.
func (t *Trace) HasErrors() bool {
	for _, s := range t.Spans {
		if s.HasError {
			return true
		}
	}
	return false
}
