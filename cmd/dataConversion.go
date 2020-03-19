package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// see https://v2.docs.influxdata.com/v2.0/reference/syntax/annotated-csv/#valid-data-types
const (
	stringDatatype       = "string"
	doubleDatatype       = "double"
	boolDatatype         = "boolean"
	longDatatype         = "long"
	uLongDatatype        = "unsignedLong"
	durationDatatype     = "duration"
	base64BinaryDataType = "base64Binary"
	timeDatatypeRFC      = "dateTime:RFC3339"
	timeDatatypeRFCNano  = "dateTime:RFC3339Nano"
)

var replaceMeasurement *strings.Replacer = strings.NewReplacer(",", "\\,", " ", "\\ ")
var replaceTag *strings.Replacer = strings.NewReplacer(",", "\\,", " ", "\\ ", "=", "\\=")
var replaceQuoted *strings.Replacer = strings.NewReplacer("\"", "\\\"", "\\", "\\\\")

func escapeMeasurement(val string) string {
	for i := 0; i < len(val); i++ {
		if val[i] == ',' || val[i] == ' ' {
			return replaceMeasurement.Replace(val)
		}
	}
	return val
}
func escapeTag(val string) string {
	for i := 0; i < len(val); i++ {
		if val[i] == ',' || val[i] == ' ' || val[i] == '=' {
			return replaceTag.Replace(val)
		}
	}
	return val
}
func quoteValue(val string) string {
	for i := 0; i < len(val); i++ {
		if val[i] == '"' || val[i] == '\\' {
			return "\"" + replaceQuoted.Replace(val) + "\""
		}
	}
	return "\"" + val + "\""
}

func toTypedValue(val string, dataType string) (interface{}, error) {
	switch dataType {
	case stringDatatype:
		return val, nil
	case timeDatatypeRFC:
		return time.Parse(time.RFC3339, val)
	case timeDatatypeRFCNano:
		return time.Parse(time.RFC3339Nano, val)
	case durationDatatype:
		return time.ParseDuration(val)
	case doubleDatatype:
		return strconv.ParseFloat(val, 64)
	case boolDatatype:
		if val == "true" {
			return true, nil
		} else if val == "false" {
			return false, nil
		}
		return nil, errors.New("Unsupported boolean value '" + val + "' , expected 'true' or 'false'")
	case longDatatype:
		return strconv.ParseInt(val, 10, 64)
	case uLongDatatype:
		return strconv.ParseUint(val, 10, 64)
	case base64BinaryDataType:
		return base64.StdEncoding.DecodeString(val)
	default:
		return nil, fmt.Errorf("%s has unsupported data type %s", val, dataType)
	}
}

func toLineProtocolValue(value interface{}) (string, error) {
	switch v := value.(type) {
	case uint64:
		return strconv.FormatUint(v, 10) + "u", nil
	case int64:
		return strconv.FormatInt(v, 10) + "i", nil
	case int:
		return strconv.FormatInt(int64(v), 10) + "i", nil
	case float64:
		if math.IsNaN(v) {
			return "", errors.New("value is NaN")
		}
		if math.IsInf(v, 0) {
			return "", errors.New("value is Infinite")
		}
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case float32:
		v32 := float64(v)
		if math.IsNaN(v32) {
			return "", errors.New("value is NaN")
		}
		if math.IsInf(v32, 0) {
			return "", errors.New("value is Infinite")
		}
		return strconv.FormatFloat(v32, 'f', -1, 64), nil
	case string:
		return quoteValue(v), nil
	case []byte:
		return base64.StdEncoding.EncodeToString(v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case time.Time:
		return strconv.FormatInt(v.UnixNano(), 10), nil
	case time.Duration:
		return strconv.FormatInt(v.Nanoseconds(), 10) + "i", nil
	default:
		return "", fmt.Errorf("unsupported value type: %T", v)
	}
}

func convert(val string, dataType string) (string, error) {
	if len(dataType) == 0 { // keep the value as it is
		return val, nil
	}
	typedVal, err := toTypedValue(val, dataType)
	if err != nil {
		return "", err
	}
	return toLineProtocolValue(typedVal)
}
