package money

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPSpanner(t *testing.T) {
	t.Run("DI", testNewHTTPSpannerNil)
	t.Run("Start", testStart)
	t.Run("DecorationNoMoneyContext", testDecorate)
	t.Run("DecorationMoneyContext", testDecorateWithMoney)
	t.Run("TrackerFinish", testTrackerFinish)
}

func testNewHTTPSpannerNil(t *testing.T) {
	var spanner *HTTPSpanner
	if spanner.Decorate("test", nil) != nil {
		t.Error("Decoration should leave handler unchanged")
	}
}

func testStart(t *testing.T) {
	var spanner = NewHTTPSpanner()
	if spanner.Start(context.Background(), Span{}) == nil {
		t.Error("was expecting a non-nil response")
	}
}

func testDecorate(t *testing.T) {
	var spanner = NewHTTPSpanner()

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if ok {
				t.Error("Tracker should not have been injected")
			}
		})
	decorated := spanner.Decorate("test", handler)
	decorated.ServeHTTP(nil, httptest.NewRequest("GET", "localhost:9090/test", nil))
}

func testDecorateWithMoney(t *testing.T) {
	var spanner = NewHTTPSpanner()

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if !ok {
				t.Error("Expected tracker to be present")
			}
		})
	decorated := spanner.Decorate("test", handler)
	inputRequest := httptest.NewRequest("GET", "localhost:9090/test", nil)
	inputRequest.Header.Add(MoneyHeader, "trace-id=abc;parent-id=1;span-id=1")
	decorated.ServeHTTP(nil, inputRequest)
}

func testTrackerFinish(t *testing.T) {
	var (
		assert  = assert.New(t)
		spanner *HTTPSpanner
		tracker *HTTPTracker
		s       Span
		r       Result
	)

	startTime, e := time.Parse(time.RFC3339, "1970-01-01T00:00:01+00:00") //1 second into epoch time = 1,000,000 microseconds
	require.Nil(t, e)

	spanner = NewHTTPSpanner()

	s = Span{
		Name:    "test-tracker-span",
		AppName: "test-app",
		TC: &TraceContext{
			TID: "test-tracker-trace",
			SID: 11,
			PID: 11,
		},
		Success:   false,
		Code:      200,
		StartTime: startTime,
		Duration:  time.Second,
		Host:      "localhost",
	}

	tracker = spanner.Start(context.Background(), s).(*HTTPTracker)
	require.NotNil(t, tracker)

	tracker.m = sync.RWMutex{}
	r = Result{
		Name:    "test-result-span",
		AppName: "test-finish-app",
		Success: true,
		Code:    201,
		Err:     nil,
	}

	tracker.Finish(r)
	assert.Equal(tracker.span.Name, r.Name)
	assert.Equal(tracker.span.AppName, r.AppName)
	assert.Equal(tracker.span.Code, r.Code)
	assert.Equal(tracker.span.Err, r.Err)
	assert.Equal(tracker.span.Success, r.Success)
}
