package money

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type MockResponse struct {
	http.ResponseWriter
	headers    http.Header
	statusCode int
}

func NewMockResponse() *MockResponse {
	return &MockResponse{
		headers: make(http.Header),
	}
}

func TestNewHTTPSpanner(t *testing.T) {
	t.Run("Start", testStart)
	//	t.Run("DecorationNoOptions", testDecorateNoOptions)
	//	t.Run("DecorationSpanDecoderON", testDecorateSpanDecoderON)
	//	t.Run("DecorationSubTracerON", testDecorateSubTracerON)
}

func testNewHTTPSpannerNil(t *testing.T) {
	var spanner *HTTPSpanner
	if spanner.Decorate(nil) != nil {
		t.Error("Decoration should leave handler unchanged")
	}
}

func testStart(t *testing.T) {
	var spanner = NewHTTPSpanner()
	if spanner.Start(context.Background(), &Span{}) == nil {
		t.Error("was expecting a non-nil response")
	}
}

func TestDecorateOff(t *testing.T) {
	var spanner = NewHTTPSpanner(SpannerOff())

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if ok {
				t.Error("Tracker should not have been injected")
			}
		})
	decorated := spanner.Decorate(handler)
	decorated.ServeHTTP(nil, httptest.NewRequest("GET", "localhost:9090/test", nil))
}

func TestDecorateSubTracerON(t *testing.T) {
	var (
		pipe   = make(chan *HTTPTracker)
		mockTC = &TraceContext{
			PID: 1,
			SID: 1,
			TID: "1",
		}

		mockSpan = &Span{
			Name: "spantest",
			TC:   mockTC,
		}

		mockHT = &HTTPTracker{
			span: mockSpan,
		}
		spanner = NewHTTPSpanner(SubTracerON(pipe))
	)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if !ok {
				t.Error("Expected tracker to be present")
			}

		})

	decorated := spanner.Decorate(handler)
	inputRequest := httptest.NewRequest("GET", "localhost:9090/test", nil)
	inputRequest.Header.Add(MoneyHeader, "trace-id=abc;parent-id=1;span-id=1")

	var r = httptest.NewRecorder()
	decorated.ServeHTTP(r, InjectTracker(inputRequest, mockHT))

	ht := <-pipe
	spew.Dump(ht)
}

func TestDecorateStarterON(t *testing.T) {
	var (
		pipe    = make(chan *HTTPTracker)
		spanner = NewHTTPSpanner(StarterON(pipe))
	)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if !ok {
				t.Error("Expected tracker to be present")
			}

		})

	decorated := spanner.Decorate(handler)
	inputRequest := httptest.NewRequest("GET", "localhost:9090/test", nil)
	inputRequest.Header.Add(MoneyHeader, "trace-id=abc;parent-id=1;span-id=1")
	var r = httptest.NewRecorder()
	decorated.ServeHTTP(r, inputRequest)
}

func TestDecorateEnderON(t *testing.T) {
	var (
		mockTC = &TraceContext{
			PID: 1,
			SID: 1,
			TID: "1",
		}

		mockSpan = &Span{
			Name: "spantest",
			TC:   mockTC,
		}

		mockHT = &HTTPTracker{
			span: mockSpan,
		}
		spanner = NewHTTPSpanner(EnderON())
	)
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, ok := TrackerFromContext(r.Context())
			if !ok {
				t.Error("Expected tracker to be present")
			}

		})

	decorated := spanner.Decorate(handler)
	inputRequest := httptest.NewRequest("GET", "localhost:9090/test", nil)
	inputRequest.Header.Add(MoneyHeader, "trace-id=abc;parent-id=1;span-id=1")
	var r = httptest.NewRecorder()
	decorated.ServeHTTP(r, InjectTracker(inputRequest, mockHT))
}