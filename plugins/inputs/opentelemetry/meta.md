## json 数据
```
resource_spans:{
resource:{attributes:{key:"message.type"  value:{string_value:"message-name"}} 
 attributes:{key:"service.name"  value:{string_value:"testservice"}}}  

instrumentation_library_spans:{
instrumentation_library:{name:"test-tracer"}  
spans:{
    trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
    span_id:".\xbd\x06c\x10ɫ*"  
    parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
    name:"Sample-0" 
    kind:SPAN_KIND_INTERNAL  
    start_time_unix_nano:1644312397453313100 
    end_time_unix_nano:1644312398464865900  
    status:{}
} 

 spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"\xd0\xf3\xe0\t\x92\xea-\xcc"  parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-1"  
\kind:SPAN_KIND_INTERNAL  start_time_unix_nano:1644312398464865900  
end_time_unix_nano:1644312399469852400 
 status:{}} 
 
spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"\x7f<\x13\xc7L\x18\xdc\xda"  
parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-2"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312399469938800 
 end_time_unix_nano:1644312400478394400  status:{}}
  
spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"^}\x1c\x18\x7f9>\x8e"  
parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-3"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312400478485700  
end_time_unix_nano:1644312401488931400  
status:{}}}}
```

## 原生 span
```
trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:".\xbd\x06c\x10ɫ*" 
 parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-0"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312397453313100  
end_time_unix_nano:1644312398464865900 
 status:{}
```

## 组装到 dkspan
```
{TraceID:943cdf007a7882e75779fe93ab199561 
ParentID:a72a805a23be4cf6 
SpanID:2ebd066310c9ab2a 
Service:testservice 
Resource: 
Operation:Sample-0 
Source:opentelemetry 
SpanType: 
SourceType: 
Env: 
Project: 
Version: 
Tags:map[messagetype:string_value:"message-name" servicename:string_value:"testservice"] 
EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: 
PID: 
Start:1644312397453313100 
Duration:1011552800 
Status:STATUS_CODE_UNSET 
Content: 
SampleRate:0}
```

### 测试 

``` text
&{TraceID:00000000000000000000000000000001 ParentID:0 SpanID:0000000000000002 Service:global.ServerName Resource:test-server Operation:span_name Source:opentelemetry SpanType:SPAN_KIND_UNSPECIFIED SourceType: Env: Project: Version: Tags:map[a:b int:123] EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: PID: Start:1645424025258951800 Duration:1000000000 Status:info Content:{"trace_id":"AAAAAAAAAAAAAAAAAAAAAQ==","span_id":"AAAAAAAAAAI=","name":"span_name","start_time_unix_nano":1645424025258951800,"end_time_unix_nano":1645424026258951800,"attributes":[{"key":"a","value":{"Value":{"StringValue":"b"}}},{"key":"int","value":{"Value":{"IntValue":123}}}],"status":{}} Priority:0 SamplingRateGlobal:0},
&{TraceID:00000000000000000000000000000001 ParentID:0 SpanID:0000000000000002 Service:global.ServerName Resource:test-server Operation:span_name Source:opentelemetry SpanType:SPAN_KIND_UNSPECIFIED SourceType: Env: Project: Version: Tags:map[a:b int:123] EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: PID: Start:1645424025258951800 Duration:1000000000 Status:info Content:{"trace_id":"AAAAAAAAAAAAAAAAAAAAAQ==","span_id":"AAAAAAAAAAI=","name":"span_name","start_time_unix_nano":1645423573257862600,"end_time_unix_nano":1645423574257862600,"attributes":[{"key":"a","value":{"Value":{"StringValue":"b"}}},{"key":"int","value":{"Value":{"IntValue":123}}}],"status":{}} Priority:0 SamplingRateGlobal:0}
   
```


          
## metric 原始数据
```
resource_metrics:{
resource:{

attributes:{key:"service.name" value:{string_value:"unknown_service:___go_build_go_opentelemetry_io_otel_example_otel_collector.exe"}}  -- service name
attributes:{key:"telemetry.sdk.language" value:{string_value:"go"}} 
attributes:{key:"telemetry.sdk.name" value:{string_value:"opentelemetry"}} 
attributes:{key:"telemetry.sdk.version" value:{string_value:"1.3.0"}}} 

instrumentation_library_metrics:{instrumentation_library:{name:"test-meter"}  -- metric name
 
metrics:
name:"an_important_metric"
 description:"Measures the cumulative epicness of the app" 
sum:{
    data_points:{
        start_time_unix_nano:1644481279198123400 time_unix_nano:1644481281107922800 as_double:10
    } 
    aggregation_temporality:AGGREGATION_TEMPORALITY_CUMULATIVE 
    is_monotonic:true
    }
}}
 
schema_url:"https://opentelemetry.io/schemas/v1.7.0"}

pl:
name : test-meter
tags : [attributes]
fileds : 
t : time now
-------------------------------------------
resource_metrics:{
resource:{
attributes:{key:"host.name" value:{string_value:"songlongqi"}} 
attributes:{key:"os.description" value:{string_value:"Windows 10 Home China 21H2 (2009) [Version 10.0.19044.1466]"}} 
attributes:{key:"os.type" value:{string_value:"windows"}} 
attributes:{key:"process.command_args" value:{array_value:{values:{string_value:"C:\\Users\\18332\\AppData\\Local\\Temp\\___go_build_go_opentelemetry_io_otel_example_otel_collector.exe"}}}} 
attributes:{key:"process.executable.name" value:{string_value:"___go_build_go_opentelemetry_io_otel_example_otel_collector.exe"}} 
attributes:{key:"process.executable.path" value:{string_value:"C:\\Users\\18332\\AppData\\Local\\Temp\\___go_build_go_opentelemetry_io_otel_example_otel_collector.exe"}} 
attributes:{key:"process.owner" value:{string_value:"SONGLONGQI\\18332"}} 
attributes:{key:"process.pid" value:{int_value:20576}} 
attributes:{key:"process.runtime.description" value:{string_value:"go version go1.16.8 windows/amd64"}} 
attributes:{key:"process.runtime.name" value:{string_value:"gc"}} 
attributes:{key:"process.runtime.version" value:{string_value:"go1.16.8"}}} 

instrumentation_library_metrics:{instrumentation_library:{name:"test-meter"} 
metrics:{name:"an_important_metric" description:"Measures the cumulative epicness of the app" 
sum:{data_points:{start_time_unix_nano:1644482920517894200 time_unix_nano:1644482920517894200 as_double:10} aggregation_temporality:AGGREGATION_TEMPORALITY_CUMULATIVE is_monotonic:true}}} 
schema_url:"https://opentelemetry.io/schemas/v1.7.0"}
```




--------- 格式化 onemetric  json --------------------
```
        {
            "resource": {
                "attributes": [
                    {
                        "key": "host.name",
                        "value": {
                            "Value": {
                                "StringValue": "songlongqi"
                            }
                        }
                    },
                    {
                        "key": "os.description",
                        "value": {
                            "Value": {
                                "StringValue": "Windows 10 Home China 21H2 (2009) [Version 10.0.19044.1526]"
                            }
                        }
                    },
                    {
                        "key": "os.type",
                        "value": {
                            "Value": {
                                "StringValue": "windows"
                            }
                        }
                    },
                    {
                        "key": "process.pid",
                        "value": {
                            "Value": {
                                "IntValue": 24052
                            }
                        }
                    },
                    {
                        "key": "service.name",
                        "value": {
                            "Value": {
                                "StringValue": "serviceNameForMetric"
                            }
                        }
                    }
                ]
            },
            "instrumentation_library_metrics": [
                {
                    "instrumentation_library": {
                        "name": "test_meter"
                    },
                    "metrics": [
                        {
                            "name": "ex.com",
                            "Data": {
                                "Histogram": {
                                    "data_points": [
                                        {
                                            "attributes": [
                                                {
                                                    "key": "A",
                                                    "value": {
                                                        "Value": {
                                                            "StringValue": "1"
                                                        }
                                                    }
                                                }
                                            ],
                                            "start_time_unix_nano": 1646290961917362200,
                                            "time_unix_nano": 1646290961919645000,
                                            "count": 1,
                                            "sum": 12,
                                            "bucket_counts": [
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                0,
                                                1
                                            ],
                                            "explicit_bounds": [
                                                0.005,
                                                0.01,
                                                0.025,
                                                0.05,
                                                0.1,
                                                0.25,
                                                0.5,
                                                1,
                                                2.5,
                                                5,
                                                10
                                            ]
                                        }
                                    ],
                                    "aggregation_temporality": 2
                                }
                            }
                        },
                        {
                            "name": "an_important_metric",
                            "description": "Measures the cumulative epicness of the app",
                            "Data": {
                                "Sum": {
                                    "data_points": [
                                        {
                                            "start_time_unix_nano": 1646290961917362200,
                                            "time_unix_nano": 1646290961919645000,
                                            "Value": {
                                                "AsDouble": 10
                                            }
                                        }
                                    ],
                                    "aggregation_temporality": 2,
                                    "is_monotonic": true
                                }
                            }
                        }
                    ]
                }
            ],
            "schema_url": "https://opentelemetry.io/schemas/v1.7.0"
        }
    ]
```

metric 在dk上的映射结构体为
``` go
type otelResourceMetric struct {
	Operation   string            `json:"operation"`   // metric.name
	Source      string            `json:"source"`      // inputName ： opentelemetry
	Attributes  map[string]string `json:"attributes"`  // tags
	Resource    string            `json:"resource"`    // global.Meter name
	Description string            `json:"description"` // metric.Description
	StartTime   uint64            `json:"start_time"`  // start time
	UnitTime    uint64            `json:"unit_time"`   // end time

	ValueType string      `json:"value_type"` // double | int | histogram | ExponentialHistogram | summary
	Value     interface{} `json:"value"`      // 5种类型 对应的值：int | float

	Content string `json:"content"` //

	// TODO : Exemplar 可获取 spanid 等
}
```        
 