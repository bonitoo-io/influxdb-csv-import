package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func readCsv(t *testing.T, data string) [][]string {
	reader := csv.NewReader(strings.NewReader(data))
	var rows [][]string
	for {
		row, err := reader.Read()
		reader.FieldsPerRecord = 0 // every row can have different number of fields
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		rows = append(rows, row)
	}
	return rows
}

// TestQueryResult validates construction of table columns from Query CSV result
func TestQueryResult(t *testing.T) {
	const csvQueryResult = `
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:17:57Z,0,time_steal,cpu,cpu1,rsavage.prod
,,0,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:07Z,0,time_steal,cpu,cpu1,rsavage.prod

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,1,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:01Z,2.7263631815907954,usage_user,cpu,cpu-total,tahoecity.prod
,,1,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:11Z,2.247752247752248,usage_user,cpu,cpu-total,tahoecity.prod
#unre`

	var lineProtocolQueryResult = []string{
		"cpu,cpu=cpu1,host=rsavage.prod time_steal=0 2020-02-25T22:17:57Z",
		"cpu,cpu=cpu1,host=rsavage.prod time_steal=0 2020-02-25T22:18:07Z",
		"cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.7263631815907954 2020-02-25T22:18:01Z",
		"cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.247752247752248 2020-02-25T22:18:11Z",
	}

	table := Table{}
	rows := readCsv(t, csvQueryResult)
	lineProtocolIndex := 0
	for i, row := range rows {
		rowProcessed := table.AddRow(row)
		fmt.Println(row)
		if i%6 < 4 {
			require.Equal(t, rowProcessed, false, "row %d", i)
		} else {
			require.Equal(t, rowProcessed, true, "row %d", i)
			line, _ := table.CreateLine(row)
			require.Equal(t, line, lineProtocolQueryResult[lineProtocolIndex])
			lineProtocolIndex++
			if i%6 == 4 {
				// verify table
				require.Equal(t, len(table.columns), 10)
				require.Equal(t, table.columns, table.Columns())
				for j, col := range table.columns {
					require.Equal(t, col.Index, j+1)
					require.Equal(t, col.Label, rows[i-1][j+1])
					if len(rows[i-2]) > j+1 {
						require.Equal(t, col.DefaultValue, rows[i-2][j+1])
					} else {
						// some traling data are missing
						require.Equal(t, col.DefaultValue, "")
					}
					require.Equal(t, col.DataType, rows[i-3][j+1])
					require.Equal(t, col.Group, rows[i-4][j+1] == "true")
				}
				// verify cached values
				table.computeIndexes()
				require.Equal(t, table.Column("_measurement"), table.cachedMeasurement)
				require.Nil(t, table.Column("_no"))
				require.NotNil(t, table.cachedMeasurement)
				require.NotNil(t, table.cachedFieldName)
				require.NotNil(t, table.cachedFieldValue)
				require.NotNil(t, table.cachedTime)
				require.NotNil(t, table.cachedTags)
				require.Equal(t, table.Measurement().Label, "_measurement")
				require.Equal(t, table.FieldName().Label, "_field")
				require.Equal(t, table.FieldValue().Label, "_value")
				require.Equal(t, table.Time().Label, "_time")
				require.Equal(t, len(table.Tags()), 2)
				require.Equal(t, table.Tags()[0].Label, "cpu")
				require.Equal(t, table.Tags()[1].Label, "host")
				require.Equal(t, len(table.Fields()), 0)
			}
		}
	}
}

//Test_ignoreLeadingComment
func Test_ignoreLeadingComment(t *testing.T) {
	var tests = []struct {
		value  string
		expect string
	}{
		{"", ""},
		{"a", "a"},
		{" #whatever", " #whatever"},
		{"#whatever", ""},
		{"#whatever ", ""},
		{"#whatever a b ", "a b "},
		{"#whatever  a b ", "a b "},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			require.Equal(t, test.expect, ignoreLeadingComment(test.value))
		})
	}

}

// TestCsvData checks data that are writen in an annotated CSV file
func TestCsvData(t *testing.T) {
	var tests = []struct {
		name  string
		csv   string
		lines []string
	}{
		{
			"simple1",
			"_measurement,a,b\ncpu,1,1",
			[]string{"cpu a=1,b=1"},
		},
		{
			"simple1b",
			"_measurement,,a,b\ncpu,whatever,1,1",
			[]string{"cpu a=1,b=1"},
		},
		{
			"simple2",
			"_measurement\ncpu,1,1",
			[]string{""}, // no fields present
		},
		{
			"simple3",
			"_time\n1,1",
			[]string{""}, // no measurement present
		},
		{
			"annotated1",
			"#linetype measurement,,\nmeasurement,a,b\ncpu,1,2",
			[]string{"cpu a=1,b=2"},
		},
		{
			"annotated2",
			"#linetype measurement,tag,field\nmeasurement,a,b\ncpu,1,2",
			[]string{"cpu,a=1 b=2"},
		},
		{
			"annotated3",
			"#linetype measurement,tag,time,field\nmeasurement,a,b,time\ncpu,1,2,3",
			[]string{"cpu,a=1 time=3 2"},
		},
		{
			"annotated4",
			"#linetype measurement,tag,ignore,field\nmeasurement,a,b,time\ncpu,1,2,3",
			[]string{"cpu,a=1 time=3"},
		},
		{
			"annotated5",
			"#linetype measurement,tag,ignore,field\nmeasurement,a,b,time\ncpu,1,2,3",
			[]string{"cpu,a=1 time=3"},
		},
		{
			"annotated5",
			"#linetype measurement,tag,ignore,field\n" +
				"#linetypea tag,tag,\n" + // this must be ignored since it not a control comment
				"measurement,a,b,time\ncpu,1,2,3",
			[]string{"cpu,a=1 time=3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := readCsv(t, test.csv)
			table := Table{}
			var lines []string
			var errors []error
			for _, row := range rows {
				rowProcessed := table.AddRow(row)
				if rowProcessed {
					line, err := table.CreateLine(row)
					lines = append(lines, line)
					errors = append(errors, err)
				}
			}
			require.Equal(t, test.lines, lines)
		})
	}
}
