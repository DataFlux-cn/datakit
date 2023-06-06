### `datetime()` {#fn-datetime}

[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7)

函数原型：`fn datetime(key, precision: str, fmt: str, tz: str = "")`

函数说明：将时间戳转成指定日期格式

函数参数

- `key`: 已经提取的时间戳
- `precision`：输入的时间戳精度(s, ms, us, ns)
- `fmt`：日期格式，提供内置日期格式且支持自定义日期格式
- `tz`: 时区 (可选参数)，将时间戳转换为指定时区的时间，默认使用主机的时区

内置日期格式：

| 内置格式      | 日期                                  | 描述                      |
| ---           | ---                                   | ---                       |
| "ANSI-C"      | "Mon Jan _2 15:04:05 2006"            |                           |
| "UnixDate"    | "Mon Jan _2 15:04:05 MST 2006"        |                           |
| "RubyDate"    | "Mon Jan 02 15:04:05 -0700 2006"      |                           |
| "RFC822"      | "02 Jan 06 15:04 MST"                 |                           |
| "RFC822Z"     | "02 Jan 06 15:04 -0700"               | RFC822 with numeric zone  |
| "RFC850"      | "Monday, 02-Jan-06 15:04:05 MST"      |                           |
| "RFC1123"     | "Mon, 02 Jan 2006 15:04:05 MST"       |                           |
| "RFC1123Z"    | "Mon, 02 Jan 2006 15:04:05 -0700"     | RFC1123 with numeric zone |
| "RFC3339"     | "2006-01-02T15:04:05Z07:00"           |                           |
| "RFC3339Nano" | "2006-01-02T15:04:05.999999999Z07:00" |                           |
| "Kitchen"     | "3:04PM"                              |                           |

自定义日期格式：

可通过占位符的组合自定义输出日期格式

| 字符  | 示例 | 描述                                                          |
| ---   | ---  | ---                                                           |
| a     | %a   | 星期的缩写，如 `Wed`                                          |
| A     | %A   | 星期的全写，如 `Wednesday`                                    |
| b     | %b   | 月份缩写，如 `Mar`                                            |
| B     | %B   | 月份的全写，如 `March`                                        |
| C     | %c   | 世纪数，当前年份除 100                                        |
| **d** | %d   | 一个月内的第几天；范围 `[01, 31]`                             |
| e     | %e   | 一个月内的第几天；范围 `[1, 31]`，使用空格填充                |
| **H** | %H   | 小时，使用 24 小时制； 范围 `[00, 23]`                        |
| I     | %I   | 小时，使用 12 小时制； 范围 `[01, 12]`                        |
| j     | %j   | 一年内的第几天，范围 `[001, 365]`                             |
| k     | %k   | 小时，使用 24 小时制； 范围 `[0, 23]`                         |
| l     | %l   | 小时，使用 12 小时制； 范围 `[1, 12]`，使用空格填充           |
| **m** | %m   | 月份，范围 `[01, 12]`                                         |
| **M** | %M   | 分钟，范围 `[00, 59]`                                         |
| n     | %n   | 表示换行符 `\n`                                               |
| p     | %p   | `AM` 或 `PM`                                                  |
| P     | %P   | `am` 或 `pm`                                                  |
| s     | %s   | 自 1970-01-01 00:00:00 UTC 来的的秒数                         |
| **S** | %S   | 秒数，范围 `[00, 60]`                                         |
| t     | %t   | 表示制表符 `\t`                                               |
| u     | %u   | 星期几，星期一为 1，范围 `[1, 7]`                             |
| w     | %w   | 星期几，星期天为 0, 范围 `[0, 6]`                             |
| y     | %y   | 年份，范围 `[00, 99]`                                         |
| **Y** | %Y   | 年份的十进制表示                                              |
| **z** | %z   | RFC 822/ISO 8601:1988 风格的时区 (如： `-0600` 或 `+0100` 等) |
| Z     | %Z   | 时区缩写，如 `CST`                                            |
| %     | %%   | 表示字符 `%`                                                  |

示例：

```python
# 待处理数据：
#    {
#        "a":{
#            "timestamp": "1610960605000",
#            "second":2
#        },
#        "age":47
#    }

# 处理脚本
json(_, a.timestamp)
datetime(a.timestamp, 'ms', 'RFC3339')
```

```python
# 处理脚本
ts = timestamp()
datetime(ts, 'ns', fmt='%Y-%m-%d %H:%M:%S', tz="UTC")

# 输出
{
  "ts": "2023-03-08 06:43:39"
}
```

```python
# 处理脚本
ts = timestamp()
datetime(ts, 'ns', '%m/%d/%y  %H:%M:%S %z', "Asia/Tokyo")

# 输出
{
  "ts": "03/08/23  15:44:59 +0900"
}
```