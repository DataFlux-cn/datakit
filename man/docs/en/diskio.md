<!-- This file required to translate to EN. -->
{{.CSS}}
# DiskIO
---

{{.AvailableArchs}}

---

diskio 采集器用于磁盘流量和时间的指标的采集

## 前置条件 {#requests}

对于部分旧版本 Windows 操作系统，如若遇到 Datakit 报错： **"The system cannot find the file specified."**

请以管理员身份运行 PowerShell，并执行：

```powershell
diskperf -Y
```

在执行成功后需要重启 Datakit 服务。

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数：
    
    | 环境变量名                            | 对应的配置参数项     | 参数示例                                                     |
    | :---                                  | ---                  | ---                                                          |
    | `ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER` | `skip_serial_number` | `true`/`false`                                               |
    | `ENV_INPUT_DISKIO_TAGS`               | `tags`               | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_DISKIO_INTERVAL`           | `interval`           | `10s`                                                        |
    | `ENV_INPUT_DISKIO_DEVICES`            | `devices`            | `'''^sdb\d*'''`                                              |
    | `ENV_INPUT_DISKIO_DEVICE_TAGS`        | `device_tags`        | `"ID_FS_TYPE", "ID_FS_USAGE"` 以英文逗号隔开                 |
    | `ENV_INPUT_DISKIO_NAME_TEMPLATES`     | `name_templates`     | `"$ID_FS_LABEL", "$DM_VG_NAME/$DM_LV_NAME"` 以英文逗号隔开   |

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
