package cmd

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const data = `
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:17:57Z,0,time_steal,cpu,cpu1,rsavage.prod
,,0,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:07Z,0,time_steal,cpu,cpu1,rsavage.prod

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,1,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:01Z,2.7263631815907954,usage_user,cpu,cpu-total,tahoecity.prod
,,1,2020-02-25T22:17:54.068926364Z,2020-02-25T22:22:54.068926364Z,2020-02-25T22:18:11Z,2.247752247752248,usage_user,cpu,cpu-total,tahoecity.prod
#unre
`

// TestTableColumns validates construction of table columns
func TestTableColumns(t *testing.T) {
	table := Table{}
	rows, err := csv.NewReader(strings.NewReader(data)).ReadAll()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	for i, row := range rows {
		rowProcessed := table.AddRow(row)
		if i%6 < 4 {
			require.Equal(t, rowProcessed, false)
		} else {
			require.Equal(t, rowProcessed, true)
			if i%6 == 4 {
				require.Equal(t, len(table.columns), 10)
				for j, col := range table.columns {
					require.Equal(t, col.Index, j+1)
					require.Equal(t, col.Label, rows[i-1][j+1])
					require.Equal(t, col.DefaultValue, rows[i-2][j+1])
					require.Equal(t, col.DataType, rows[i-3][j+1])
					require.Equal(t, col.Group, rows[i-4][j+1] == "true")
				}
			}
		}
	}
}
