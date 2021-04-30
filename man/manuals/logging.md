{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

## 简介

采集文件尾部数据（类似`tail -f`），上报到 DataFlux 中。

## 配置

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logging.conf.sample` 并命名为 `logging.conf`。示例如下：

``` toml
[[inputs.logging]]
    # 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定
    # 推荐使用绝对路径
    logfiles = ["/var/log/syslog"]
    
    # 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    ignore = [""]
    
    # 数据来源，如果为空，则默认使用 'default'
    source = ""

    # 新增标记tag，如果为空，则默认使用 $source
    service = ""

    # pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 pipeline
    pipeline = ""

    # 过滤对应 status:
    #   `emerg`,`alert`,`critical`,`error`,`warning`,`info`,`debug`,`OK`
    ignore_status = []

    # 选择编码，如果编码有误会导致数据无法查看。默认为空即可:
    #    `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
    character_encoding = ""

    ## 设置正则表达式，例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
    ## 符合此正则匹配的数据，将被认定为有效数据，否则会累积追加到上一条有效数据的末尾
    ## 使用3个单引号 '''this-regexp''' 避免转义
    ## 正则表达式链接：https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^\S'''
    
    # 自定义 tags
    [inputs.logging.tags]
      # tags1 = "value1"
```

### match 使用说明

设置正则表达式，符合此正则匹配的数据，将被认定为有效数据，否则会累积追加到上一条有效数据的末尾。
正则表达式文档[链接](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)

示例，原数据：
```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

match 配置 `^\d{4}-\d{2}-\d{2}.*`

结果数据
```
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 1.0" 1611746438938808642
testing,filename=/tmp/094318188 message="2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File \"/usr/local/lib/python3.6/dist-packages/flask/app.py\", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
" 1611746441941718584
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 5.0" 1611746443938917265
```

原数据的第3行至第6行，在经过 match 处理后，被追加到有效的第2行的末尾。

### pipeline 配置和使用

[pipeline 文档](/man?input=pipeline)

使用 pipeline 功能时，取其中的 `time` 字段作为此条数据的产生时间。如果没有 `time` 字段或解析此字段失败，默认使用当前时间

数据必须含有 `status` 字段。如果使用 pipeline 功能时且得到有效的 `status`，将在，否则默认使用 “info”。

有效的 `status` 字段值（不区分大小写）：

| status 有效字段值                | 对应值     |
| :---                             | ---        |
| `a`, `alert`                     | `alert`    |
| `c`, `critical`                  | `critical` |
| `e`, `error`                     | `error`    |
| `w`, `warning`                   | `warning`  |
| `n`, `notice`                    | `notice`   |
| `i`, `info`                      | `info`     |
| `d`, `debug`, `trace`, `verbose` | `debug`    |
| `o`, `s`, `OK`                   | `OK`       |

示例：假定文本数据如下：

```
12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6
```
pipeline 脚本：

```python
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")
grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")
group_in(serverity, ["#"], "warning", status)
cast(pid, "int")
default_time(time)
```

最终结果：

```python
{
    "message": "12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6",
    "msg": "Server started, Redis version 3.0.6",
    "pid": 12115,
    "role": "M",
    "serverity": "#",
    "status": "warning",
    "time": 1610127941572000000
}
```

- 如果配置文件中 `pipeline_path` 为空，默认使用 `<source-name>.p`
- 如果 `<source-name.p>` 不存在，将不启用 pipeline 功能
- 所有 pipeline 脚本文件，统一存放在 datakit 安装路径下的 pipeline 目录下
- 默认情况下，采集器会自动发现新文件，以确保符合规则的新文件能够尽快采

### glob 规则简述（图表数据[来源](https://rgb-24bit.github.io/blog/2018/glob.html)）

使用 glob 规则更方便地指定日志文件，以及自动发现和文件过滤。

| 通配符   | 描述                               | 例子           | 匹配                       | 不匹配                      |
| :--      | ---                                | ---            | ---                        | ----                        |
| `*`      | 匹配任意数量的任何字符，包括无     | `Law*`         | Law, Laws, Lawyer          | GrokLaw, La, aw             |
| `?`      | 匹配任何单个字符                   | `?at`          | Cat, cat, Bat, bat         | at                          |
| `[abc]`  | 匹配括号中给出的一个字符           | `[CB]at`       | Cat, Bat                   | cat, bat                    |
| `[a-z]`  | 匹配括号中给出的范围中的一个字符   | `Letter[0-9]`  | Letter0, Letter1 … Letter9 | Letters, Letter, Letter10   |
| `[!abc]` | 匹配括号中未给出的一个字符         | `[!C]at`       | Bat, bat, cat              | Cat                         |
| `[!a-z]` | 匹配不在括号内给定范围内的一个字符 | `Letter[!3-5]` | Letter1…                   | Letter3 … Letter5, Letterxx |

另需说明，除上述 glob 标准规则外，采集器也支持 `**` 进行递归地文件遍历，如示例配置所示。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
