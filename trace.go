package money

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Trace Context decoding errors
var (
	errPairsCount = errors.New("expecting three pairs in trace context")
	errBadPair    = errors.New("expected trace context header to have pairs")
	errBadTrace   = errors.New("malformatted trace context header")
)

// TraceContext encapsutes all the core information of any given span
// In a single trace, the TID is the same across all spans and
// the SID and the PID is what links all spans together
type TraceContext struct {
	TID string //Trace ID
	SID int64  //Span ID
	PID int64  //Parent ID
}

// decodeTraceContext returns a TraceContext from the given value "raw"
// raw is typically taken directly from http.Request headers
// for now, it is overly strict with the expected format
// TODO: could we use regex here instead for simplicity?
func decodeTraceContext(raw string) (tc *TraceContext, err error) {
	tc = new(TraceContext)

	pairs := strings.Split(raw, ";")

	if len(pairs) < 3 {
		return nil, errPairsCount
	}

	seen := make(map[string]bool)

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")

		if len(kv) != 2 {
			return nil, errBadPair
		}

		var k, v = kv[0], kv[1]

		switch {
		case k == tIDKey && !seen[k]:
			tc.TID, seen[k] = v, true

		case k == sIDKey && !seen[k]:
			var pv int64
			if pv, err = strconv.ParseInt(v, 10, 64); err != nil {
				return nil, err
			}

			tc.SID, seen[k] = pv, true
		case k == pIDKey && !seen[k]:
			var pv int64
			if pv, err = strconv.ParseInt(v, 10, 64); err != nil {
				return nil, err
			}
			tc.PID, seen[k] = pv, true
		}
	}

	if (!seen[tIDKey] || !seen[sIDKey] || !seen[pIDKey]) {
		return nil, errBadTrace
	}

	return
}

// typeInferenceTC  returns a concatenated string of all field values that exist in a trace context from a map[string]interface{}
func typeInferenceTC(tc interface{}) string {
	tcs := tc.(map[string]interface{})

	m := map[string]string{}

	for k, v := range tcs {
		switch v.(type) {
		case int:
			m[k] = fmt.Sprintf("%v", tcs[k].(int))
		case float64:
			m[k] = fmt.Sprintf("%v", tcs[k].(float64))
		case string:
			m[k] = tcs[k].(string)
		}
	}

	return fmt.Sprintf("%s=%v;%s=%v;%s=%v", pIDKey, m["PID"], sIDKey, m["SID"], tIDKey, m["TID"])
}

// EncodeTraceContext encodes the TraceContext into a string.
func encodeTraceContext(tc *TraceContext) string {
	return fmt.Sprintf("%s=%v;%s=%v;%s=%v", pIDKey, tc.PID, sIDKey, tc.SID, tIDKey, tc.TID)
}

// This is useful if you want to pass your trace context over an outgoing request or just need a string formatted trace context for any other purpose.
func EncodeTraceContext(tc *TraceContext) string {
	return encodeTraceContext(tc)
}

// SubTrace creates a child trace context for current
func SubTrace(current *TraceContext) *TraceContext {
	rand.Seed(time.Now().Unix())
	return &TraceContext{
		PID: current.SID,
		SID: rand.Int63(),
		TID: current.TID,
	}
}
