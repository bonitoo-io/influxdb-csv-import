package cmd

const (
	labelFieldName   = "_field"
	labelFieldValue  = "_value"
	labelTime        = "_time"
	labelStartTime   = "_start"
	labelStopTime    = "_stop"
	labelMeasurement = "_measurement"
)

type headerDescriptor struct {
	label string
	flag  uint8
	setup func(column *TableColumn, value string)
}

var headerTypes = []headerDescriptor{
	{"#group", 1, func(column *TableColumn, value string) {
		column.Group = value == "true"
	}},
	{"#datatype", 2, func(column *TableColumn, value string) {
		column.DataType = value
	}},
	{"#default", 4, func(column *TableColumn, value string) {
		column.DefaultValue = value
	}},
}

// TableColumn represents metadata of a flux <a href="http://bit.ly/flux-spec#table">table</a>.
type TableColumn struct {
	// label such as "_start", "_stop", "_time"
	Label string
	// "string", "long", "dateTime:RFC3339" ...
	DataType string
	// table's group key indicator
	Group bool
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
	cachedTags        []TableColumn
}

// AddRow adds a header row that defines layout of the table
func (t *Table) AddRow(row []string) bool {
	// support just header with #
	if len(row[0]) == 0 || row[0][0] != '#' {
		if !t.readTableData {
			for i := 0; i < len(t.columns); i++ {
				col := &t.columns[i]
				if col.Index >= len(row) {
					continue // missing value
				} else {
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
		if row[0] == supportedHeader.label {
			t.indexed = false
			t.readTableData = false
			// create new columns when data change (overriding headers or no headers)
			if t.partBits == 0 || t.partBits&supportedHeader.flag == 1 {
				t.partBits = supportedHeader.flag
				t.columns = make([]TableColumn, len(row)-1)
				for i := 0; i < len(row)-1; i++ {
					t.columns[i].Index = i + 1
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
	return true
}

func (t *Table) computeIndexes() {
	t.cachedMeasurement = nil
	t.cachedTime = nil
	t.cachedFieldName = nil
	t.cachedFieldValue = nil
	t.cachedTags = nil
	for i := 0; i < len(t.columns); i++ {
		col := t.columns[i]
		switch {
		case col.Label == labelMeasurement:
			t.cachedMeasurement = &col
		case col.Label == labelTime:
			t.cachedTime = &col
		case col.Label == labelFieldName:
			t.cachedFieldName = &col
		case col.Label == labelFieldValue:
			t.cachedFieldValue = &col
		case col.Group && col.Label != labelStartTime && col.Label != labelStopTime:
			t.cachedTags = append(t.cachedTags, col)
		}
	}
	t.indexed = true
}

// CreateMetric produces a metric instance out of the supplied row
func (t *Table) CreateMetric(row []string) interface{} {
	if !t.indexed {
		t.computeIndexes()
	}
	// TODO create a real metric
	val := row[t.cachedMeasurement.Index]
	for _, tag := range t.cachedTags {
		val += "," + tag.Label + "=" + row[tag.Index]
	}
	val += " "
	val += row[t.cachedFieldName.Index] + "=" + row[t.cachedFieldValue.Index]
	if t.cachedTime != nil {
		val += " " + row[t.cachedTime.Index]
	}
	return val
}

// // Column returns the first column of the supplied label or nil
// func (t *Table) Column(label string) *TableColumn {
// 	for i := 0; i < len(t.columns); i++ {
// 		if t.columns[i].Label == label {
// 			return &t.columns[i]
// 		}
// 	}
// 	return nil
// }

// // Columns returns available columns
// func (t *Table) Columns() []TableColumn {
// 	return t.columns
// }

// // Measurement returns measurement column or nil
// func (t *Table) Measurement() *TableColumn {
// 	if !t.indexed {
// 		t.computeIndexes()
// 	}
// 	return t.cachedMeasurement
// }

// // Time returns time column or nil
// func (t *Table) Time() *TableColumn {
// 	if !t.indexed {
// 		t.computeIndexes()
// 	}
// 	return t.cachedTime
// }

// // FieldName returns field name column or nil
// func (t *Table) FieldName() *TableColumn {
// 	if !t.indexed {
// 		t.computeIndexes()
// 	}
// 	return t.cachedFieldName
// }

// // FieldValue returns field value column or nil
// func (t *Table) FieldValue() *TableColumn {
// 	if !t.indexed {
// 		t.computeIndexes()
// 	}
// 	return t.cachedFieldValue
// }

// // Tags returns tags
// func (t *Table) Tags() []TableColumn {
// 	if !t.indexed {
// 		t.computeIndexes()
// 	}
// 	return
// 	t.cachedTags
// }
