{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`


# 简介

RabbitMQ 采集器是通过插件 `rabbitmq-management` 采集数据监控 RabbitMQ ,它能够：

- RabbitMQ overview 总览，比如连接数、队列数、消息总数等
- 跟踪 RabbitMQ queue 信息，比如队列大小，消费者计数等
- 跟踪 RabbitMQ node 信息，比如使用的 `socket` `mem` 等
- 跟踪 RabbitMQ exchange 信息 ，比如 `message_publish_count` 等


## 前置条件

- 安装 `rabbitmq` 以 `Ubuntu` 为例

    ```shell
    sudo apt-get update
    sudo apt-get install rabbitmq-server
    sudo service rabbitmq-server start
    ```
      
- 开启 `REST API plug-ins` 
    
    ```shell
    sudo rabbitmq-plugins enable rabbitmq-management
    ```
      
- 创建 user，比如：
    
    ```shell
    sudo rabbitmqctl add_user dataflux <SECRET>
    sudo rabbitmqctl set_permissions  -p / dataflux "^aliveness-test$" "^amq\.default$" ".*"
    sudo rabbitmqctl set_user_tags dataflux monitoring
    ```

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

如需采集 RabbitMQ 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 RabbitMQ 日志文件的绝对路径。比如：

```toml
    [[inputs.rabbitmq]]
      ...
      [inputs.rabbitmq.log]
        files = ["/var/log/rabbitmq/rabbit@your-hostname.log"]
```

  
开启日志采集以后，默认会产生日志来源（`source`）为 `rabbitmq` 的日志。

**注意**

- 日志采集仅支持采集已安装 DataKit 主机上的日志
