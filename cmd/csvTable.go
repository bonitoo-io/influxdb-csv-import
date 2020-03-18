package cmd

import (
	"errors"
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
	setup func(column *TableColumn, value string)
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
	{"#group", 1, func(column *TableColumn, value string) {
		column.Group = strings.HasSuffix(value, "true")
	}},
	{"#datatype", 2, func(column *TableColumn, value string) {
		column.DataType = ignoreLeadingComment(value)
	}},
	{"#default", 4, func(column *TableColumn, value string) {
		column.DefaultValue = ignoreLeadingComment(value)
	}},
	{"#linetype", 8, func(column *TableColumn, value string) {
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
		}
	}},
}

// TableColumn represents metadata of a flux <a href="http://bit.ly/flux-spec#table">table</a>.
type TableColumn struct {
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

// Table gathers metadata about columns
type Table struct {
	// all Table columns
	columns []TableColumn
	// bitmap indicating presence of group, datatype and default comments
	partBits uint8
	// indicates that it is ready to read table data
	readTableData bool
	// indicated whether a table layout has changed
	indexed bool

	/* cached data */

	cachedMeasurement *TableColumn
	cachedTime        *TableColumn
	cachedFieldName   *TableColumn
	cachedFieldValue  *TableColumn
	cachedFields      []TableColumn
	cachedTags        []TableColumn
}

// AddRow adds header, comment or data row
func (t *Table) AddRow(row []string) bool {
	// support just header with #
	if len(row[0]) == 0 || row[0][0] != '#' {
		if !t.readTableData {
			if t.partBits == 0 {
				// create table if it does not exist yet
				t.columns = make([]TableColumn, len(row))
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
			var firstColumnIndex int = 0
			if len(supportedHeader.label) == len(row[0]) {
				firstColumnIndex = 1
			} else {
				if row[0][len(supportedHeader.label)] != ' ' {
					continue // not a comment from the supported header
				}
			}
			t.indexed = false
			t.readTableData = false
			// create new columns when data change (overriding headers or no headers)
			if t.partBits == 0 || t.partBits&supportedHeader.flag == 1 {
				t.partBits = supportedHeader.flag
				t.columns = make([]TableColumn, len(row)-firstColumnIndex)
				for i := 0; i < len(row)-firstColumnIndex; i++ {
					t.columns[i].Index = i + firstColumnIndex
				}
			} else {
				t.partBits = t.partBits | supportedHeader.flag
			}
			for j := 0; j < len(t.columns); j++ {
				col := &t.columns[j]
				if col.Index >= len(row) {
					continue // missing value
				} else {
					supportedHeader.setup(col, row[col.Index])
				}
			}

			return false
		}
	}
	return row[0][0] != '#'
}

func (t *Table) computeIndexes() {
	if !t.indexed {
		t.recomputeIndexes()
	}
}

func (t *Table) recomputeIndexes() {
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

// CreateLine produces a protocol line of the supplied row or nil
func (t *Table) CreateLine(row []string) (line string, err error) {
	t.computeIndexes()

	// TODO escape values
	if t.cachedMeasurement == nil {
		return "", errors.New("no measurement column identified")
	}
	val := row[t.cachedMeasurement.Index]
	for _, tag := range t.cachedTags {
		val += "," + tag.Label + "=" + row[tag.Index]
	}
	val += " "
	fieldAdded := false
	if t.cachedFieldName != nil {
		val += row[t.cachedFieldName.Index] + "=" + row[t.cachedFieldValue.Index]
		fieldAdded = true
	}
	for _, field := range t.cachedFields {
		if !fieldAdded {
			fieldAdded = true
		} else {
			val += ","
		}
		val += field.Label + "=" + row[field.Index]
	}
	if !fieldAdded {
		return "", errors.New("no field columns found")
	}

	if t.cachedTime != nil {
		val += " " + row[t.cachedTime.Index]
	}
	return val, nil
}

// Column returns the first column of the supplied label or nil
func (t *Table) Column(label string) *TableColumn {
	for i := 0; i < len(t.columns); i++ {
		if t.columns[i].Label == label {
			return &t.columns[i]
		}
	}
	return nil
}

// Columns returns available columns
func (t *Table) Columns() []TableColumn {
	return t.columns
}

// Measurement returns measurement column or nil
func (t *Table) Measurement() *TableColumn {
	t.computeIndexes()
	return t.cachedMeasurement
}

// Time returns time column or nil
func (t *Table) Time() *TableColumn {
	t.computeIndexes()
	return t.cachedTime
}

// FieldName returns field name column or nil
func (t *Table) FieldName() *TableColumn {
	t.computeIndexes()
	return t.cachedFieldName
}

// FieldValue returns field value column or nil
func (t *Table) FieldValue() *TableColumn {
	t.computeIndexes()
	return t.cachedFieldValue
}

// Tags returns tags
func (t *Table) Tags() []TableColumn {
	t.computeIndexes()
	return t.cachedTags
}

// Fields returns fields
func (t *Table) Fields() []TableColumn {
	t.computeIndexes()
	return t.cachedFields
}
