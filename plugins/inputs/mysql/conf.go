package mysql

const (
	configSample = `
[[inputs.mysql]]
    host = "localhost"
    user = "datakit"
    pass = "<PASS>"
    port = 3306
    # sock = "<SOCK>"
    # charset = "utf8"

    ## @param connect_timeout - number - optional - default: 10
    # connect_timeout = 10

    ## @param service - string - optional
    # service = "<SERVICE>"

    interval = "10s"

    ## @param inno_db
    innodb = false

    [inputs.mysql.log]
    # required, glob logfiles
    files = ["/var/log/mysql/*.log"]

    # glob filteer
    ignore = [""]

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    # regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^(# Time|\d{4}-\d{2}-\d{2}|\d{6}\s+\d{2}:\d{2}:\d{2}).*'''

    # grok pipeline script path
    pipeline = "mysql.p"

    # [[inputs.mysql.custom_queries]]
    #     sql = "SELECT foo, COUNT(*) FROM table.events GROUP BY foo"
    #     metric = "xxxx"
    #     tagKeys = ["column1", "column1"]
    #     fieldKeys = ["column3", "column1"]

    [inputs.mysql.tags]
    # tag1 = val1
    # tag2 = val2
`

	pipelineCfg = `
grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
grok(_, "%{TIMESTAMP_ISO8601:time} %{INT:thread_id} \\[%{NOTSPACE:status}\\] %{GREEDYDATA:msg}")

add_pattern("date2", "%{YEAR}%{MONTHNUM2}%{MONTHDAY} %{TIME}")
grok(_, "%{date2:time} \\s+(InnoDB:|\\[%{NOTSPACE:status}\\])\\s+%{GREEDYDATA:msg}")

add_pattern("timeline", "# Time: %{TIMESTAMP_ISO8601:time}")
add_pattern("userline", "# User@Host: %{GREEDYDATA}\\[(%{NOTSPACE:db_ip})?\\]%{GREEDYDATA}")
add_pattern("kvline01", "# Query_time: %{NUMBER:query_time}\\s+Lock_time: %{NUMBER:lock_time}\\s+Rows_sent: %{INT:rows_sent}\\s+Rows_examined: %{INT:rows_examined}")

add_pattern("kvline02", "# Thread_id: %{GREEDYDATA}")
add_pattern("kvline03", "# Bytes_sent: %{GREEDYDATA}")

# multi-line SQLs
add_pattern("sqls", "(?s)(.*)")

grok(_, "%{timeline}\\n%{userline}\\n%{kvline01}(\\n)?(%{kvline02})?(\\n)?(%{kvline03})?\\n%{sqls:db_slow_statement}")

cast(thread_id, "int")
cast(query_id, "int")
cast(rows_sent, "int")
cast(rows_examined, "int")

cast(bytes_sent, "int")
cast(bytes_received, "int")
cast(killed, "int")
cast(errno, "int")
cast(query_timestamp, "int")
cast(query_time, "float")
cast(lock_time, "float")

nullif(thread_id, 0)
nullif(db_host, "")
nullif(killed, 0)
nullif(bytes_sent, 0)
nullif(bytes_received, 0)
nullif(errno, 0)

default_time(time)
`
)
