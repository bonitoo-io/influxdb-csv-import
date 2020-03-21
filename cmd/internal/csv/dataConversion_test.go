package cmd

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test_escapeMeasurement
func Test_escapeMeasurement(t *testing.T) {
	var tests = []struct {
		value  string
		expect string
	}{
		{"a", "a"}, {"", ""},
		{"a,", `a\,`},
		{"a ", `a\ `},
		{"a=", `a=`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.Equal(t, test.expect, escapeMeasurement(test.value))
		})
	}
}

// Test_escapeTag
func Test_escapeTag(t *testing.T) {
	var tests = []struct {
		value  string
		expect string
	}{
		{"a", "a"}, {"", ""},
		{"a,", `a\,`},
		{"a ", `a\ `},
		{"a=", `a\=`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.Equal(t, test.expect, escapeTag(test.value))
		})
	}
}

// Test_quoteValue
func Test_quoteValue(t *testing.T) {
	var tests = []struct {
		value  string
		expect string
	}{
		{"a", `"a"`}, {"", `""`},
		{`a"`, `"a\""`},
		{`a\`, `"a\\"`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.Equal(t, test.expect, quoteValue(test.value))
		})
	}
}

// Test_toTypedValue
func Test_toTypedValue(t *testing.T) {
	epochTime, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	var tests = []struct {
		dataType string
		value    string
		expect   interface{}
	}{
		{"string", "a", "a"},
		{"double", "1.0", 1.0},
		{"boolean", "true", true},
		{"boolean", "false", false},
		{"boolean", "", nil},
		{"long", "1", (int64)(1)},
		{"unsignedLong", "1", (uint64)(1)},
		{"duration", "1ns", time.Duration(1)},
		{"base64Binary", "YWFh", []byte("aaa")},
		{"dateTime:RFC3339", "1970-01-01T00:00:00Z", epochTime},
		{"dateTime:RFC3339Nano", "1970-01-01T00:00:00.0Z", epochTime},
		{"u.type", "", nil},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := toTypedValue(test.value, test.dataType)
			if err != nil && test.expect != nil {
				require.Nil(t, err.Error())
			}
			require.Equal(t, test.expect, val)
		})
	}
}

// Test_toLineProtocolValue
func Test_toLineProtocolValue(t *testing.T) {
	epochTime, _ := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	var tests = []struct {
		value  interface{}
		expect string
	}{
		{uint64(1), "1u"},
		{int64(1), "1i"},
		{int(1), "1i"},
		{float64(1.1), "1.1"},
		{math.NaN(), ""},
		{math.Inf(1), ""},
		{float32(1), "1"},
		{float32(math.NaN()), ""},
		{float32(math.Inf(1)), ""},
		{"a", `"a"`},
		{[]byte("aaa"), "YWFh"},
		{true, "true"},
		{false, "false"},
		{epochTime, "0"},
		{time.Duration(100), "100i"},
		{struct{}{}, ""},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := toLineProtocolValue(test.value)
			if err != nil && test.expect != "" {
				require.Nil(t, err.Error())
			}
			require.Equal(t, test.expect, val)
		})
	}
}

// Test_convert
func Test_convert(t *testing.T) {
	var tests = []struct {
		dataType string
		value    string
		expect   string
	}{
		{"", "1", "1"},
		{"long", "a", ""},
		{"timestamp", "a", ""},
		{"string", "a", `"a"`},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val, err := convert(test.value, test.dataType)
			if err != nil && test.expect != "" {
				require.Nil(t, err.Error())
			}
			require.Equal(t, test.expect, val)
		})
	}
}

// Test_IsTypeSupported
func Test_IsTypeSupported(t *testing.T) {
	require.Equal(t, IsTypeSupported(stringDatatype), true)
	require.Equal(t, IsTypeSupported(doubleDatatype), true)
	require.Equal(t, IsTypeSupported(boolDatatype), true)
	require.Equal(t, IsTypeSupported(longDatatype), true)
	require.Equal(t, IsTypeSupported(uLongDatatype), true)
	require.Equal(t, IsTypeSupported(durationDatatype), true)
	require.Equal(t, IsTypeSupported(base64BinaryDataType), true)
	require.Equal(t, IsTypeSupported(timeDatatypeRFC), true)
	require.Equal(t, IsTypeSupported(timeDatatypeRFCNano), true)
	require.Equal(t, IsTypeSupported(""), true)
	require.Equal(t, IsTypeSupported(" "), false)
}
