---
title     : 'Nginx'
summary   : 'Collect metrics of Nginx'
tags:
  - 'WEB SERVER'
  - 'MIDDLEWARE'
__int_icon      : 'icon/nginx'
dashboard :
  - desc  : 'Nginx'
    path  : 'dashboard/en/nginx'
monitor   :
  - desc  : 'None'
    path  : '-'
---


{{.AvailableArchs}}

---

NGINX collector can take many metrics from NGINX instances, such as the total number of requests, connections, cache and other metrics, and collect the metrics into Guance Cloud to help monitor and analyze various abnormal situations of NGINX.

## Config {#config}

### Requirements {#requirements}

- NGINX version >= `1.8.0`; Already tested version:
    - [x] 1.23.2
    - [x] 1.22.1
    - [x] 1.21.6
    - [x] 1.18.0
    - [x] 1.14.2
    - [x] 1.8.0

- NGINX collects the data of `http_stub_status_module` by default. When the `http_stub_status_module` is opened, see [here](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html){:target="_blank"}, which will report the data of NGINX measurements later.

- If you are using [VTS](https://github.com/vozlt/nginx-module-vts){:target="_blank"} or want to monitor more data, it is recommended to turn on VTS-related data collection by setting the option `use_vts` to `true` in `{{.InputName}}.conf`. For how to start VTS, see [here](https://github.com/vozlt/nginx-module-vts#synopsis){:target="_blank"}.

- After VTS function is turned on, the following measurements can be generated:

    - `nginx`
    - `nginx_server_zone`
    - `nginx_upstream_zone` (NGINX needs to configure [`upstream` related configuration](http://nginx.org/en/docs/http/ngx_http_upstream_module.html){:target="_blank"})
    - `nginx_cache_zone`    (NGINX needs to configure [`cache` related configuration](https://docs.nginx.com/nginx/admin-guide/content-cache/content-caching/){:target="_blank"})

- Take the example of generating the `nginx_upstream_zone` measurements. An example of NGINX-related configuration is as follows:

```nginx
    ...
    http {
       ...
       upstream your-upstreamname {
         server upstream-ip:upstream-port;
      }
       server {
       ...
       location / {
       root  html;
       index  index.html index.htm;
       proxy_pass http://yourupstreamname;
     }}}

```

- After the VTS function has been turned on, it is no longer necessary to collect the data of the `http_stub_status_module` module, because the data of the VTS module will include the data of the `http_stub_status_module` module.

- NGINX Plus users can still use the `http_stub_status_module` to collect basic data. Additionally, `http_api_module` should be enabled in the NGINX configuration file ([Reference](https://nginx.org/en/docs/http/ngx_http_api_module.html){:target="_blank"}) and set status_zone in the server blocks you want to monitor. The configuration example is as follows:

``` nginx
# enable http_api_module
server {
  listen 8080;
  location /api {
     api write=on;
  }
}
# monitor more detailed metrics
server {
  listen 80;
  status_zone <ZONE_NAME>;
  ...
}
```

- To enable NGINX Plus collection, you need to set the option `use_plus_api` to true in the `{{.InputName}}.conf` file and uncomment the `plus_api_url` option. (Note: VTS does not support NGINX Plus).

- NGINX Plus can generate the following measurements:

    - `nginx_location_zone`

### Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    [Inject collector configuration through ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting) to enable the collector

???+ attention

    The `url` address is subject to the specific configuration of nginx. The common usage is to use the `/basic_status` route.
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
[inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## Custom Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Log {#logging}

To collect NGINX logs, open `files` in {{.InputName}}.conf and write to the absolute path of the NGINX log file. For example:

```toml
    [[inputs.nginx]]
      ...
      [inputs.nginx.log]
    files = ["/var/log/nginx/access.log","/var/log/nginx/error.log"]
```

When log collection is turned on, logs with a log `source` of `nginx` are generated by default.

>Note: DataKit must be installed on the NGINX host to collect NGINX logs.

### Log Pipeline Feature Cut Field Description {#pipeline}

- NGINX error log cutting

Example error log text:

```log
2021/04/21 09:24:04 [alert] 7#7: *168 write() to "/var/log/nginx/access.log" failed (28: No space left on device) while logging request, client: 120.204.196.129, server: localhost, request: "GET / HTTP/1.1", host: "47.98.103.73"
```

The list of cut fields is as follows:

| Field Name       | Field Value                                   | Description                         |
| ---          | ---                                      | ---                          |
| status       | error                                    | Log level (alert changed to error)   |
| client_ip    | 120.204.196.129                          | client ip address            |
| server       | localhost                                | server address                  |
| http_method  | GET                                      | http request mode                |
| http_url     | /                                        | http request url                 |
| http_version | 1.1                                      | http version                 |
| ip_or_host   | 47.98.103.73                             | requestor ip or host             |
| msg          | 7#7: *168 write()...host: \"47.98.103.73 | Log content                     |
| time         | 1618968244000000000                      | Nanosecond timestamp (as line protocol time) |

Example of error log text:

```log
2021/04/29 16:24:38 [emerg] 50102#0: unexpected ";" in /usr/local/etc/nginx/nginx.conf:23
```

The list of cut fields is as follows:

| Field Name | Field Value                                                          | Description                         |
| ---    | ---                                                             | ---                          |
| `status` | `error`                                                           | Log level (`emerg` changed to `error`)   |
| `msg`    | `50102#0: unexpected \";\" in /usr/local/etc/nginx/nginx.conf:23` | log content                     |
| `time`   | `1619684678000000000`                                             | Nanosecond timestamp (as row protocol time) |

- NGINX access log cutting

Example of access log text:

```log
127.0.0.1 - - [24/Mar/2021:13:54:19 +0800] "GET /basic_status HTTP/1.1" 200 97 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.72 Safari/537.36"
```

The list of cut fields is as follows:

| Field Name       | Field Value                       | Description                         |
| ---          | ---                          | ---                          |
| `client_ip`    | `127.0.0.1`                    | Log level (`emerg` changed to `error`)   |
| `status`       | `ok`                           | log level                     |
| `status_code`  | `200`                          | http code                    |
| `http_method`  | `GET`                          | http request method                |
| `http_url`     | `/basic_status`                | http request url                 |
| `http_version` | `1.1`                          | http version                 |
| `agent`        | `Mozilla/5.0... Safari/537.36` | User-Agent                   |
| `browser`      | `Chrome`                       | browser                       |
| `browserVer`   | `89.0.4389.72`                 | browser version                   |
| `isMobile`     | `false`                        | Is it a cell phone                     |
| `engine`       | `AppleWebKit`                  | engine                         |
| `os`           | `Intel Mac OS X 11_1_0`        | system                         |
| `time`         | `1619243659000000000`          | Nanosecond timestamp (as line protocol time) |
