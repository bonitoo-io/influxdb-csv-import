package cmd

import (
	"fmt"
	"log"
	"strings"
)

const (
	labelFieldName   = "_field"
	labelFieldValue  = "_value"
	labelTime        = "_time"
	labelMeasurement = "_measurement"
)

type headerDescriptor struct {
	label string
	flag  uint8
	setup func(column *CsvTableColumn, value string) error
}

func ignoreLeadingComment(value string) string {
	if len(value) > 0 && value[0] == '#' {
		pos := strings.Index(value, " ")
		if pos > 0 {
			return strings.TrimLeft(value[pos+1:], " ")
		}
		return ""
	}
	return value
}

var headerTypes = []headerDescriptor{
	{"#group", 1, func(column *CsvTableColumn, value string) error {
		column.Group = strings.HasSuffix(value, "true")
		return nil
	}},
	{"#datatype", 2, func(column *CsvTableColumn, value string) error {
		column.DataType = ignoreLeadingComment(value)
		return nil
	}},
	{"#default", 4, func(column *CsvTableColumn, value string) error {
		column.DefaultValue = ignoreLeadingComment(value)
		return nil
	}},
	{"#linetype", 8, func(column *CsvTableColumn, value string) error {
		val := ignoreLeadingComment(value)
		switch {
		case val == "tag":
			column.Group = true
		case strings.HasPrefix(val, "ignore"):
			column.Ignored = true
		case val == "time":
			column.Label = labelTime
		case val == "measurement":
			column.Label = labelMeasurement
		case val == "field" || val == "":
			// detect line type
		default:
			return fmt.Errorf("unsupported line type: %s", value)
		}
		return nil
	}},
}

// CsvTableColumn represents metadata of a flux <a href="http://bit.ly/flux-spec#table">table</a>.
type CsvTableColumn struct {
	// label such as "_start", "_stop", "_time"
	Label string
	// "string", "long", "dateTime:RFC3339" ...
	DataType string
	// table's group/tag indicator
	Group bool
	// is this column ignored
	Ignored bool
	// default value to be used for rows where value is an empty string.
	DefaultValue string
	// index of this column in the table row
	Index int
}

// CsvTable gathers metadata about columns
type CsvTable struct {
	// all Table columns
	columns []CsvTableColumn
	// bitmap indicating presence of group, datatype and default comments
	partBits uint8
	// indicates that it is ready to read table data
	readTableData bool
	// indicated whether a table layout has changed
	indexed bool

	/* cached data */

	cachedMeasurement *CsvTableColumn
	cachedTime        *CsvTableColumn
	cachedFieldName   *CsvTableColumn
	cachedFieldValue  *CsvTableColumn
	cachedFields      []CsvTableColumn
	cachedTags        []CsvTableColumn
}

// AddRow adds header, comment or data row
func (t *CsvTable) AddRow(row []string) bool {
	// support just header with #
	if len(row[0]) == 0 || row[0][0] != '#' {
		if !t.readTableData {
			if t.partBits == 0 {
				// create table if it does not exist yet
				t.columns = make([]CsvTableColumn, len(row))
				for i := 0; i < len(row); i++ {
					t.columns[i].Index = i
				}
			}
			for i := 0; i < len(t.columns); i++ {
				col := &t.columns[i]
				if len(col.Label) == 0 && col.Index < len(row) {
					col.Label = row[col.Index]
				}
			}
			t.readTableData = true
			return false
		}
		return true
	}
	for i := 0; i < len(headerTypes); i++ {
		supportedHeader := &headerTypes[i]
		if strings.HasPrefix(strings.ToLower(row[0]), supportedHeader.label) {
			if len(row[0]) > len(supportedHeader.label) && row[0][len(supportedHeader.label)] != ' ' {
				continue // not a comment from the supported header
			}
			t.indexed = false
			t.readTableData = false
			// create new columns when data change (overriding headers or no headers)
			if t.partBits == 0 || t.partBits&supportedHeader.flag == 1 {
				t.partBits = supportedHeader.flag
				t.columns = make([]CsvTableColumn, len(row))
				for i := 0; i < len(row); i++ {
					t.columns[i].Index = i
				}
			} else {
				t.partBits = t.partBits | supportedHeader.flag
			}
			for j := 0; j < len(t.columns); j++ {
				col := &t.columns[j]
				if col.Index >= len(row) {
					continue // missing value
				} else {
					err := supportedHeader.setup(col, row[col.Index])
					if err != nil {
						log.Println("WARNING:", err)
					}
				}
			}

			return false
		}
	}
	return row[0][0] != '#'
}

func (t *CsvTable) computeIndexes() bool {
	if !t.indexed {
		t.recomputeIndexes()
		return true
	}
	return false
}

func (t *CsvTable) recomputeIndexes() {
	t.cachedMeasurement = nil
	t.cachedTime = nil
	t.cachedFieldName = nil
	t.cachedFieldValue = nil
	t.cachedTags = nil
	t.cachedFields = nil
	canContainFields := t.Column(labelFieldName) == nil
	for i := 0; i < len(t.columns); i++ {
		col := t.columns[i]
		switch {
		case len(strings.TrimSpace(col.Label)) == 0 || col.Ignored:
			break
		case col.Label == labelMeasurement:
			t.cachedMeasurement = &col
		case col.Label == labelTime:
			t.cachedTime = &col
		case col.Label == labelFieldName:
			t.cachedFieldName = &col
		case col.Label == labelFieldValue:
			t.cachedFieldValue = &col
		case col.Label[0] == '_':
			break
		case col.Group:
			t.cachedTags = append(t.cachedTags, col)
		default:
			if canContainFields {
				t.cachedFields = append(t.cachedFields, col)
			}
		}
	}

	t.indexed = true
}

// CreateLine produces a protocol line of the supplied row or returned error
func (t *CsvTable) CreateLine(row []string) (line string, err error) {
	builder := strings.Builder{}
	builder.Grow(100)
	err = t.AppendLine(&builder, row)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

// AppendLine appends a protocol to the supplied builder or returns non-nil error
func (t *CsvTable) AppendLine(builder *strings.Builder, row []string) error {
	if t.computeIndexes() {
		// validate column data types
		if t.cachedFieldValue != nil && !IsTypeSupported(t.cachedFieldValue.DataType) {
			return fmt.Errorf("column '%s': data type '%s' is not supported", t.cachedFieldValue.Label, t.cachedFieldValue.DataType)
		}
		for _, c := range t.cachedFields {
			if !IsTypeSupported(c.DataType) {
				return fmt.Errorf("column '%s': data type '%s' is not supported", c.Label, c.DataType)
			}
		}
	}

	if t.cachedMeasurement == nil {
		return fmt.Errorf("no measurement column found (columns: %d)", len(t.columns))
	}
	builder.WriteString(escapeMeasurement(row[t.cachedMeasurement.Index]))
	for _, tag := range t.cachedTags {
		if len(row[tag.Index]) > 0 {
			builder.WriteString(",")
			builder.WriteString(escapeTag(tag.Label))
			builder.WriteString("=")
			builder.WriteString(escapeTag(row[tag.Index]))
		}
	}
	builder.WriteString(" ")
	fieldAdded := false
	if t.cachedFieldName != nil && t.cachedFieldValue != nil {
		converted, err := convert(row[t.cachedFieldValue.Index], t.cachedFieldValue.DataType)
		if err != nil {
			return err
		}
		if len(converted) > 0 {
			builder.WriteString(escapeTag(row[t.cachedFieldName.Index]))
			builder.WriteString("=")
			builder.WriteString(converted)
			fieldAdded = true
		}
	}
	for _, field := range t.cachedFields {
		converted, err := convert(row[field.Index], field.DataType)
		if err != nil {
			return err
		}
		if len(converted) > 0 {
			if !fieldAdded {
				fieldAdded = true
			} else {
				builder.WriteString(",")
			}
			builder.WriteString(escapeTag(field.Label))
			builder.WriteString("=")
			builder.WriteString(converted)
		}
	}
	if !fieldAdded {
		return fmt.Errorf("no field column found (columns: %d)", len(t.columns))
	}

	if t.cachedTime != nil {
		timeVal := row[t.cachedTime.Index]
		var dataType = t.cachedTime.DataType
		if timeVal != "" && dataType == "" {
			//try to detect data type
			if strings.Index(timeVal, ".") >= 0 {
				dataType = "dateTime:RFC3339Nano"
			} else if strings.Index(timeVal, "-") >= 0 {
				dataType = "dateTime:RFC3339"
			}
		}
		timeVal, err := convert(timeVal, dataType)
		if err != nil {
			return err
		}
		if timeVal != "" {
			builder.WriteString(" ")
			builder.WriteString(timeVal)
		}
	}
	return nil
}

// Column returns the first column of the supplied label or nil
func (t *CsvTable) Column(label string) *CsvTableColumn {
	for i := 0; i < len(t.columns); i++ {
		if t.columns[i].Label == label {
			return &t.columns[i]
		}
	}
	return nil
}

// Columns returns available columns
func (t *CsvTable) Columns() []CsvTableColumn {
	return t.columns
}

// Measurement returns measurement column or nil
func (t *CsvTable) Measurement() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedMeasurement
}

// Time returns time column or nil
func (t *CsvTable) Time() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedTime
}

// FieldName returns field name column or nil
func (t *CsvTable) FieldName() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedFieldName
}

// FieldValue returns field value column or nil
func (t *CsvTable) FieldValue() *CsvTableColumn {
	t.computeIndexes()
	return t.cachedFieldValue
}

// Tags returns tags
func (t *CsvTable) Tags() []CsvTableColumn {
	t.computeIndexes()
	return t.cachedTags
}

// Fields returns fields
func (t *CsvTable) Fields() []CsvTableColumn {
	t.computeIndexes()
	return t.cachedFields
}
