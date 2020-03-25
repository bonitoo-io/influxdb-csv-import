# influxdb-csv-import
CSV data written to influx

https://github.com/influxdata/influxdb/issues/17003 introduces a new CSV format for existing _influx write_ command.  CSV data on input are transformed to line protocol with the help of CSV annotations.

## CSV Annotations

* https://v2.docs.influxdata.com/v2.0/reference/syntax/annotated-csv/#annotations
   * all of them are supported
* additionally
   * column names that start with _ are OOTB ignored, unless: _measurement, _time, _field, _value
      * _measurement:  measurement part
      * _time: timestamp part
      * _field: column that contains field name
      * _value: column that contains field value
   * *#linetype* annotation associated a particular csv column protocol with a line part
      * supported values are: _measurement_, _tag_, _time_, _ignore(d)_
      * default is =field= unless _field column is present (ignored then)
   * time column can be specified as an int64 number or in RFC3339 format

## DRY RUN
Run "write dryrun" command to write lines to stdout instead of InfluxDB. This "dryrun" command helps with validation and troubleshooting of CSV data.

## Example 1 - Flux Query Result
*influx write dryrun --file doc/examples/fluxQueryResult.csv*

```bash
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
```
line protocol data:
```
cpu,cpu=cpu1,host=rsavage.prod time_steal=0 1582669077000000000
cpu,cpu=cpu1,host=rsavage.prod time_steal=0 1582669087000000000
cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.7263631815907954 1582669081000000000
cpu,cpu=cpu-total,host=tahoecity.prod usage_user=2.247752247752248 1582669091000000000
```
## Example 2 - Simple Annotated CSV file
*influx write dryrun --file doc/examples/annotatedLinepart.csv*

```bash
#linepart measurement,tag,tag,field,field,ignored,time
m,cpu,host,time_steal,usage_user,nothing,time
cpu,cpu1,rsavage.prod,0,2.7,a,1482669077000000000
cpu,cpu1,rsavage.prod,0,2.2,b,1482669087000000000
```

line protocol data: 
```
cpu,cpu=cpu1,host=rsavage.prod time_steal=0,usage_user=2.7 1482669077000000000
cpu,cpu=cpu1,host=rsavage.prod time_steal=0,usage_user=2.2 1482669087000000000
```
Note that all fields are of type double.

## Example 3 - Annotated CSV file with Data Types
*influx write dryrun --file doc/examples/annotatedDatatype.csv*

```bash
#datatype ,,string,double,boolean,long,unsignedLong,duration,
#linepart measurement,tag,,,,,,,time
#default test,annotatedDatatypes,,,,,,
m,name,s,d,b,l,ul,dur,time
,,str1,1.0,true,1,1,1ms,1
,,str2,2.0,false,2,2,2us,2020-01-11T10:10:10Z
```

line protocol data: 
```
test,name=annotatedDatatypes s="str1",d=1,b=true,l=1i,ul=1u,dur=1000000i 1
test,name=annotatedDatatypes s="str2",d=2,b=false,l=2i,ul=2u,dur=2000i 1578737410000000000
```
