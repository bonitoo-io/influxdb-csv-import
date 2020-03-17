package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const csvData = `
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
#unre
`

var lineProtocol = []string{
	"cpu,cpu=cpu1,host=rsavage.prod time_steal=0 2020-02-25T22:17:57Z",
	"cpu,cpu=cpu1,host=rsavage.prod time_steal=0 2020-02-25T22:18:07Z",
	"cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.7263631815907954 2020-02-25T22:18:01Z",
	"cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.247752247752248 2020-02-25T22:18:11Z",
}

// TestTableColumns validates construction of table columns
func TestTableColumns(t *testing.T) {
	table := Table{}
	reader := csv.NewReader(strings.NewReader(csvData))
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
	lineProtocolIndex := 0
	for i, row := range rows {
		rowProcessed := table.AddRow(row)
		if i%6 < 4 {
			require.Equal(t, rowProcessed, false, "row %d", i)
		} else {
			require.Equal(t, rowProcessed, true, "row %d", i)
			require.Equal(t, fmt.Sprint(table.CreateMetric(row)), lineProtocol[lineProtocolIndex])
			lineProtocolIndex++
			if i%6 == 4 {
				// verify table
				require.Equal(t, len(table.columns), 10)
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
				require.NotNil(t, table.cachedMeasurement)
				require.NotNil(t, table.cachedFieldName)
				require.NotNil(t, table.cachedFieldValue)
				require.NotNil(t, table.cachedTime)
				require.NotNil(t, table.cachedTags)
				require.Equal(t, table.cachedMeasurement.Label, "_measurement")
				require.Equal(t, table.cachedFieldName.Label, "_field")
				require.Equal(t, table.cachedFieldValue.Label, "_value")
				require.Equal(t, table.cachedTime.Label, "_time")
				require.Equal(t, len(table.cachedTags), 2)
				require.Equal(t, table.cachedTags[0].Label, "cpu")
				require.Equal(t, table.cachedTags[1].Label, "host")
			}
		}
	}
}
