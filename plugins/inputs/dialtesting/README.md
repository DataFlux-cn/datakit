# 网络拨测功能定义

全局定义：

- 所有拨测产生的数据，都以行协议方式，通过 `/v1/write/metric` 接口存为指标数据

## HTTP 拨测任务定义

```python
{
	"id": "dialt_xxxxxxxxxxxxxxxxx", # 拨测任务 ID

	"url": "http://example.com/some/api", 
	"method": "POST",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway。这里的 token
  # 对 Dataflux SAAS/PASSS 而言，实际上隐含了工作空间信息
	"post_url": "https://dataway.cn?token=tkn_xxx",

	# 用于更新任务时，停止执行
	"stop": false, 

	"name": "give your test a name",
	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week

  ###############
	# 区域
  ###############

	# 对 push 端而言，可多选
	"regions": ["hangzhou", "beijing", "chengdu"],

	# 对 pull 端而言，只能单选
	"region": "hangzhou"
	
	"advance_options": {
		"request_options": {
			"follow_redirect": true,
			"headers": {
				"header1": "value1",
				"header2": "value2"
			},
			"cookies": "",
			"auth": {
				"username": "",
				"password": ""
			}
		},
		"request_body": { # 以下几个类型只能单选
			"body_type": "text/plain|application/json|text/xml|None",
			"body": ""
		},
		"certificate": {
			"ignore_server_certificate_error": false,
			"private_key": "",
			"certificate": ""
		},
		"proxy": {
			"url": "",
			"headers": {
				"header1": "value1"
			}
		}
	},

	"success_when":  [ 
		{

			# body|header|response_time|status_code 都是单个判定条件，它们各自出现的时候，表示 AND 的关系

			"body":{},
			"header": {
				"header-name":{ # 以下几个条件只能单选
					"contains":"",
					"does_not_contains": "",
					"is": "",
					"is_not": "",
					"match_regex": "",
					"does_not_match_regex": ""
				},
				"another-header-name": "...",
			},
			"response_time": {
				"less_than": "100ms"
			},
			"status_code": { # 以下几个条件只能单选
				"is": 200,
				"is_not": 400,
				"match_regex": "ok*",
				"not_match_regex": "*bad"
			}
		},
		
		{
			"AND_another_assert": "..."
		}
	]
}
```

### HTTP 拨测行协议定义

```python
{
	"measurement": "http_dial_testing",
	"tags": {
		"name": "",
		"url": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 region，故这里只有单个值
		"region": "",

		"result": "OK", # 只有 OK/FAIL 两种状态，便于分组以及过滤查找

		"status_code": "200/OK" # 便于分组以及过滤

		# HTTP 协议版本，HTTP/1.0 HTTP/1.1 等
		"proto": "HTTP/1.0"
	},

	"fields": {
		# 如失败，如实描述。如成功，无此指标
		"fail_reason": "字符串描述失败原因"

		# HTTP 相应时间, 单位 ms
		"response_time": 300,

		# 返回 body 长度，单位字节。如无 body，则无此指标，或者填 0
		"content_length": 1024,

		# 只有 1/-1 两种状态, 1 表示成功, -1 表示失败, 便于 UI 绘制状态跃迁图（TODO: 现在前端图标支持负数么）
		"success": 1, 

		# 其它指标可再增加...
	},
	"time": time.Now()
}
```

注意，这里的 `fail_reason` 要描述 `body/header/response_time/status_code` 各自失败的原因。如果可以，所有原因都描述一遍，如 `response time larger than 100ms; status code match regex 4*`

## TCP 拨测任务定义

```python
{
	"id": "dialt_xxxxxxxxxxxxxxxxx", # 拨测任务 ID
	"host": "www.example.com",
	"port": "443",
	"name": "give your test a name",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway
	"post_url": "https://dataway.cn?token=tkn_xxx",

	# 用于更新任务时，停止执行
	"stop": false, 

	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week

  ###############
	# 区域
  ###############

	# 对 push 端而言，可多选
	"regions": ["hangzhou", "beijing", "chengdu"],

	# 对 pull 端而言，只能单选
	"region": "hangzhou"

	"success_when":  [
		{
			"response_time": {
				"less_than": "100ms"
			}
		}
	]
}
```

### TCP 拨测行协议定义

```python
{
	"measurement": "tcp_dial_testing",
	"tags": {
		"name": "",
		"host": "",
		"port": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 region，这里只有单个值
		"region": "",

		"result": "OK", # 只有 OK/FAIL 两种状态，便于分组以及过滤查找
	},

	"fields": {

		# 如失败，如实描述。如成功，无此指标
		"fail_reason": "字符串描述失败原因"

		# TCP 连接建立时间, 单位 ms
		"dial_time": 30,

		# 域名解析时间，单位 ms
		"resolve_time": 30,

		# 只有 1/-1 两种状态, 1 表示成功, -1 表示失败
		"success": 1,

		# 其它指标可再增加...
	}

	"time": time.Now()
}
```

## DNS 拨测任务定义

```python
{
	"id": "dialt_xxxxxxxxxxxxxxxxx", # 拨测任务 ID
	"domain": "www.example.com",
	"dns_server": "",
	"name": "give your test a name",

	# 拨测数据的存放地址，对 SAAS 而言，是 openway.dataflux.cn
	# 对 PAAS 而言，需要一个单独的公网可访问的 Dataway
	"post_url": "https://dataway.cn?token=tkn_xxx",

	# 用于更新任务时，停止执行
	"stop": false, 

	"tags": {
		"tag1": "val1",
		"tag2": "val2"
	},

	"frequency": "1m",   # 1min ~ 1 week
	"regions": ["hangzhou", "beijing", "chengdu"],

	"success_when":  [
		{
			"response_time": {
				"less_than": "100ms"
			},
			"at_least_one_record": {
				"of_type_a": {
					"is": "",
					"contains": "",
					"match_regex": "",
					"not_match_regex": ""
				},
				"of_type_aaaa": {},
				"of_type_cname": {},
				"of_type_mx": {},
				"of_type_txt": {}
			},
			"every_record": {}
		},
		{
			"AND_another_assert": "..."
		}
	]
}
```

关于 DNS 的各种 [`of_type_xxx`](https://support.dnsimple.com/categories/dns/)

### DNS 拨测行协议定义

```python
{
	"measurement": "dns_dial_testing",
	"tags": {
		"name": "",
		"domain": "",
		"dns_server": "",
		"用户额外指定的各种": "tags",

		# 每个具体的 datakit 只会在一个 region，故这里只有单个值
		"region": "", 
		"result": "OK", # 只有 OK/FAIL 两种状态，便于分组以及过滤查找
	},

	"fields": {

		# 如失败，如实描述。如成功，无此指标
		"fail_reason": "字符串描述失败原因"

		# DNS 响应时间, 单位 ms
		"response_time": 30,

		# 只有 1/-1 两种状态, 1 表示成功, -1 表示失败
		"success": 1,

		# 其它指标可再增加...
	},
	"time": time.Now()
}
```

## 架构设计

### 整体架构设计

![拨测任务整体架构](net-dial-testing-arch.png)

术语定义：

- dialtesting：拨测服务的中心服务器，它提供一组 HTTP 接口，供授信的第三方推送拨测任务
- datakit: 实际上是 datakit 中开启了具体的拨测采集器，此处统称 datakit，它跟线上的其它 datakit 没有实质区别
- commit: 具体任务的 JSON 描述，它可以维护在 DataFlux 的 MySQL 中，也可以是一个简单的 JSON 文件
- push: 任何授信的第三方（dataflux, curl 等），都能通过 HTTP 接口，往拨测中心**推送**拨测任务，一般而言，一个 commit 对应一个拨测任务，也可能多个 commit 对应的只有一个拨测任务（该任务经过多次修改，每个修改对应一个 commit）
- clone: datakit 上采集器初次启动时，从中心同步**指定 region 上所有任务的 commit**
- pull: datakit 以一定的频率，从中心拉取特定 region 上**最新**的服务
- region: 拨测服务可能在全球设置 datakit 拨测节点，一个节点就是一个 region。一个节点可能有多个 datakit 参与，此处可设置一定的负载均衡策略（如全量 clone，但等分运行）
- fork: 每个拨测任务可选择在**一个或多个 region**，假定选择了 3 个 region，实际上是该拨测任务的 3 个 fork

说明：

- 解耦考虑：DataFlux 也好，其它第三方也罢，可以各自维护一套自己的拨测任务逻辑，用以管理其拨测任务。比如，SAAS 某工作空间，新建了一个拨测任务 A，在 SAAS 平台，A 可能是 MySQL 中的一条表记录；对 dialtesting 而言，A 只是 SAAS 本地任务的一个 commit，**尚未** push 到 dialtesting
- 便于部署：当 PAAS/SAAS 开通了拨测服务之后，即可将这些拨测任务 push 到 dialtesting 服务端。对于未开通拨测服务的平台，push 接口**应当**报错，但不影响具体 PAAS/SAAS 在 Web 前端提交新的拨测任务。一旦配置了 dialtesting 正确的授信信息，即可开通拨测服务（以 DataFlux 为例，可在拨测任务表中新增一列，代表是否 push 成功）
- 拨测服务暂定域名 `dialtesting.dataflux.cn`，基于跨域的考虑，拨测服务的 API 不会在 Web 页面发起
- 关于授信，可通过 AK/SK 等 API 签名的方式，故 dialtesting 应该提供完整的授权 API 以及开发文档，并尽可能提供常见语言的 SDK
- 关于 fork： 逻辑上，单个任务的多 region fork，多个 region 之间相互独立。当第三方关闭某 region 的 fork 时，比如，原来某任务 A 创建了 hangzhou/chengdu 两个 region 的 fork，但用户取消了 chengdu 的拨测任务后，理应会向 chengdu push 一个新的 commit，该 commit 中 task 的 `stop` 字段为 `true`

### 数据库表定义

```sql
-- 存储拨测任务信息
CREATE TABLE IF NOT EXITS task (
		`id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
		`uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 dialt_ 前缀',
		`region` varchar(48) NOT NULL COMMENT '部署区域（只能有一个区域）'

		`access_key` varchar(20) NOT NULL COMMENT '推送 commit 的 AK'

		`create_at` int(11) NOT NULL DEFAULT '-1',

		`task` text NOT NULL COMMENT '任务的 json 描述',
		`hash` varchar(128) NOT NULL COMMENT '任务 hash，md5(access_key+task)',

		INDEX `hash` (`idx_hash`) COMMENT '便于鉴定重复推送',

		PRIMARY KEY (`id`),
		UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 存储 AK/SK 信息
CREATE TABLE IF NOT EXITS aksk (
		`id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增 ID',
		`uuid` varchar(48) NOT NULL COMMENT '全局唯一 ID，带 aksk_ 前缀',
		`access_key` varchar(20) NOT NULL COMMENT '推送 commit 的 AK'
    `secret_key` varchar(40) NOT NULL COMMENT '推送 commit 的 SK'

		`create_at` int(11) NOT NULL DEFAULT '-1',
    `status` int(11) NOT NULL DEFAULT '0' COMMENT '状态 0: ok/1: 故障/2: 停用/3: 删除',
		PRIMARY KEY (`id`),
		UNIQUE KEY `uk_uuid` (`uuid`) COMMENT 'UUID 做成全局唯一',
    UNIQUE KEY `uk_ak` (`access_key`) COMMENT 'AK 做成全局唯一',
    UNIQUE KEY `uk_sk` (`secret_key`) COMMENT 'SK 做成全局唯一',
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### API 定义

#### `/v1/push | POST`

```
POST /v1/push HTTP/1.1
Authorization: DIALTESTING <AK>:<sign>
Content-Type: application/json

<具体的 task-json>

HTTP/1.1 200 OK  # 无 body 返回
```

#### `/v1/pull | GET`

参数：

- `region`：必填参数，指定拉取区域
- `from`：可选参数，如不填，则拉取指定区域的所有任务（clone）。否则拉取 `id >= from` 的所有任务（pull）

示例：

```
POST /v1/pull?region=<region>&from=<id> HTTP/1.1
Authorization: DIALTESTING <AK>:<sign>

HTTP/1.1 200 OK
{
	"access_key-1": [ // 某 PAAS 平台的 AK，下面即该平台 push 过来的所有 commit
		{ "tcp-task" },
		{ "http-task" },
		{ "dns-task" },
		...
	],

	"access_key-2": [
		{ "tcp-task" },
		{ "http-task" },
		{ "dns-task" },
		...
	],
}
```

注意，datakit 作为客户端 pull 任务，也需通过 AK/SK 签名方式拉取

### 拨测任务策略

中心任务管理策略

- 任何经过认证的第三方，都可以往 dialtesting 推送（push）拨测任务
- 对任意一个已有任务的更新、删除等操作，都会创建一个新的任务提交（commit），**`id`，`uuud` 等均不同**。但对于同一个任务 commit 的 push，如果 hash 值不变，push 接口直接 200 返回，不会创建新的 任务 commit

datakit 端任务处理策略

1. DataKit 采集器启动时，通过指定 region，从中心 clone 所有该 region 的任务。通过一定的合并策略，采集器最终执行合并后的具体的任务。合并策略：
	- 轮询一遍 clone 下来的所有 commit（如总共 10K 个 commit，其中**有效任务** 1K 个）
	- 取固定几个字段的值，通过一定的 hash 算法，即可判定是否是同一个任务
	- hash 算法：`md5(AK + task-json)` 除了上述定义的基础字段外，第三方平台可在其中添加任何其它字段，如工作空间信息，这些都会计入 hash 计算。一些不太理智的第三方，可提交全部相同的 task，但这不影响 dialtesting 的正常运转。授信的第三方，应该提交差异化的 `task-json`。考虑到计算 hash 的性能开销以及高频率，这里暂用 md5（参见[这里](https://stackoverflow.com/questions/14139727/sha-256-or-md5-for-file-integrity)）

2. DataKit 采集器以一定频率，从中心同步最新的任务。
	- 对于更新了配置的任务，直接更新执行（将新的任务 json 发给当前运行的 go routine 即可）
	- 对于删除的任务，停止执行
	- 对于新增的任务，新开任务执行
关于更新频率，可在初次 clone 时，由中心带下去，便于统一调整

### DataKit 配置

开启对应拨测采集器，其 conf 如下

```python
[[inputs.dialtesting]]

	# required
	region = "hangzhou" 

	# default dialtesting.dataflux.cn
	server = "dialtesting.dataflux.cn" 

	[[inputs.dialtesting.tags]]
	# 各种可能的 tag
```
