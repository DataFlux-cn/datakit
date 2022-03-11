# Datakit Tracing Data Flow

> Third Party Tracing Agent --> Datakit Frontend --> Datakit Backend --> Data Center

## Third Party Tracing Agent

目前 Datakit 支持的第三方 Tracing 数据包括：

- DDTrace
- Apache Jaeger
- OpenTelemetry
- Skywalking
- Zipkin

## Datakit Tracing Frontend

Datakit Frontend 即 Tracing Agent 负责接收并转换第三方 Tracing 数据结构，Datakit 内部使用 [DatakitSpan](datakit-tracing-struct) 数据结构。

Datakit Frontend 会解析接收到的 Tracing 数据并转换成 [DatakitSpan](datakit-tracing-struct) 然后发送到 [Datakit Tracing Backend](datakit-tracing#e7b3d9f2)。

Datakit Frontend 还负责配置 [Datakit Tracing Backend](datakit-tracing#e7b3d9f2) 的运算单元。

## Tracing 通用配置

```toml
## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
## that want to send to data center. Those keys set by client code will take precedence over
## keys in [tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
customer_tags = ["key1", "key2", ...]

## Keep rare tracing resources list switch.
## If some resources are rare enough(not presend in 1 hour), those resource will always send
## to data center and do not consider samplers and filters.
keep_rare_resource = false

## Ignore tracing resources map like service:[resources...].
## The service name is the full service name in current application.
## The resource list is regular expressions uses to block resource names.
[close_resource]
  service1 = ["resource1", "resource2", ...]
  service2 = ["resource1", "resource2", ...]

## Sampler config uses to set global sampling strategy.
## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
##   -1: always reject any tracing data send to datakit
##    0: accept tracing data and calculate with sampling_rate
##    1: always send to data center and do not consider sampling_rate
## sampling_rate used to set global sampling rate
[sampler]
  priority = 0
  sampling_rate = 1.0
```

## Datakit Tracing Backend

Datakit Tracing Backend 包括几个部分 Tracing <!--Statistics,--> Filters 和 Samplers：

<!-- - Tracing Statistics: 统计 Tracing 链路上的业务状态，例如：访问耗时，错误率等。 -->
- Filters:
  - keep_rare_resource: 当系统监测到某些链路在一小时之内没有发送任何 Tracing 数据那么将被认定为稀有并被透穿到 Data Center。
  - close_resource: 按照正则规则关闭某些 Service 下的一个或多个 Resource。
- Samplers: 基于概率的 Tracing 数据采样。多服务环境下采样率必须配置一致才能达到采样效果，
  - 例一：A-Service(0.3) --> B-Service(0.3) --> C-Service(0.3) 配置正确，最终采样率为 30%。
  - 例二：A-Service(0.1) --> B-Service(0.3) --> C-Service(0.1) 配置错误，链路不能正常工作。