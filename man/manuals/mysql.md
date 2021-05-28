{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

MySQL 指标采集，收集以下数据：

- mysql global status 基础数据采集
- scheam 相关数据
- innodb 相关指标
- 支持自定义查询数据采集

>  主从模式相关的 MySQL 指标采集尚未支持（Comming Soon...）

## 前置条件

- MySQL 版本 5.7+

- 创建监控账号（一般情况，需用 MySQL `root` 账号登陆才能创建 MySQL 用户）

```sql
    CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';
    
    -- MySQL 8.0+ create the datakitMonitor user with the native password hashing method
    CREATE USER 'datakitMonitor'@'localhost' IDENTIFIED WITH mysql_native_password by '<UNIQUEPASSWORD>';
```

备注：`localhost` 是本地连接，具体参考[这里](https://dev.mysql.com/doc/refman/8.0/en/creating-accounts.html)

- 授权

```sql
    GRANT PROCESS ON *.* TO 'datakitMonitor'@'localhost';
    show databases like 'performance_schema';
    GRANT SELECT ON performance_schema.* TO 'datakitMonitor'@'localhost';
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志采集

如需采集 MySQL 的日志，将配置中 log 相关的配置打开，如需要开启 MySQL 慢查询日志，需要开启慢查询日志，在 MySQL 中执行以下语句

```sql
SET GLOBAL slow_query_log = 'ON';

-- 未使用索引的查询也认为是一个可能的慢查询
set global log_queries_not_using_indexes = 'ON';
```

```python
[inputs.mysql.log]
    # 填入绝对路径
    files = ["/var/log/mysql/*.log"] 
```

> 注意：在使用日志采集时，需要将 DataKit 安装在 MySQL 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器

### RDS 格式的慢日志切割

MySQL 支持 RDS 格式的慢日志切割，只需在配置文件中将 `pipeline` 参数改为 `mysql_rds.p` 即可。
