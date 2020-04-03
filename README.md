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
*influx write dryrun --file doc/examples/annotatedSimple.csv*

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

## STEP 2
These examples are related to https://github.com/influxdata/influxdb/issues/17004

### Modified: Example 2 - Simple CSV file
*influx write dryrun --file doc/examples/annotatedSimple.csv*

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

Data type can be supplied in the column name, the CSV can be shortened to:

```
m|measurement,cpu|tag,host|tag,time_steal|double,usage_user|double,nothing|ignored,time|dateTime:number
cpu,cpu1,rsavage.prod,0,2.7,a,1482669077000000000
cpu,cpu1,rsavage.prod,0,2.2,b,1482669087000000000
```
*influx write dryrun --file doc/examples/labelsWithDataTypes_labels.csv*

### Modified: Example 3 - Data Types with default values
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

Default value can be supplied in the column label after data type, the CSV could be also:

```
m|measurement|test,name|tag|annotatedDatatypes,s|string,d|double,b|boolean,l|long,ul|unsignedLong,dur|duration,time|dateTime
,,str1,1.0,true,1,1,1ms,1
,,str2,2.0,false,2,2,2us,2020-01-11T10:10:10Z
```
*influx write dryrun --file doc/examples/annotatedDatatype_labels.csv*

### Example 4 - Advanced usage
*influx write dryrun --file doc/examples/datetypeFormats_labels.csv*

```
#constant measurement,test
#constant tag,name,datetypeFormats
#timezone -0500
t|dateTime:2006-01-02|1970-01-02,"d|double:,. ","b|boolean:y,Y:n,N|y"
1970-01-01,"123.456,78",
,"123 456,78",Y
```
   - measurement and extra tags is defined using the `#constant` annotation
   - timezone for dateTime is to `-0500` (EST)
   - `t` column is of `dateTime` data type of format is `2006-01-02`, default value is _January 2nd 1970_
   - `d` column is of `double` data type with `,` as a fraction delimiter and `. ` as ignored separators that  used to visually separate large numbers into groups
   - `b` column os of `boolean` data type that considers `y` or `Y` truthy, `n` or `N` falsy and empty column values as truthy 


line protocol data:
```
test,name=datetypeFormats d=123456.78,b=true 18000000000000
test,name=datetypeFormats d=123456.78,b=true 104400000000000
```

You can prepend and/or remove first lines from input data using command line options. You can also 
override column name and define extra annotations that drive data processing. For example:

*influx write dryrun --skipHeader=4 --header "#constant measurement,test2" --header "t|dateTime:2006-01-02,_|ignored,s|string|unknown" --file doc/examples/datetypeFormats_labels.csv*
    - removes the first 4 lines from input and prepends the following that define
       - measurement of the data
       - column header row with 
          - custom dateTime format without a default value
          - string column with a default value

line protocol data:
```
test2 s="unknown" 0
test2 s="Y"
```
### Example 5 - Custom column separator
*influx write dryrun --file doc/examples/columnSeparator.csv*
```
sep=;
m|measurement;available|boolean:y,Y:|n;dt|dateTime:number
test;nil;1
test;N;2
test;";";3
test;;4
test;Y;5
```
   - the first line can define a column separator character for next lines, here: `;`
   - other lines use this separator, `available|boolean:y,Y` does not need to be wrapped in double quotes

line protocol data:
```
test available=false 1
test available=false 2
test available=false 3
test available=false 4
test available=true 5
```
### Example 6 - CSV conversion troubleshooting
`influx write dryrun` helps with troubleshooting together with the following options
   - `--debug` 
      - prints out internal representation of command line arguments including the default values for the flags that drive CSV processing
      - prints out metadata about CSV columns that used for processing
   - `--skipRowOnError`
      - if a row cannot be parsed, it is ignored and the parsing error is printed to to stderr
      - any error stop the processing without this option present

Having csv input:
```
m|measurement,usage_user|double
cpu,2.7
cpu,nil
cpu,
,2.9
```

*influx write dryrun --skipRowOnError --file doc/examples/troubleshooting.csv 2>/dev/null*
```
cpu usage_user=2.7
```

*influx write dryrun --skipRowOnError --file doc/examples/troubleshooting.csv 1>/dev/null*
```
2020/04/03 15:14:51 line 3: column 'usage_user': strconv.ParseFloat: parsing "nil": invalid syntax
2020/04/03 15:14:51 line 4: no field data found
2020/04/03 15:14:51 line 5: column 'm': no measurement supplied
```
