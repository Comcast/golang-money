package money

import (
	"reflect"
	"testing"
)

func TestDecodeTraceContext(t *testing.T) {
	tests := []struct {
		name string
		i    string
		o    *TraceContext
		e    error
	}{
		{
			i: "",
			o: nil,
			e: errPairsCount,
		},
		{
			name: "ideal",
			i:    "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;parent-id=3285573610483682037;span-id=3285573610483682037",
			o: &TraceContext{
				PID: 3285573610483682037,
				SID: 3285573610483682037,
				TID: "de305d54-75b4-431b-adb2-eb6b9e546013",
			},
			e: nil,
		},
		{
			name: "extra",
			i:    "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;parent-id=3285573610483682037;span-id=3285573610483682037;sampled=1",
			o: &TraceContext{
				PID: 3285573610483682037,
				SID: 3285573610483682037,
				TID: "de305d54-75b4-431b-adb2-eb6b9e546013",
			},
			e: nil,
		},

		{
			name: "duplicateEntries",
			i:    "parent-id=1;parent-id=3285573610483682037;span-id=3285573610483682037",
			o:    nil,
			e:    errBadTrace,
		},
		{
			name: "badPair",
			i:    "parent-id=de305d54-75b=4-431b-adb2-eb6b9e546013;parent-id=3285573610483682037;span-id=3285573610483682037",
			o:    nil,
			e:    errBadPair,
		},

		{
			name: "NoRealPairs",
			i:    "one=1;two=2;three=3",
			o:    nil,
			e:    errBadTrace,
		},

		{
			name: "missingTraceId",
			i:    "foo=bar;parent-id=3285573610483682037;span-id=3285573610483682037",
			o:    nil,
			e:    errBadTrace,
		},

		{
			name: "missingParentId",
			i:    "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;foo=bar;span-id=3285573610483682037",
			o:    nil,
			e:    errBadTrace,
		},

		{
			name: "missingSpanId",
			i:    "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;parent-id=3285573610483682037;foo=bar",
			o:    nil,
			e:    errBadTrace,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualO, actualE := decodeTraceContext(test.i)
			if actualE != test.e || !reflect.DeepEqual(actualO, test.o) {
				t.Errorf("I was expecting '%v' '%v'", test.e, test.o)
				t.Errorf("but got '%v' '%v'", actualE, actualO)
			}
		})
	}
}

func TestDecodeTraceContextOtherCases(t *testing.T) {
	t.Run("NonIntSID", func(t *testing.T) {
		i := "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;parent-id=3285573610483682037;span-id=NaN"
		tc, e := decodeTraceContext(i)

		if tc != nil || e == nil {
			t.Errorf("expected tc to be nil and error to be non-nil but got '%v' and '%v'", tc, e)
		}
	})

	t.Run("NonIntPID", func(t *testing.T) {
		i := "trace-id=de305d54-75b4-431b-adb2-eb6b9e546013;parent-id=NaN;span-id=123"
		tc, e := decodeTraceContext(i)

		if tc != nil || e == nil {
			t.Errorf("expected tc to be nil and error to be non-nil but got '%v' and '%v'", tc, e)
		}
	})
}

//TODO: need to if a string is in the right order accordance.
// Tests if typeInferenceTC  outputs a the fields of a TID in the correct order.
func TestTypeInferenceTC(t *testing.T) {
	in := map[string]interface{}{
		"PID": 1,
		"SID": 1,
		"TID": "one",
	}

	var expected = "parent-id=1;span-id=1;trace-id=one"
	var actual = typeInferenceTC(in)

	if actual != expected {
		t.Errorf("Wrong Format for Trace Context string, need '%v' but got '%v'", expected, actual)
	}
}

func TestEncodeTraceContext(t *testing.T) {
	in := &TraceContext{
		PID: 1,
		TID: "one",
		SID: 1,
	}
	actual, expected := EncodeTraceContext(in), "parent-id=1;span-id=1;trace-id=one"

	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSubtrace(t *testing.T) {
	current := &TraceContext{
		TID: "123",
		SID: 1,
	}
	st := SubTrace(current)

	if st.PID != current.SID {
		t.Errorf("Expected pid to be %v but got %v", current.PID, st.PID)
	}

	if st.SID == 0 {
		t.Error("Expected sid to be defined")
	}

	if st.TID != current.TID {
		t.Errorf("Expected tid to be %v but got %v", current.TID, st.TID)
	}
}
