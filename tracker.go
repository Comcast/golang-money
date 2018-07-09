package money

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"
)

type contextKey int

const (
	//contextKeyTracker is the key for child spans management component
	contextKeyTracker contextKey = iota
)

//Header keys
const (
	MoneyHeader      = "X-MoneyTrace"
	MoneySpansHeader = "X-MoneySpans"

	//money-trace context keys
	tIDKey = "trace-id"
	pIDKey = "parent-id"
	sIDKey = "span-id"
)

//Transactor is an HTTP transactor type
type Transactor func(*http.Request) (*http.Response, error)

// Tracker is the management interface for an active span.  It can be used to create
// child spans and to mark the current span as finished.
type Tracker interface {
	Spanner

	// Finish concludes this span with the given result
	Finish(Result)

	// String provides the representation of the managed span
	String() string

	//DecorateTransactor provides a strategy to inspect transactor arguments and outputs
	DecorateTransactor(Transactor, ...SpanForwardingOptions) Transactor

	//Spans returns a list of string-encoded Money spans that have been created under this tracker
	Spans() []string
}

//SpanForwardingOptions allows gathering data from an HTTP response
//into string-encoded golang money spans
//application code is responsible to only inspect the response and if otherwise, put back data
//(i.e if body is read)
//An use case for this is extracting WRP spans into golang money spans
type SpanForwardingOptions func(*http.Response) []string

//HTTPTracker is the management type for child spans
type HTTPTracker struct {
	Spanner
	m    sync.RWMutex
	span Span

	//spans contains the string-encoded value of all spans created under this tracker
	//should be modifiable by multiple goroutines
	spans []string

	done bool //indicates whether the span associated with this tracker is finished
}

//DecorateTransactor configures a transactor to both
//inject Money Trace Context into outgoing requests
//and extract Money Spans from their responses (if any)
func (t *HTTPTracker) DecorateTransactor(transactor Transactor, options ...SpanForwardingOptions) Transactor {
	return func(r *http.Request) (resp *http.Response, e error) {
		t.m.RLock()
		r.Header.Add(MoneyHeader, EncodeTraceContext(t.span.TC))
		t.m.RUnlock()

		if resp, e = transactor(r); e == nil {
			t.m.Lock()
			defer t.m.Unlock()

			//the default behavior is always run
			for k, vs := range resp.Header {
				if k == MoneySpansHeader {
					for _, v := range vs {
						t.spans = append(t.spans, v)
					}
				}
			}

			//options allow converting different span types into money-compatible ones
			for _, o := range options {
				for _, span := range o(resp) {
					t.spans = append(t.spans, span)
				}
			}
		}
		return
	}
}

//Start defines the money trace context for span s based
//on the underlying HTTPTracker span before delegating the
//start process to the Spanner
//if such underlying span has already finished, the returned
//tracker is nil
func (t *HTTPTracker) Start(ctx context.Context, s Span) (tracker Tracker) {
	t.m.RLock()
	defer t.m.RUnlock()

	if !t.done {
		s.TC = SubTrace(t.span.TC)
		tracker = t.Spanner.Start(ctx, s)
	}

	return
}

//Finish is an idempotent operation that marks the end of the underlying HTTPTracker span
func (t *HTTPTracker) Finish(r Result) {
	t.m.Lock()
	defer t.m.Unlock()

	if !t.done {
		t.span.Duration = time.Since(t.span.StartTime)
		t.span.Host, _ = os.Hostname()
		t.span.Name = r.Name
		t.span.AppName = r.AppName
		t.span.Code = r.Code
		t.span.Err = r.Err
		t.span.Success = r.Success

		t.spans = append(t.spans, t.span.String())

		t.done = true
	}
}

//String returns the string representation of the span associated with this
//HTTPTrackertracker once such span has finished, zero value otherwise
func (t *HTTPTracker) String() (v string) {
	t.m.RLock()
	defer t.m.RUnlock()

	if t.done {
		v = t.span.String()
	}

	return
}

//Spans returns the list of string-encoded spans under this tracker
//once the main span under the tracker is finished, zero value otherwise
func (t *HTTPTracker) Spans() (spans []string) {
	t.m.RLock()
	defer t.m.RUnlock()

	if t.done {
		spans = make([]string, len(t.spans))
		copy(spans, t.spans)
	}

	return
}

//TrackerFromContext extracts a tracker contained in the given context, if any
func TrackerFromContext(ctx context.Context) (t Tracker, ok bool) {
	t, ok = ctx.Value(contextKeyTracker).(Tracker)
	return
}
