# DataKit

## 安装手册

参见[这里](https://gitlab.jiagouyun.com/zy-docs/pd-forethought-helps/blob/dev/03-%E6%95%B0%E6%8D%AE%E9%87%87%E9%9B%86/02-datakit%E9%87%87%E9%9B%86%E5%99%A8/index.md)

## 编译

### 选择不同的编译输出

```
$ make test     # 编译测试环境
$ make pub_test # 发布 datakit 到测试环境

$ make release  # 编译线上发布版本
$ make pub_test # 发布 datakit 到线上环境

# 将 datakit 以镜像方式发布到 https://registry.jiagouyun.com
# 注意：registry.jiagouyun.com 需要一定的权限才能发布镜像
$ make pub_image

$ make agent # 编译不同平台的 telegraf 到 embed 目录
```

#### telegraf 单独发布

为便于 CI 集成，CI 中移除了 telegraf 的编译、打包、发布流程，故 telegraf 的发布还是人工形式：

```
$ make agent           # 编译各个平台的 telegraf
$ make pub_agent       # 将 telegraf 发布到各个环境（测试、预发、生产）
```

因 telegraf 不常更新，每次 datakit 发布无需额外发布 telegraf。**如果新版的 telegraf 有采集器被集成到 datakit 中**，则需要重新打包、发布一次 telegraf

## datakit 使用示例

列举当前 datakit 支持的采集器列表，可 `grep` 输出，采集器带 `[d]` 前缀的为 datakit 采集器，带 `[t]` 为 telegraf 采集器

```
# 查看阿里云采集器列表
$ ./datakit -tree | grep aliyun
aliyun
  |--[d] aliyunddos
  |--[d] aliyunsecurity
  |--[d] aliyunlog
  |--[d] aliyuncms
  |--[d] aliyuncdn
  |--[d] aliyuncost
  |--[d] aliyunprice
  |--[d] aliyunrdsslowLog
  |--[d] aliyunactiontrail

# 查看CPU集器列表
$ ./datakit -tree | grep cpu
cpu
  |--[t] cpu

# 查看所有采集器列表，末尾会列出采集器个数统计
$ ./datakit -tree 
...
===================================
total: 95, datakit: 33, agent: 62
```

## 采集器开发

### 约束

采集器开发遵循如下几个约束

- 采集器目前分为三类：
	- 集成在 datakit 中的采集器，它们代码位于 `plugins/inputs/` 目录下
	- telegraf 采集器，telegraf 进程和 datakit 分离运行，由 datakit 启动
	- 外部采集器，它们和 datakit 主进程分离运行，但是由 datakit 来启动。它们代码位于 `plugins/externals/` 目录下。
		- 注意：外部采集器的数据，均以 gRPC 的方式发送给 datakit

- 所有采集器示例配置模板（示例模板中**不要带有中文字符**，在 Windows 下可能出现乱码，不便于用户编辑）

```
# 采集器名称可用小驼峰或连写（如 oraclemonitor 或 oracleMonitor），不建议使用其它分隔字符（如 oracle_monitor 或 oracle-monitor）

#[inputs.xxx]     # 此处也可以是 [[inputs.xxx]] 这种形式，即支持批量配置，此处的 xxx 是采集器名称
#key1 = "val1"
#key2 = 123
#key3 = false
#someOtherKey = "key-value"   # 建议用小驼峰或下划线分割（some_other_key）的方式来命名字段
#...
#
#[inputs.xxx.tags] # 以此类推，此处也可以是 [[inputs.xxx.tags]]
#	tag1 = "val1"
#	tag2 = "val2"
#	...

#[inputs.xxx.tags]
#	ip = "1.2.3.4"          # 对一些专业缩写，可用全大写或全小写(ip 或 IP)，但不用 Ip 这种
# CIDR = "192.168.1.0/24" # 此处 CIDR 和 cidr 都可以
# host = "dataflux.cn"    # 对主机命名，可用 HOST/host 或 ip/IP
# interval = "1s"         # 所有时间单位，统一用 go 中 time.ParseDuration() 可接受的字符串形式，如 300ms, -1.5h, 2h45m 等
#	...

#[inputs.xxx.tags]
#	someFiled = "xxx"       # required：对于一些必须配置的字段，必须在 config-sample 中标记其为 required
#	someOtherFiled = "yyy"  # 未标记 required 的字段，默认为 optional

#[inputs.xxx.tags]
#	some_filed = "xxx"  # 老的配置字段
#	someFiled = "xxx"   # 为了兼容老的配置字段，在代码中，应该定义多个同义字段，不能直接删除老的字段，这会导致老的配置解析出错
######## 示例 #########
// 原有对象定义
type Obj struct {
	SomeField string `toml:"some_field"` 
	...
}

// 新对象定义
type Obj struct {

	// 因不符合命名规范，新版本更新了其 tag 标签
	SomeField_DEPRECATED string `toml:"some_field"` // 这个字段得留着，解析到老的配置后，手动丢给 SomeField
	SomeField string `toml:"someField"`
	...
}
#################
```

- 采集器采集到的数据，tag 来源有三种：
	- 用户在具体采集器中配置了 tags，如上面 `[[inputs.xxx.tags]]` 所示
	- 数据源中本来就可以抽取一些字段作为行协议的 tag
	- datakit 主配置文件中，配置了 `global_tags`

在构造行协议时，这些 tags 的覆盖优先级逐次降低，假定数据源中带有 `host=abc` 这个字段，采集器将其作为 tag 加入到了行协议中，如果用户在采集器配置中也加了 `host=abc123`，那么源数据中 `host` 被覆盖成  `abc123`。如果 `global_tags` 中也配置了 `host=xyz`，此时 `host` 值维持 `abc123` 不变。

假定数据源中没有 `host` 这个 tag，用户也没在采集器上配置 `host`，那么行协议中的 `host` 值为 `xyz`。

- 对于有动态库依赖的采集器，或者其它语言开发的采集器，应该将代码放在 `plugins/externals` 目录下，并且在 `cmd/make/make.go` 中确定对应的编译/打包设定。
