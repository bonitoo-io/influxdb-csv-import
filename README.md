# influxdb-csv-import
CSV data written to influx


## DONE: STEP 1 - merged to influxdata/influxdb
https://github.com/influxdata/influxdb/issues/17003 introduces a new CSV format for existing _influx write_ command.  CSV data on input are transformed to line protocol with the help of CSV annotations.
### CSV Annotations
* https://v2.docs.influxdata.com/v2.0/reference/syntax/annotated-csv/#annotations
   * all of them are supported
* additionally
   * *#datatype* annotation is enhanced with:
      * to mark non-field data: 
          * _measurement_, _tag_, _time_
          * _ignore_ simply ignores the column
          * _time_ is also supported, it is an alias for the existing _dateTime_
          * _dateTime_ is either as a long (int64) number or RFC3339 format, this type is always used to represent time of measurement; or you can specify
             * _dateTime:RFC3339_ for RFC3339 time
             * _dateTime:number_ to expect a long number
      * a _field_ data type can be used to let detect field's data type
      * default datatype for a column is =field= unless _field column is present (ignored then)
      * there can be at most 1 _dateTime_ column
   * the following columns will have an extra semantics: _measurement, _time, _field, _value
      * _measurement:  name of measurement
      * _time: time of measurement
      * _field: column that contains field name
      * _value: column that contains field value
* new `influx write` flags
```sh
  -f, --file string        The path to the file to import
      --format string      Input format, either lp (Line Protocol) or csv (Comma Separated Values). Defaults to lp unless '.csv' extension
```  
### DRY RUN
Run "write dryrun" command to write lines to stdout instead of InfluxDB. This "dryrun" command helps with validation and troubleshooting of CSV data.

### Example 1 - Flux Query Result
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
### Example 2 - Simple Annotated CSV file
*influx write dryrun --file doc/examples/annotatedLinepart.csv*

```bash
#datatype measurement,tag,tag,double,double,ignored,dateTime:number
m,cpu,host,time_steal,usage_user,nothing,time
cpu,cpu1,rsavage.prod,0,2.7,a,1482669077000000000
cpu,cpu1,rsavage.prod,0,2.2,b,1482669087000000000
```

line protocol data: 
```
cpu,cpu=cpu1,host=rsavage.prod time_steal=0,usage_user=2.7 1482669077000000000
cpu,cpu=cpu1,host=rsavage.prod time_steal=0,usage_user=2.2 1482669087000000000
```

### Example 3 - Annotated CSV file with Data Types
*influx write dryrun --file doc/examples/annotatedDatatype.csv*

```bash
#datatype measurement,tag,string,double,boolean,long,unsignedLong,duration,dateTime
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

## TODO: STEP 2
Further set of enhancements that helps to process CSV files without actually changing them:
   
- `--header` option in `influx write` command let your add annotation or header rows without changing the data on input (supplied via `--file` or stdin)
   - you can supply more `--header` options, the rows will be prepended in the order as they appear on command line
- `#constant` annotation adds a constant column to the data
   - the format of a constant annotation row is `#constant,datatype,name,value`, so you have to specify a supported datatype, column name and a constant value
   - `column` can be omitted for _dateTime_ or _measurement_ columns, so the annotation can be simply `#constant,measurement,cpu`
   - note that you can add constant annotations to existing data using `--header` options of `influx write` cli
- a measurement column can be of `dateTime:format` datatype to use a custom _format_ to parse column values
   - the format layout is described in https://golang.org/pkg/time layout, for example `dateTime:2006-01-02` parses 4-digit-year , '-' , 2-digit month , 2 digit day of month
   - `dateTime:RFC3339`, `dateTime:RFC3339Nano` and `dateTime:number` are predefined formats
      - _RFC3339 format_ is 2006-01-02T15:04:05Z07:00
      - _RFC3339Nano_ format is 2006-01-02T15:04:05.999999999Z07:00
      - _number_ represent UTCs time since epoch in nanoseconds
- a _double_, _long_ or _unsignedLong_ field column can also have `format` 
   - the `format` is a single character that is used to separate integer and fractional part (usually `.` or `,`) of the number followed by additional characters that are ignored (such as as `, _`), these characters are ussually used to separate large numbers into groups
   - for example `double:,.` is a double data type that parses number from Spanish locale, where numbers look like `3.494.826.157,123`
   - for _long_ or _unsignedLong_ types, everything after and including a fraction character is ignored, for example `double:,.` will parse `3.494.826.157,123` as `3.494.826157`
   - note that you have to escape column delimiters when they appear in a column value, for example
      - `#constant,"double:,.",myColumn,"1,234.011"`
- a CSV file can start with a line `sep=;` to inform about a character that is used to separate columns in data rows, by default `,` is used as a column separator
- `--debug` and `--logCsvErrors` options helps with troubleshooting of CSV conversions
   - `--debug` prints to stderr debugging information about columns that are used to create protocol lines out of csv data rows
   - `--logCsvErrors` prints CSV data errors to stderr and continue with CSV processing
      - any error stops the processing without this option specified
