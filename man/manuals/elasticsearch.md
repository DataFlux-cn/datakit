{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

ElasticSearch 采集器主要采集节点运行情况、集群健康、JVM 性能状况、索引性能、检索性能等。

## 前置条件

- ElasticSearch 默认采集 `Node Stats` 指标，如果需要采集 `Cluster-Health` 相关指标，需要设置 `cluster_health = true`
- 设置 `cluster_health = true` 可产生如下指标集
  - `elasticsearch_cluster_health`
- 设置 `cluster_stats = true` 可产生如下指标集
  - `elasticsearch_cluster_stats`

- 其他信息请参考配置文件说明

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 


## 日志采集

如需采集 ElasticSearch 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 ElasticSearch 日志文件的绝对路径。比如：

```
[[inputs.elasticsearch]]
  ...
[inputs.elasticsearch.log]
files = ["/path/to/your/file.log"]
```


开启日志采集以后，默认会产生日志来源（`source`）为 `elasticsearch` 的日志。

**注意**

- 日志采集仅支持采集已安装 DataKit 主机上的日志
