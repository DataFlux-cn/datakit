package inputs

type TelegrafInput struct {
	name            string
	Catalog, Sample string
}

func (ti *TelegrafInput) Enabled() (int, []string) {

	mtx.RLock()
	defer mtx.RUnlock()

	arr, ok := inputInfos[ti.name]
	if !ok {
		return 0, nil
	}

	cfgs := []string{}
	for _, i := range arr {
		cfgs = append(cfgs, i.cfg)
	}
	return len(arr), cfgs
}

var (
	TelegrafInputs = map[string]*TelegrafInput{ // Name: Catalog

		"disk":          &TelegrafInput{name: "disk", Catalog: "host"},
		"cpu":           &TelegrafInput{name: "cpu", Catalog: "host"},
		"diskio":        &TelegrafInput{name: "diskio", Catalog: "host"},
		"mem":           &TelegrafInput{name: "mem", Catalog: "host"},
		"kernel":        &TelegrafInput{name: "kernel", Catalog: "host"},
		"swap":          &TelegrafInput{name: "swap", Catalog: "host"},
		"system":        &TelegrafInput{name: "system", Catalog: "host"},
		`systemd_units`: &TelegrafInput{name: "systemd_units", Catalog: "host"},

		"nvidia_smi": &TelegrafInput{name: "nvidia_smi", Catalog: "nvidia"},

		"iptables":        &TelegrafInput{name: "iptables", Catalog: "network"},
		"ping":            &TelegrafInput{name: "ping", Catalog: "network"},
		"net":             &TelegrafInput{name: "net", Catalog: "network"},
		"net_response":    &TelegrafInput{name: "net_response", Catalog: "network"},
		"http":            &TelegrafInput{name: "http", Catalog: "network"},
		"socket_listener": &TelegrafInput{name: "socket_listener", Catalog: "network"},

		"nginx":   &TelegrafInput{name: "nginx", Catalog: "nginx"},
		"tengine": &TelegrafInput{name: "tengine", Catalog: "tengine"},
		"apache":  &TelegrafInput{name: "apache", Catalog: "apache"},

		"mysql":         &TelegrafInput{name: "mysql", Catalog: "db"},
		"postgresql":    &TelegrafInput{name: "postgresql", Catalog: "db"},
		"mongodb":       &TelegrafInput{name: "mongodb", Catalog: "db"},
		"redis":         &TelegrafInput{name: "redis", Catalog: "db"},
		"elasticsearch": &TelegrafInput{name: "elasticsearch", Catalog: "db"},
		"sqlserver":     &TelegrafInput{name: "sqlserver", Catalog: "db"},
		"memcached":     &TelegrafInput{name: "memcached", Catalog: "db"},
		"solr":          &TelegrafInput{name: "solr", Catalog: "db"},
		"clickhouse":    &TelegrafInput{name: "clickhouse", Catalog: "db"},
		`influxdb`:      &TelegrafInput{name: "influxdb", Catalog: "db"},

		"openldap":  &TelegrafInput{name: "openldap", Catalog: "openldap"},
		"phpfpm":    &TelegrafInput{name: "phpfpm", Catalog: "phpfpm"},
		"activemq":  &TelegrafInput{name: "activemq", Catalog: "activemq"},
		"zookeeper": &TelegrafInput{name: "zookeeper", Catalog: "zookeeper"},
		"ceph":      &TelegrafInput{name: "ceph", Catalog: "ceph"},
		"dns_query": &TelegrafInput{name: "dns_query", Catalog: "dns_query"},

		"docker":     &TelegrafInput{name: "docker", Catalog: "docker"},
		"docker_log": &TelegrafInput{name: "docker_log", Catalog: "docker"},

		"rabbitmq":       &TelegrafInput{name: "rabbitmq", Catalog: "rabbitmq"},
		"nsq":            &TelegrafInput{name: "nsq", Catalog: "nsq"},
		"nsq_consumer":   &TelegrafInput{name: "nsq_consumer", Catalog: "nsq"},
		"kafka_consumer": &TelegrafInput{name: "kafka_consumer", Catalog: "kafka"},
		"mqtt_consumer":  &TelegrafInput{name: "mqtt_consumer", Catalog: "mqtt"},

		"fluentd":   &TelegrafInput{name: "fluentd", Catalog: "fluentd"},
		"haproxy":   &TelegrafInput{name: "haproxy", Catalog: "haproxy"},
		"jenkins":   &TelegrafInput{name: "jenkins", Catalog: "jenkins"},
		"kapacitor": &TelegrafInput{name: "kapacitor", Catalog: "kapacitor"},
		"ntpq":      &TelegrafInput{name: "ntpq", Catalog: "ntpq"},
		"openntpd":  &TelegrafInput{name: "openntpd", Catalog: "openntpd"},
		"processes": &TelegrafInput{name: "processes", Catalog: "processes"},
		"x509_cert": &TelegrafInput{name: "x509_cert", Catalog: "tls"},
		"nats":      &TelegrafInput{name: "nats", Catalog: "nats"},

		"win_services":      &TelegrafInput{name: "win_services", Catalog: "windows"},
		"win_perf_counters": &TelegrafInput{name: "win_perf_counters", Catalog: "windows"},

		"cloudwatch": &TelegrafInput{name: "cloudwatch", Catalog: "aws"},
		"vsphere":    &TelegrafInput{name: "vsphere", Catalog: "vmware"},
		"snmp":       &TelegrafInput{name: "snmp", Catalog: "snmp"},
		"exec":       &TelegrafInput{name: "exec", Catalog: "exec"},
		"syslog":     &TelegrafInput{name: "syslog", Catalog: "syslog"},
		"varnish":    &TelegrafInput{name: "varnish", Catalog: "varnish"},

		"kube_inventory": &TelegrafInput{name: "kube_inventory", Catalog: "k8s"},
		"kubernetes":     &TelegrafInput{name: "kubernetes", Catalog: "k8s"},

		"jolokia2_agent": &TelegrafInput{name: "jolokia2_agent", Catalog: "jolokia2_agent"},

		"amqp":          &TelegrafInput{name: "amqp", Catalog: "amqp"},
		"amqp_consumer": &TelegrafInput{name: "amqp_consumer", Catalog: "amqp"},

		"github": &TelegrafInput{name: "github", Catalog: "github"},
		"uwsgi":  &TelegrafInput{name: "uwsgi", Catalog: "uwsgi"},

		`consul`:     &TelegrafInput{name: "consul", Catalog: "consul"},
		`kibana`:     &TelegrafInput{name: "kibana", Catalog: "kibana"},
		`modbus`:     &TelegrafInput{name: "modbus", Catalog: "modbus"},
		`dotnetclr`:  &TelegrafInput{name: "dotnetclr", Catalog: "dotnetclr"},
		`aspdotnet`:  &TelegrafInput{name: "aspdotnet", Catalog: "aspdotnet"},
		`msexchange`: &TelegrafInput{name: "msexchange", Catalog: "msexchange"},
	}
)

func HaveTelegrafInputs() bool {

	mtx.RLock()
	defer mtx.RUnlock()

	for k, _ := range TelegrafInputs {
		_, ok := inputInfos[k]
		if ok {
			return true
		}
	}

	return false
}

func applyTelegrafSamples() {
	TelegrafInputs[`amqp_consumer`].Sample = `
[[inputs.amqp_consumer]]
# Broker to consume from.
#   deprecated in 1.7; use the brokers option
 url = "amqp://localhost:5672/influxdb"

# Brokers to consume from.  If multiple brokers are specified a random broker
# will be selected anytime a connection is established.  This can be
# helpful for load balancing when not using a dedicated load balancer.
 brokers = ["amqp://localhost:5672/influxdb"]

# Authentication credentials for the PLAIN auth_method.
 username = ""
 password = ""

# Name of the exchange to declare.  If unset, no exchange will be declared.
 exchange = "telegraf"

# Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
 exchange_type = "topic"

# If true, exchange will be passively declared.
 exchange_passive = false

# Exchange durability can be either "transient" or "durable".
 exchange_durability = "durable"

# Additional exchange arguments.
 exchange_arguments = { }
 exchange_arguments = {"hash_propery" = "timestamp"}

# AMQP queue name
 queue = "telegraf"

# AMQP queue durability can be "transient" or "durable".
 queue_durability = "durable"

# If true, queue will be passively declared.
 queue_passive = false

# A binding between the exchange and queue using this binding key is
# created.  If unset, no binding is created.
 binding_key = "#"

# Maximum number of messages server should give to the worker.
 prefetch_count = 50

# Maximum messages to read from the broker that have not been written by an
# output.  For best throughput set based on the number of metrics within
# each message and the size of the output's metric_batch_size.
#
# For example, if each message from the queue contains 10 metrics and the
# output metric_batch_size is 1000, setting this to 100 will ensure that a
# full batch is collected and the write is triggered immediately without
# waiting until the next flush_interval.
 max_undelivered_messages = 1000

# Auth method. PLAIN and EXTERNAL are supported
# Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
# described here: https://www.rabbitmq.com/plugins.html
 auth_method = "PLAIN"

# Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.pem"
#tls_key = "/etc/telegraf/key.pem"
# Use TLS but skip chain & host verification
#insecure_skip_verify = false

# Content encoding for message payloads, can be set to "gzip" to or
# "identity" to apply no encoding.
 content_encoding = "identity"

# Data format to consume.
# Each data format has its own unique set of configuration options, read
# more about them here:
# https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
 data_format = "influx"`

	TelegrafInputs[`github`].Sample = `
[[inputs.github]]
# List of repositories to monitor
repositories = [
	"influxdata/telegraf",
	"influxdata/influxdb"
]

# Github API access token.  Unauthenticated requests are limited to 60 per hour.
 access_token = ""

# Github API enterprise url. Github Enterprise accounts must specify their base url.
 enterprise_base_url = ""

# Timeout for HTTP requests.
 http_timeout = "5s"`

	TelegrafInputs[`amqp`].Sample = `
[[inputs.amqp_consumer]]
# Broker to consume from.
#   deprecated in 1.7; use the brokers option
 url = "amqp://localhost:5672/influxdb"

# Brokers to consume from.  If multiple brokers are specified a random broker
# will be selected anytime a connection is established.  This can be
# helpful for load balancing when not using a dedicated load balancer.
 brokers = ["amqp://localhost:5672/influxdb"]

# Authentication credentials for the PLAIN auth_method.
 username = ""
 password = ""

# Name of the exchange to declare.  If unset, no exchange will be declared.
 exchange = "telegraf"

# Exchange type; common types are "direct", "fanout", "topic", "header", "x-consistent-hash".
 exchange_type = "topic"

# If true, exchange will be passively declared.
 exchange_passive = false

# Exchange durability can be either "transient" or "durable".
 exchange_durability = "durable"

# Additional exchange arguments.
 exchange_arguments = { }
 exchange_arguments = {"hash_propery" = "timestamp"}

# AMQP queue name
 queue = "telegraf"

# AMQP queue durability can be "transient" or "durable".
 queue_durability = "durable"

# If true, queue will be passively declared.
 queue_passive = false

# A binding between the exchange and queue using this binding key is
# created.  If unset, no binding is created.
 binding_key = "#"

# Maximum number of messages server should give to the worker.
 prefetch_count = 50

# Maximum messages to read from the broker that have not been written by an
# output.  For best throughput set based on the number of metrics within
# each message and the size of the output's metric_batch_size.
#
# For example, if each message from the queue contains 10 metrics and the
# output metric_batch_size is 1000, setting this to 100 will ensure that a
# full batch is collected and the write is triggered immediately without
# waiting until the next flush_interval.
 max_undelivered_messages = 1000

# Auth method. PLAIN and EXTERNAL are supported
# Using EXTERNAL requires enabling the rabbitmq_auth_mechanism_ssl plugin as
# described here: https://www.rabbitmq.com/plugins.html
 auth_method = "PLAIN"

# Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.pem"
#tls_key = "/etc/telegraf/key.pem"
# Use TLS but skip chain & host verification
#insecure_skip_verify = false

# Content encoding for message payloads, can be set to "gzip" to or
# "identity" to apply no encoding.
 content_encoding = "identity"

# Data format to consume.
# Each data format has its own unique set of configuration options, read
# more about them here:
# https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
 data_format = "influx"`

	TelegrafInputs[`socket_listener`].Sample = `
# Generic socket listener capable of handling multiple socket types.
[[inputs.socket_listener]]
	## collectd
	service_address = "udp://:25826"
        data_format = "collectd"

        ## Authentication file for cryptographic security levels
        collectd_auth_file = "/etc/collectd/auth_file"

        ## One of none (default), sign, or encrypt
        collectd_security_level = "encrypt"

        ## Path of to TypesDB specifications
        collectd_typesdb = ["/usr/share/collectd/types.db"]

        ## Multi-value plugins can be handled two ways.
        ## "split" will parse and store the multi-value plugin data into separate measurements
        ## "join" will parse and store the multi-value plugin as a single multi-value measurement.
        ## "split" is the default behavior for backward compatability with previous versions of influxdb.
        collectd_parse_multivalue = "split"

	# ----
	# URL to listen on
	 service_address = "tcp://:8094"
	 service_address = "tcp://127.0.0.1:http"
	 service_address = "tcp4://:8094"
	 service_address = "tcp6://:8094"
	 service_address = "tcp6://[2001:db8::1]:8094"
	 service_address = "udp://:8094"
	 service_address = "udp4://:8094"
	 service_address = "udp6://:8094"
	 service_address = "unix:///tmp/telegraf.sock"
	 service_address = "unixgram:///tmp/telegraf.sock"

	# Change the file mode bits on unix sockets.  These permissions may not be
	# respected by some platforms, to safely restrict write permissions it is best
	# to place the socket into a directory that has previously been created
	# with the desired permissions.
	#   ex: socket_mode = "777"
	 socket_mode = ""

	# Maximum number of concurrent connections.
	# Only applies to stream sockets (e.g. TCP).
	# 0 (default) is unlimited.
	 max_connections = 1024

	# Read timeout.
	# Only applies to stream sockets (e.g. TCP).
	# 0 (default) is unlimited.
	 read_timeout = "30s"

	# Optional TLS configuration.
	# Only applies to stream sockets (e.g. TCP).
	#tls_cert = "/etc/telegraf/cert.pem"
	#tls_key  = "/etc/telegraf/key.pem"
	# Enables client authentication if set.
	#tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

	# Maximum socket buffer size (in bytes when no unit specified).
	# For stream sockets, once the buffer fills up, the sender will start backing up.
	# For datagram sockets, once the buffer fills up, metrics will start dropping.
	# Defaults to the OS default.
	 read_buffer_size = "64KiB"

	# Period between keep alive probes.
	# Only applies to TCP sockets.
	# 0 disables keep alive probes.
	# Defaults to the OS configuration.
	 keep_alive_period = "5m"

	# Data format to consume.
	# Each data format has its own unique set of configuration options, read
	# more about them here:
	# https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
	 data_format = "influx"

	# Content encoding for message payloads, can be set to "gzip" to or
	# "identity" to apply no encoding.
	 content_encoding = "identity"`

	TelegrafInputs[`jolokia2_agent`].Sample = `
[[inputs.jolokia2_agent]]
urls = ["http://agent:8080/jolokia"]

[[inputs.jolokia2_agent.metric]]
	name  = "jvm_runtime"
	mbean = "java.lang:type=Runtime"
	paths = ["Uptime"]

#Optionally, specify TLS options for communicating with agents:
#[[inputs.jolokia2_agent]]
#	urls = ["https://agent:8080/jolokia"]
#	tls_ca   = "/var/private/ca.pem"
#	tls_cert = "/var/private/client.pem"
#	tls_key  = "/var/private/client-key.pem"
#	#insecure_skip_verify = false
#
#	[[inputs.jolokia2_agent.metric]]
#	name  = "jvm_runtime"
#	mbean = "java.lang:type=Runtime"
#	paths = ["Uptime"]

# #Jolokia Proxy Configuration
# #The jolokia2_proxy input plugin reads JMX metrics from one or more targets by interacting with a Jolokia ## ## #proxy REST endpoint.

#[[inputs.jolokia2_proxy]]
#  url = "http://proxy:8080/jolokia"
#
#  #default_target_username = ""
#  #default_target_password = ""
#  [[inputs.jolokia2_proxy.target]]
#    url = "service:jmx:rmi:///jndi/rmi://targethost:9999/jmxrmi"
#    # username = ""
#    # password = ""
#
#  [[inputs.jolokia2_proxy.metric]]
#    name  = "jvm_runtime"
#    mbean = "java.lang:type=Runtime"
#    paths = ["Uptime"]`

	TelegrafInputs[`kafka_consumer`].Sample = `
[[inputs.kafka_consumer]]
# Kafka brokers.
 brokers = ["localhost:9092"]

# Topics to consume.
topics = ["telegraf"]

# When set this tag will be added to all metrics with the topic as the value.
# topic_tag = ""

# Optional Client id
#client_id = "datakit-agent"

# Set the minimal supported Kafka version.  Setting this enables the use of new
# Kafka features and APIs.  Must be 0.10.2.0 or greater.
#   ex: version = "1.1.0"
# version = ""

# Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.pem"
#tls_key = "/etc/telegraf/key.pem"
# Use TLS but skip chain & host verification
#insecure_skip_verify = false

# SASL authentication credentials.  These settings should typically be used
# with TLS encryption enabled using the "enable_tls" option.
# sasl_username = "kafka"
# sasl_password = "secret"

# SASL protocol version.  When connecting to Azure EventHub set to 0.
# sasl_version = 1

# Name of the consumer group.
# consumer_group = "telegraf_metrics_consumers"

# Initial offset position; one of "oldest" or "newest".
# offset = "oldest"

# Consumer group partition assignment strategy; one of "range", "roundrobin" or "sticky".
 balance_strategy = "range"

# Maximum length of a message to consume, in bytes (default 0/unlimited);
# larger messages are dropped
 max_message_len = 1000000

# Maximum messages to read from the broker that have not been written by an
# output.  For best throughput set based on the number of metrics within
# each message and the size of the output's metric_batch_size.
#
# For example, if each message from the queue contains 10 metrics and the
# output metric_batch_size is 1000, setting this to 100 will ensure that a
# full batch is collected and the write is triggered immediately without
# waiting until the next flush_interval.
# max_undelivered_messages = 1000

# Data format to consume.
# Each data format has its own unique set of configuration options, read
# more about them here:
# https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
 data_format = "influx"`

	TelegrafInputs[`mqtt_consumer`].Sample = `
[[inputs.mqtt_consumer]]
## Broker URLs for the MQTT server or cluster.  To connect to multiple
## clusters or standalone servers, use a seperate plugin instance.
##   example: servers = ["tcp://localhost:1883"]
##            servers = ["ssl://localhost:1883"]
##            servers = ["ws://localhost:1883"]
servers = ["tcp://127.0.0.1:1883"]

## Topics that will be subscribed to.
topics = [
  "telegraf/host01/cpu",
  "telegraf/+/mem",
  "sensors/#",
]

## The message topic will be stored in a tag specified by this value.  If set
## to the empty string no topic tag will be created.
# topic_tag = "topic"

## QoS policy for messages
##   0 = at most once
##   1 = at least once
##   2 = exactly once
##
## When using a QoS of 1 or 2, you should enable persistent_session to allow
## resuming unacknowledged messages.
# qos = 0

## Connection timeout for initial connection in seconds
# connection_timeout = "30s"

## Maximum messages to read from the broker that have not been written by an
## output.  For best throughput set based on the number of metrics within
## each message and the size of the output's metric_batch_size.
##
## For example, if each message from the queue contains 10 metrics and the
## output metric_batch_size is 1000, setting this to 100 will ensure that a
## full batch is collected and the write is triggered immediately without
## waiting until the next flush_interval.
# max_undelivered_messages = 1000

## Persistent session disables clearing of the client session on connection.
## In order for this option to work you must also set client_id to identify
## the client.  To receive messages that arrived while the client is offline,
## also set the qos option to 1 or 2 and don't forget to also set the QoS when
## publishing.
# persistent_session = false

## If unset, a random client ID will be generated.
# client_id = ""

## Username and password to connect MQTT server.
# username = "telegraf"
# password = "metricsmetricsmetricsmetrics"

## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false

## Data format to consume.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
data_format = "influx"
`
	TelegrafInputs[`kube_inventory`].Sample = `
[[inputs.kube_inventory]]
# URL for the Kubernetes API
url = "https://127.0.0.1"

# Namespace to use. Set to "" to use all namespaces.
 namespace = "default"

# Use bearer token for authorization. ('bearer_token' takes priority)
# If both of these are empty, we'll use the default serviceaccount:
# at: /run/secrets/kubernetes.io/serviceaccount/token
 bearer_token = "/path/to/bearer/token"
# OR
 bearer_token_string = "abc_123"

# Set response_timeout (default 5 seconds)
 response_timeout = "5s"

# Optional Resources to exclude from gathering
# Leave them with blank with try to gather everything available.
# Values can be - "daemonsets", deployments", "endpoints", "ingress", "nodes",
# "persistentvolumes", "persistentvolumeclaims", "pods", "services", "statefulsets"
#resource_exclude = [ "deployments", "nodes", "statefulsets" ]

# Optional Resources to include when gathering
# Overrides resource_exclude if both set.
#resource_include = [ "deployments", "nodes", "statefulsets" ]

# Optional TLS Config
#tls_ca = "/path/to/cafile"
#tls_cert = "/path/to/certfile"
#tls_key = "/path/to/keyfile"
# Use TLS but skip chain & host verification
#insecure_skip_verify = false`

	TelegrafInputs[`kubernetes`].Sample = `
[[inputs.kubernetes]]
# URL for the kubelet
url = "http://127.0.0.1:10255"

# Use bearer token for authorization. ('bearer_token' takes priority)
# If both of these are empty, we'll use the default serviceaccount:
# at: /run/secrets/kubernetes.io/serviceaccount/token
 bearer_token = "/path/to/bearer/token"
# OR
 bearer_token_string = "abc_123"

# Pod labels to be added as tags.  An empty array for both include and
# exclude will include all labels.
 label_include = []
 label_exclude = ["*"]

# Set response_timeout (default 5 seconds)
 response_timeout = "5s"

# Optional TLS Config
#tls_ca = /path/to/cafile
#tls_cert = /path/to/certfile
#tls_key = /path/to/keyfile
# Use TLS but skip chain & host verification
#insecure_skip_verify = false`

	TelegrafInputs[`nsq`].Sample = `
[[inputs.nsq]]
# An array of NSQD HTTP API endpoints
endpoints  = ["http://localhost:4151"]

# Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.pem"
#tls_key = "/etc/telegraf/key.pem"
# Use TLS but skip chain & host verification
#insecure_skip_verify = false`

	TelegrafInputs[`nsq_consumer`].Sample = `
 # Read NSQ topic for metrics.
 [[inputs.nsq_consumer]]
   ## Server option still works but is deprecated, we just prepend it to the nsqd array.
   # server = "localhost:4150"
   ## An array representing the NSQD TCP HTTP Endpoints
   nsqd = ["localhost:4150"]
   ## An array representing the NSQLookupd HTTP Endpoints
   nsqlookupd = ["localhost:4161"]
   topic = "telegraf"
   channel = "consumer"
   max_in_flight = 100

   ## Maximum messages to read from the broker that have not been written by an
   ## output.  For best throughput set based on the number of metrics within
   ## each message and the size of the output's metric_batch_size.
   ##
   ## For example, if each message from the queue contains 10 metrics and the
   ## output metric_batch_size is 1000, setting this to 100 will ensure that a
   ## full batch is collected and the write is triggered immediately without
   ## waiting until the next flush_interval.
   # max_undelivered_messages = 1000

   ## Data format to consume.
   ## Each data format has its own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
   data_format = "influx"`

	TelegrafInputs[`varnish`].Sample = `
# A plugin to collect stats from Varnish HTTP Cache
[[inputs.varnish]]
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["MAIN.*"]
  ## stats may also be set to ["*"], which will collect all stats
  stats = ["MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually appened after -n in varnish cli
  # instance_name = instanceName

  ## Timeout for varnishstat command
  # timeout = "1s"`

	TelegrafInputs[`syslog`].Sample = `
 # Accepts syslog messages following RFC5424 format with transports as per RFC5426, RFC5425, or RFC6587
 [[inputs.syslog]]
   ## Specify an ip or hostname with port - eg., tcp://localhost:6514, tcp://10.0.0.1:6514
   ## Protocol, address and port to host the syslog receiver.
   ## If no host is specified, then localhost is used.
   ## If no port is specified, 6514 is used (RFC5425#section-4.1).
   server = "tcp://:6514"

   ## TLS Config
   # tls_allowed_cacerts = ["/etc/telegraf/ca.pem"]
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"

   ## Period between keep alive probes.
   ## 0 disables keep alive probes.
   ## Defaults to the OS configuration.
   ## Only applies to stream sockets (e.g. TCP).
   # keep_alive_period = "5m"

   ## Maximum number of concurrent connections (default = 0).
   ## 0 means unlimited.
   ## Only applies to stream sockets (e.g. TCP).
   # max_connections = 1024

   ## Read timeout is the maximum time allowed for reading a single message (default = 5s).
   ## 0 means unlimited.
   # read_timeout = "5s"

   ## The framing technique with which it is expected that messages are transported (default = "octet-counting").
   ## Whether the messages come using the octect-counting (RFC5425#section-4.3.1, RFC6587#section-3.4.1),
   ## or the non-transparent framing technique (RFC6587#section-3.4.2).
   ## Must be one of "octet-counting", "non-transparent".
   # framing = "octet-counting"

   ## The trailer to be expected in case of non-trasparent framing (default = "LF").
   ## Must be one of "LF", or "NUL".
   # trailer = "LF"

   ## Whether to parse in best effort mode or not (default = false).
   ## By default best effort parsing is off.
   # best_effort = false

   ## Character to prepend to SD-PARAMs (default = "_").
   ## A syslog message can contain multiple parameters and multiple identifiers within structured data section.
   ## Eg., [id1 name1="val1" name2="val2"][id2 name1="val1" nameA="valA"]
   ## For each combination a field is created.
   ## Its name is created concatenating identifier, sdparam_separator, and parameter name.
   # sdparam_separator = "_"`

	TelegrafInputs[`exec`].Sample = `
 # Read metrics from one or more commands that can output to stdout
 [[inputs.exec]]
   ## Commands array
   commands = [
     "/tmp/test.sh",
     "/usr/bin/mycollector --foo=bar",
     "/tmp/collect_*.sh"
   ]

   ## Timeout for each command to complete.
   timeout = "5s"

   ## measurement name suffix (for separating different commands)
   name_suffix = "_mycollector"

   ## Data format to consume.
   ## Each data format has its own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
   data_format = "influx"`

	TelegrafInputs[`snmp`].Sample = `
 # Retrieves SNMP values from remote agents
 [[inputs.snmp]]
   agents = [ "127.0.0.1:161" ]
   ## Timeout for each SNMP query.
   timeout = "5s"
   ## Number of retries to attempt within timeout.
   retries = 3
   ## SNMP version, values can be 1, 2, or 3
   version = 2

   ## SNMP community string.
   community = "public"

   ## The GETBULK max-repetitions parameter
   max_repetitions = 10

   ## SNMPv3 auth parameters
   #sec_name = "myuser"
   #auth_protocol = "md5"      # Values: "MD5", "SHA", ""
   #auth_password = "pass"
   #sec_level = "authNoPriv"   # Values: "noAuthNoPriv", "authNoPriv", "authPriv"
   #context_name = ""
   #priv_protocol = ""         # Values: "DES", "AES", ""
   #priv_password = ""

   ## measurement name
   name = "system"
   [[inputs.snmp.field]]
     name = "hostname"
     oid = ".1.0.0.1.1"
   [[inputs.snmp.field]]
     name = "uptime"
     oid = ".1.0.0.1.2"
   [[inputs.snmp.field]]
     name = "load"
     oid = ".1.0.0.1.3"
   [[inputs.snmp.field]]
     oid = "HOST-RESOURCES-MIB::hrMemorySize"

   [[inputs.snmp.table]]
     ## measurement name
     name = "remote_servers"
     inherit_tags = [ "hostname" ]
     [[inputs.snmp.table.field]]
       name = "server"
       oid = ".1.0.0.0.1.0"
       is_tag = true
     [[inputs.snmp.table.field]]
       name = "connections"
       oid = ".1.0.0.0.1.1"
     [[inputs.snmp.table.field]]
       name = "latency"
       oid = ".1.0.0.0.1.2"

   [[inputs.snmp.table]]
     ## auto populate table's fields using the MIB
     oid = "HOST-RESOURCES-MIB::hrNetworkTable"`

	TelegrafInputs[`win_perf_counters`].Sample = `
[[inputs.win_perf_counters]]
[[inputs.win_perf_counters.object]]
 ##Processor usage, alternative to native, reports on a per core.
ObjectName = "Processor"
Instances = ["*"]
Counters = ["% Idle Time", "% Interrupt Time", "% Privileged Time", "% User Time", "% Processor Time"]
Measurement = "win_cpu"
IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
 ##Disk times and queues
ObjectName = "LogicalDisk"
Instances = ["*"]
Counters = ["% Idle Time", "% Disk Time","% Disk Read Time", "% Disk Write Time", "% User Time", "Current Disk Queue Length"]
Measurement = "win_disk"
IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
ObjectName = "System"
Counters = ["Context Switches/sec","System Calls/sec", "Processor Queue Length"]
Instances = ["------"]
Measurement = "win_system"
IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
 ##Example query where the Instance portion must be removed to get data back, such as from the Memory object.
ObjectName = "Memory"
Counters = ["Available Bytes","Cache Faults/sec","Demand Zero Faults/sec","Page Faults/sec","Pages/sec","Transition Faults/sec","Pool #Nonpaged Bytes","Pool Paged Bytes"]
Instances = ["------"] # Use 6 x - to remove the Instance bit from the query.
Measurement = "win_mem"
IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
 ##more counters for the Network Interface Object can be found at
 https://msdn.microsoft.com/en-us/library/ms803962.aspx
ObjectName = "Network Interface"
Counters = ["Bytes Received/sec","Bytes Sent/sec","Packets Received/sec","Packets Sent/sec"]
Instances = ["*"] # Use 6 x - to remove the Instance bit from the query.
Measurement = "win_net"
IncludeTotal=false #Set to true to include _Total instance when querying for all (*). `

	TelegrafInputs[`vsphere`].Sample = `
# Read metrics from one or many vCenters
# See: https://github.com/influxdata/telegraf/tree/master/plugins/inputs/vsphere
[[inputs.vsphere]]
	## List of vCenter URLs to be monitored. These three lines must be uncommented
	## and edited for the plugin to work.
	vcenters = [ "https://vcenter.local/sdk" ]
	username = "user@corp.local"
	password = "secret"

	## VMs
	## Typical VM metrics (if omitted or empty, all metrics are collected)
	# vm_include = [ "/*/vm/**"] # Inventory path to VMs to collect (by default all are collected)
	vm_metric_include = [
	"cpu.demand.average",
	]
	vm_metric_exclude = [] ## Nothing is excluded by default
	vm_instances = true ## true by default

	## Hosts
	## Typical host metrics (if omitted or empty, all metrics are collected)
	host_include = [ "/*/host/**"] # Inventory path to hosts to collect (by default all are collected)
	host_metric_include = [
	"cpu.coreUtilization.average",
	]
	## Collect IP addresses? Valid values are "ipv4" and "ipv6"
	# ip_addresses = ["ipv6", "ipv4" ]

	# host_metric_exclude = [] ## Nothing excluded by default
	# host_instances = true ## true by default

	## Clusters
	# cluster_include = [ "/*/host/**"] # Inventory path to clusters to collect (by default all are collected)
	# cluster_metric_include = [] ## if omitted or empty, all metrics are collected
	# cluster_metric_exclude = [] ## Nothing excluded by default
	# cluster_instances = false ## false by default

	## Datastores
	# cluster_include = [ "/*/datastore/**"] # Inventory path to datastores to collect (by default all are collected)
	# datastore_metric_include = [] ## if omitted or empty, all metrics are collected
	# datastore_metric_exclude = [] ## Nothing excluded by default
	# datastore_instances = false ## false by default

	## Datacenters
	# datacenter_include = [ "/*/host/**"] # Inventory path to clusters to collect (by default all are collected)
	# datacenter_metric_include = [] ## if omitted or empty, all metrics are collected
	# datacenter_metric_exclude = [ "*" ] ## Datacenters are not collected by default.
	# datacenter_instances = false ## false by default

	## Plugin Settings
	## separator character to use for measurement and field names (default: "_")
	# separator = "_"

	## number of objects to retreive per query for realtime resources (vms and hosts)
	## set to 64 for vCenter 5.5 and 6.0 (default: 256)
	# max_query_objects = 256

	## number of metrics to retreive per query for non-realtime resources (clusters and datastores)
	## set to 64 for vCenter 5.5 and 6.0 (default: 256)
	# max_query_metrics = 256

	## number of go routines to use for collection and discovery of objects and metrics
	# collect_concurrency = 1
	# discover_concurrency = 1

	## whether or not to force discovery of new objects on initial gather call before collecting metrics
	## when true for large environments this may cause errors for time elapsed while collecting metrics
	## when false (default) the first collection cycle may result in no or limited metrics while objects are discovered
	# force_discover_on_init = false

	## the interval before (re)discovering objects subject to metrics collection (default: 300s)
	# object_discovery_interval = "300s"

	## timeout applies to any of the api request made to vcenter
	# timeout = "60s"

	## When set to true, all samples are sent as integers. This makes the output
	## data types backwards compatible with Telegraf 1.9 or lower. Normally all
	## samples from vCenter, with the exception of percentages, are integer
	## values, but under some conditions, some averaging takes place internally in
	## the plugin. Setting this flag to "false" will send values as floats to
	## preserve the full precision when averaging takes place.
	# use_int_samples = true

	## Custom attributes from vCenter can be very useful for queries in order to slice the
	## metrics along different dimension and for forming ad-hoc relationships. They are disabled
	## by default, since they can add a considerable amount of tags to the resulting metrics. To
	## enable, simply set custom_attribute_exlude to [] (empty set) and use custom_attribute_include
	## to select the attributes you want to include.
	# by default, since they can add a considerable amount of tags to the resulting metrics. To
	# enable, simply set custom_attribute_exlude to [] (empty set) and use custom_attribute_include
	# to select the attributes you want to include.
	# custom_attribute_include = []
	# custom_attribute_exclude = ["*"] # Default is to exclude everything

	## Optional SSL Config
	# ssl_ca = "/path/to/cafile"
	# ssl_cert = "/path/to/certfile"
	# ssl_key = "/path/to/keyfile"
	## Use SSL but skip chain & host verification
	# insecure_skip_verify = false`

	TelegrafInputs["cloudwatch"].Sample = `
	[[inputs.cloudwatch]]
  ## Amazon Region
  region = "us-east-1"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  # The minimum period for Cloudwatch metrics is 1 minute (60s). However not all
  # metrics are made available to the 1 minute period. Some are collected at
  # 3 minute, 5 minute, or larger intervals. See https://aws.amazon.com/cloudwatch/faqs/#monitoring.
  # Note that if a period is configured that is smaller than the minimum for a
  # particular metric, that metric will not be returned by the Cloudwatch API
  # and will not be collected by Telegraf.
  #
  ## Requested CloudWatch aggregation Period (required - must be a multiple of 60s)
  period = "5m"

  ## Collection Delay (required - must account for metrics availability via CloudWatch API)
  delay = "5m"

  ## Recommended: use metric 'interval' that is a multiple of 'period' to avoid
  ## gaps or overlap in pulled data
  interval = "5m"

  ## Configure the TTL for the internal cache of metrics.
  # cache_ttl = "1h"

  ## Metric Statistic Namespace (required)
  namespace = "AWS/ELB"

  ## Maximum requests per second. Note that the global default AWS rate limit is
  ## 50 reqs/sec, so if you define multiple namespaces, these should add up to a
  ## maximum of 50.
  ## See http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
  # ratelimit = 25

  ## Timeout for http requests made by the cloudwatch client.
  # timeout = "5s"

  ## Namespace-wide statistic filters. These allow fewer queries to be made to
  ## cloudwatch.
  # statistic_include = [ "average", "sum", "minimum", "maximum", sample_count" ]
  # statistic_exclude = []

  ## Metrics to Pull
  ## Defaults to all Metrics in Namespace if nothing is provided
  ## Refreshes Namespace available metrics every 1h
  #[[inputs.cloudwatch.metrics]]
  #  names = ["Latency", "RequestCount"]
  #
  #  ## Statistic filters for Metric.  These allow for retrieving specific
  #  ## statistics for an individual metric.
  #  # statistic_include = [ "average", "sum", "minimum", "maximum", sample_count" ]
  #  # statistic_exclude = []
  #
  #  ## Dimension filters for Metric.  All dimensions defined for the metric names
  #  ## must be specified in order to retrieve the metric statistics.
  #  [[inputs.cloudwatch.metrics.dimensions]]
  #    name = "LoadBalancerName"
  #    value = "p-example"`

	TelegrafInputs["win_services"].Sample = `
[[inputs.win_services]]
# Reports information about Windows service status.
# Monitoring some services may require running Telegraf with administrator privileges.
# Names of the services to monitor. Leave empty to monitor all the available services on the host
service_names = [
	"LanmanServer",
	"TermService",
]`

	TelegrafInputs["x509_cert"].Sample = `
 # Reads metrics from a SSL certificate
 [[inputs.x509_cert]]
   ## List certificate sources
   sources = ["/etc/ssl/certs/ssl-cert-snakeoil.pem", "tcp://example.org:443"]

   ## Timeout for SSL connection
   # timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"`

	TelegrafInputs["tengine"].Sample = `
 # Read Tengine's basic status information (ngx_http_reqstat_module)
 [[inputs.tengine]]
   # An array of Tengine reqstat module URI to gather stats.
   urls = ["http://127.0.0.1/us"]

   # HTTP response timeout (default: 5s)
   # response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.cer"
   # tls_key = "/etc/telegraf/key.key"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs["rabbitmq"].Sample = `
 # Reads metrics from RabbitMQ servers via the Management Plugin
 [[inputs.rabbitmq]]
   ## Management Plugin url. (default: http://localhost:15672)
   # url = "http://localhost:15672"
   ## Tag added to rabbitmq_overview series; deprecated: use tags
   # name = "rmq-server-1"
   ## Credentials
   # username = "guest"
   # password = "guest"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## Optional request timeouts
   ##
   ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
   ## for a server's response headers after fully writing the request.
   # header_timeout = "3s"
   ##
   ## client_timeout specifies a time limit for requests made by this client.
   ## Includes connection time, any redirects, and reading the response body.
   # client_timeout = "4s"

   ## A list of nodes to gather as the rabbitmq_node measurement. If not
   ## specified, metrics for all nodes are gathered.
   # nodes = ["rabbit@node1", "rabbit@node2"]

   ## A list of queues to gather as the rabbitmq_queue measurement. If not
   ## specified, metrics for all queues are gathered.
   # queues = ["telegraf"]

   ## A list of exchanges to gather as the rabbitmq_exchange measurement. If not
   ## specified, metrics for all exchanges are gathered.
   # exchanges = ["telegraf"]

   ## Queues to include and exclude. Globs accepted.
   ## Note that an empty array for both will include all queues
   queue_name_include = []
   queue_name_exclude = []`

	TelegrafInputs["openntpd"].Sample = `
 # Get standard NTP query metrics from OpenNTPD.
 [[inputs.openntpd]]
   ## Run ntpctl binary with sudo.
   # use_sudo = false

   ## Location of the ntpctl binary.
   # binary = "/usr/sbin/ntpctl"

   ## Maximum time the ntpctl binary is allowed to run.
   # timeout = "5ms"`

	TelegrafInputs["ntpq"].Sample = `
 # Get standard NTP query metrics, requires ntpq executable.
 [[inputs.ntpq]]
   ## If false, set the -n ntpq flag. Can reduce metric gather time.
   dns_lookup = true`

	TelegrafInputs["jenkins"].Sample = `
 # Read jobs and cluster metrics from Jenkins instances
 [[inputs.jenkins]]
   ## The Jenkins URL
   url = "http://my-jenkins-instance:8080"
   # username = "admin"
   # password = "admin"

   ## Set response_timeout
   response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use SSL but skip chain & host verification
   # insecure_skip_verify = false

   ## Optional Max Job Build Age filter
   ## Default 1 hour, ignore builds older than max_build_age
   # max_build_age = "1h"

   ## Optional Sub Job Depth filter
   ## Jenkins can have unlimited layer of sub jobs
   ## This config will limit the layers of pulling, default value 0 means
   ## unlimited pulling until no more sub jobs
   # max_subjob_depth = 0

   ## Optional Sub Job Per Layer
   ## In workflow-multibranch-plugin, each branch will be created as a sub job.
   ## This config will limit to call only the lasted branches in each layer,
   ## empty will use default value 10
   # max_subjob_per_layer = 10

   ## Jobs to exclude from gathering
   # job_exclude = [ "job1", "job2/subjob1/subjob2", "job3/*"]

   ## Nodes to exclude from gathering
   # node_exclude = [ "node1", "node2" ]

   ## Worker pool for jenkins plugin only
   ## Empty this field will use default value 5
   # max_connections = 5`

	TelegrafInputs["ceph"].Sample = `
 # Collects performance metrics from the MON and OSD nodes in a Ceph storage cluster.
 [[inputs.ceph]]
   ## This is the recommended interval to poll.  Too frequent and you will lose
   ## data points due to timeouts during rebalancing and recovery
   interval = '1m'

   ## All configuration values are optional, defaults are shown below

   ## location of ceph binary
   ceph_binary = "/usr/bin/ceph"

   ## directory in which to look for socket files
   socket_dir = "/var/run/ceph"

   ## prefix of MON and OSD socket files, used to determine socket type
   mon_prefix = "ceph-mon"
   osd_prefix = "ceph-osd"

   ## suffix used to identify socket files
   socket_suffix = "asok"

   ## Ceph user to authenticate as
   ceph_user = "client.admin"

   ## Ceph configuration to use to locate the cluster
   ceph_config = "/etc/ceph/ceph.conf"

   ## Whether to gather statistics via the admin socket
   gather_admin_socket_stats = true

   ## Whether to gather statistics via ceph commands
   gather_cluster_stats = false`

	TelegrafInputs["zookeeper"].Sample = `
 # Reads 'mntr' stats from one or many zookeeper servers
 [[inputs.zookeeper]]
   ## An array of address to gather stats about. Specify an ip or hostname
   ## with port. ie localhost:2181, 10.0.0.1:2181, etc.

   ## If no servers are specified, then localhost is used as the host.
   ## If no port is specified, 2181 is used
   servers = [":2181"]

   ## Timeout for metric collections from all servers.  Minimum timeout is "1s".
   # timeout = "5s"

   ## Optional TLS Config
   # enable_tls = true
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## If false, skip chain & host verification
   # insecure_skip_verify = true`

	TelegrafInputs["haproxy"].Sample = `
 # Read metrics of haproxy, via socket or csv stats page
 [[inputs.haproxy]]
   ## An array of address to gather stats about. Specify an ip on hostname
   ## with optional port. ie localhost, 10.10.3.33:1936, etc.
   ## Make sure you specify the complete path to the stats endpoint
   ## including the protocol, ie http://10.10.3.33:1936/haproxy?stats

   ## If no servers are specified, then default to 127.0.0.1:1936/haproxy?stats
   servers = ["http://myhaproxy.com:1936/haproxy?stats"]

   ## Credentials for basic HTTP authentication
   # username = "admin"
   # password = "admin"

   ## You can also use local socket with standard wildcard globbing.
   ## Server address not starting with 'http' will be treated as a possible
   ## socket, so both examples below are valid.
   # servers = ["socket:/run/haproxy/admin.sock", "/run/haproxy/*.sock"]

   ## By default, some of the fields are renamed from what haproxy calls them.
   ## Setting this option to true results in the plugin keeping the original
   ## field names.
   # keep_field_names = false

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs["fluentd"].Sample = `
 # Read metrics exposed by fluentd in_monitor plugin
 [[inputs.fluentd]]
   ## This plugin reads information exposed by fluentd (using /api/plugins.json endpoint).
   ##
   ## Endpoint:
   ## - only one URI is allowed
   ## - https is not supported
   endpoint = "http://localhost:24220/api/plugins.json"

   ## Define which plugins have to be excluded (based on "type" field - e.g. monitor_agent)
   exclude = [
 	  "monitor_agent",
 	  "dummy",
   ]`

	TelegrafInputs["cpu"].Sample = `
[[inputs.cpu]]
  ## Whether to report per-cpu stats or not
  percpu = true
 ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics.
  collect_cpu_time = false
 ## If true, compute and report the sum of all non-idle CPU states.
  report_active = false
	[inputs.cpu.tags]
		host = '{{.Hostname}}'
	`

	TelegrafInputs[`disk`].Sample = `
#Read metrics about disk usage by mount point
[[inputs.disk]]
 ## By default stats will be gathered for all mount points.
 ## Set mount_points will restrict the stats to only the specified mount points.
 # mount_points = ["/"]

 ## Ignore mount points by filesystem type.
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

	[inputs.disk.tags]
	host = '{{.Hostname}}'
	`

	TelegrafInputs[`diskio`].Sample = `
# Read metrics about disk IO by device
[[inputs.diskio]]
  # By default, telegraf will gather stats for all devices including
  # disk partitions.
  # Setting devices will restrict the stats to the specified devices.
   devices = ["sda", "sdb", "vd*"]
  # Uncomment the following line if you need disk serial numbers.
   skip_serial_number = false

  # On systems which support it, device metadata can be added in the form of
  # tags.
  # Currently only Linux is supported via udev properties. You can view
  # available properties for a device by running:
  # 'udevadm info -q property -n /dev/sda'
  # Note: Most, but not all, udev properties can be accessed this way. Properties
  # that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
   device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]

  # Using the same metadata source as device_tags, you can also customize the
  # name of the device via templates.
  # The 'name_templates' parameter is a list of templates to try and apply to
  # the device. The template may contain variables in the form of '$PROPERTY' or
  # '${PROPERTY}'. The first template which does not contain any variables not
  # present for the device is used as the device name tag.
  # The typical use case is for LVM volumes, to get the VG/LV name instead of
  # the near-meaningless DM-0 name.
   name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]

	[inputs.diskio.tags]
	host = '{{.Hostname}}'
	 `

	TelegrafInputs[`kernel`].Sample = `
# Get kernel statistics from /proc/stat
[[inputs.kernel]]
 # no configuration
[inputs.kernel.tags]
host = '{{.Hostname}}'

 # Get kernel statistics from /proc/vmstat
 [[inputs.kernel_vmstat]]
   # no configuration

	[inputs.kernel_vmstat.tags]
	host = '{{.Hostname}}'
	 `

	TelegrafInputs[`mem`].Sample = `
# Read metrics about memory usage
[[inputs.mem]]
  # no configuration

	[inputs.mem.tags]
	host = '{{.Hostname}}'
	`

	TelegrafInputs[`processes`].Sample = `
# Get the number of processes and group them by status
[[inputs.processes]]
 # no configuration

# Monitor process cpu and memory usage
[[inputs.procstat]]
# PID file to monitor process
pid_file = "/var/run/nginx.pid"
# executable name (ie, pgrep <exe>)
 exe = "nginx"
# pattern as argument for pgrep (ie, pgrep -f <pattern>)
 pattern = "nginx"
# user as argument for pgrep (ie, pgrep -u <user>)
 user = "nginx"
# Systemd unit name
 systemd_unit = "nginx.service"
# CGroup name or path
 cgroup = "systemd/system.slice/nginx.service"

# Windows service name
 win_service = ""

# override for process_name
# This is optional; default is sourced from /proc/<pid>/status
 process_name = "bar"

# Field name prefix
 prefix = ""

# When true add the full cmdline as a tag.
 cmdline_tag = false

# Add PID as a tag instead of a field; useful to differentiate between
# processes whose tags are otherwise the same.  Can create a large number
# of series, use judiciously.
 pid_tag = false

# Method to use when finding process IDs.  Can be one of 'pgrep', or
# 'native'.  The pgrep finder calls the pgrep executable in the PATH while
# the native finder performs the search directly in a manor dependent on the
# platform.  Default is 'pgrep'
 pid_finder = "pgrep"`

	TelegrafInputs[`swap`].Sample = `
# Read metrics about swap memory usage
[[inputs.swap]]
  # no configuration

	[inputs.swap.tags]
	host = '{{.Hostname}}'
	`

	TelegrafInputs[`system`].Sample = `
# Read metrics about system load & uptime
[[inputs.system]]
  ## Uncomment to remove deprecated metrics.
  fielddrop = ["uptime_format"]

	[inputs.system.tags]
	host = '{{.Hostname}}'
	`

	TelegrafInputs[`nvidia_smi`].Sample = `
# Pulls statistics from nvidia GPUs attached to the host
[[inputs.nvidia_smi]]
  ## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/usr/bin/nvidia-smi"

  ## Optional: timeout for GPU polling
  # timeout = "5s"`

	TelegrafInputs[`activemq`].Sample = `
 # Gather ActiveMQ metrics
 [[inputs.activemq]]
   ## ActiveMQ WebConsole URL
   url = "http://127.0.0.1:8161"

   ## Required ActiveMQ Endpoint
   ##   deprecated in 1.11; use the url option
   # server = "127.0.0.1"
   # port = 8161

   ## Credentials for basic HTTP authentication
   # username = "admin"
   # password = "admin"

   ## Required ActiveMQ webadmin root path
   # webadmin = "admin"

   ## Maximum time to receive response.
   # response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`dns_query`].Sample = `
 # Query given DNS server and gives statistics
 [[inputs.dns_query]]
   ## servers to query
   servers = ["8.8.8.8"]

   ## Network is the network protocol name.
   # network = "udp"

   ## Domains or subdomains to query.
   # domains = ["."]

   ## Query record type.
   ## Posible values: A, AAAA, CNAME, MX, NS, PTR, TXT, SOA, SPF, SRV.
   # record_type = "A"

   ## Dns server port.
   # port = 53

   ## Query timeout in seconds.
   # timeout = 2`

	TelegrafInputs[`docker`].Sample = `
 # Read metrics about docker containers
 [[inputs.docker]]
   ## Docker Endpoint
   ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
   ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
   endpoint = "unix:///var/run/docker.sock"

   ## Set to true to collect Swarm metrics(desired_replicas, running_replicas)
   gather_services = false

   ## Only collect metrics for these containers, collect all if empty
   container_names = []

   ## Containers to include and exclude. Globs accepted.
   ## Note that an empty array for both will include all containers
   container_name_include = []
   container_name_exclude = []

   ## Container states to include and exclude. Globs accepted.
   ## When empty only containers in the "running" state will be captured.
   ## example: container_state_include = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
   ## example: container_state_exclude = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
   # container_state_include = []
   # container_state_exclude = []

   ## Timeout for docker list, info, and stats commands
   timeout = "5s"

   ## Whether to report for each container per-device blkio (8:0, 8:1...) and
   ## network (eth0, eth1, ...) stats or not
   perdevice = true
   ## Whether to report for each container total blkio and network stats or not
   total = false
   ## Which environment variables should we use as a tag
   ##tag_env = ["JAVA_HOME", "HEAP_SIZE"]

   ## docker labels to include and exclude as tags.  Globs accepted.
   ## Note that an empty array for both will include all labels as tags
   docker_label_include = []
   docker_label_exclude = []

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`docker_log`].Sample = `
 # Read logging output from the Docker engine
 [[inputs.docker_log]]
   ## Docker Endpoint
   ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
   ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
   # endpoint = "unix:///var/run/docker.sock"

   ## When true, container logs are read from the beginning; otherwise
   ## reading begins at the end of the log.
   # from_beginning = false

   ## Timeout for Docker API calls.
   # timeout = "5s"

   ## Containers to include and exclude. Globs accepted.
   ## Note that an empty array for both will include all containers
   # container_name_include = []
   # container_name_exclude = []

   ## Container states to include and exclude. Globs accepted.
   ## When empty only containers in the "running" state will be captured.
   # container_state_include = []
   # container_state_exclude = []

   ## docker labels to include and exclude as tags.  Globs accepted.
   ## Note that an empty array for both will include all labels as tags
   # docker_label_include = []
   # docker_label_exclude = []

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`elasticsearch`].Sample = `
 # Read stats from one or more Elasticsearch servers or clusters
 [[inputs.elasticsearch]]
   ## specify a list of one or more Elasticsearch servers
   # you can add username and password to your url to use basic authentication:
   # servers = ["http://user:pass@localhost:9200"]
   servers = ["http://localhost:9200"]

   ## Timeout for HTTP requests to the elastic search server(s)
   http_timeout = "5s"

   ## When local is true (the default), the node will read only its own stats.
   ## Set local to false when you want to read the node stats from all nodes
   ## of the cluster.
   local = true

   ## Set cluster_health to true when you want to also obtain cluster health stats
   cluster_health = false

   ## Adjust cluster_health_level when you want to also obtain detailed health stats
   ## The options are
   ##  - indices (default)
   ##  - cluster
   # cluster_health_level = "indices"

   ## Set cluster_stats to true when you want to also obtain cluster stats.
   cluster_stats = false

   ## Only gather cluster_stats from the master node. To work this require local = true
   cluster_stats_only_from_master = true

   ## Indices to collect; can be one or more indices names or _all
   indices_include = ["_all"]

   ## One of "shards", "cluster", "indices"
   indices_level = "shards"

   ## node_stats is a list of sub-stats that you want to have gathered. Valid options
   ## are "indices", "os", "process", "jvm", "thread_pool", "fs", "transport", "http",
   ## "breaker". Per default, all stats are gathered.
   # node_stats = ["jvm", "http"]

   ## HTTP Basic Authentication username and password.
   # username = ""
   # password = ""

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`http`].Sample = `
 # Read formatted metrics from one or more HTTP endpoints
 [[inputs.http]]
   ## One or more URLs from which to read formatted metrics
   urls = [
     "http://localhost/metrics"
   ]

   ## HTTP method
   # method = "GET"

   ## Optional HTTP headers
   # headers = {"X-Special-Header" = "Special-Value"}

   ## Optional HTTP Basic Auth Credentials
   # username = "username"
   # password = "pa$$word"

   ## HTTP entity-body to send with POST/PUT requests.
   # body = ""

   ## HTTP Content-Encoding for write request body, can be set to "gzip" to
   ## compress body or "identity" to apply no encoding.
   # content_encoding = "identity"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## Amount of time allowed to complete the HTTP request
   # timeout = "5s"

   ## Data format to consume.
   ## Each data format has its own unique set of configuration options, read
   ## more about them here:
   ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
   # data_format = "influx"

 # HTTP/HTTPS request given an address a method and a timeout
 [[inputs.http_response]]
   ## Deprecated in 1.12, use 'urls'
   ## Server address (default http://localhost)
   # address = "http://localhost"

   ## List of urls to query.
   # urls = ["http://localhost"]

   ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
   # http_proxy = "http://localhost:8888"

   ## Set response_timeout (default 5 seconds)
   # response_timeout = "5s"

   ## HTTP Request Method
   # method = "GET"

   ## Whether to follow redirects from the server (defaults to false)
   # follow_redirects = false

   ## Optional HTTP Request Body
   # body = '''
   # {'fake':'data'}
   # '''

   ## Optional substring or regex match in body of the response
   # response_string_match = "\"service_status\": \"up\""
   # response_string_match = "ok"
   # response_string_match = "\".*_status\".?:.?\"up\""

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## HTTP Request Headers (all values must be strings)
   # [inputs.http_response.headers]
   #   Host = "github.com"

   ## Interface to use when dialing an address
   # interface = "eth0"


 # Read flattened metrics from one or more JSON HTTP endpoints
 [[inputs.httpjson]]
   ## NOTE This plugin only reads numerical measurements, strings and booleans
   ## will be ignored.

   ## Name for the service being polled.  Will be appended to the name of the
   ## measurement e.g. httpjson_webserver_stats
   ##
   ## Deprecated (1.3.0): Use name_override, name_suffix, name_prefix instead.
   name = "webserver_stats"

   ## URL of each server in the service's cluster
   servers = [
     "http://localhost:9999/stats/",
     "http://localhost:9998/stats/",
   ]
   ## Set response_timeout (default 5 seconds)
   response_timeout = "5s"

   ## HTTP method to use: GET or POST (case-sensitive)
   method = "GET"

   ## List of tag names to extract from top-level of JSON server response
   # tag_keys = [
   #   "my_tag_1",
   #   "my_tag_2"
   # ]

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false

   ## HTTP parameters (all values must be strings).  For "GET" requests, data
   ## will be included in the query.  For "POST" requests, data will be included
   ## in the request body as "x-www-form-urlencoded".
   # [inputs.httpjson.parameters]
   #   event_type = "cpu_spike"
   #   threshold = "0.75"

   ## HTTP Headers (all values must be strings)
   # [inputs.httpjson.headers]
   #   X-Auth-Token = "my-xauth-token"
   #   apiVersion = "v1"`

	TelegrafInputs[`iptables`].Sample = `
 # Gather packets and bytes throughput from iptables
 [[inputs.iptables]]
   ## iptables require root access on most systems.
   ## Setting 'use_sudo' to true will make use of sudo to run iptables.
   ## Users must configure sudo to allow telegraf user to run iptables with no password.
   ## iptables can be restricted to only list command "iptables -nvL".
   use_sudo = false
   ## Setting 'use_lock' to true runs iptables with the "-w" option.
   ## Adjust your sudo settings appropriately if using this option ("iptables -w 5 -nvl")
   use_lock = false
   ## Define an alternate executable, such as "ip6tables". Default is "iptables".
   # binary = "ip6tables"
   ## defines the table to monitor:
   table = "filter"
   ## defines the chains to monitor.
   ## NOTE: iptables rules without a comment will not be monitored.
   ## Read the plugin documentation for more information.
   chains = [ "INPUT" ]`

	TelegrafInputs[`redis`].Sample = `
 # Read metrics from one or many redis servers
 [[inputs.redis]]
   ## specify servers via a url matching:
   ##  [protocol://][:password]@address[:port]
   ##  e.g.
   ##    tcp://localhost:6379
   ##    tcp://:password@192.168.99.100
   ##    unix:///var/run/redis.sock
   ##
   ## If no servers are specified, then localhost is used as the host.
   ## If no port is specified, 6379 is used
   servers = ["tcp://localhost:6379"]

   ## specify server password
   # password = "s#cr@t%"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = true`

	TelegrafInputs[`mongodb`].Sample = `
 # Read metrics from one or many MongoDB servers
 [[inputs.mongodb]]
   ## An array of URLs of the form:
   ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
   ## For example:
   ##   mongodb://user:auth_key@10.10.3.30:27017,
   ##   mongodb://10.10.3.33:18832,
   servers = ["mongodb://127.0.0.1:27017"]

   ## When true, collect per database stats
   # gather_perdb_stats = false
   ## When true, collect per collection stats
   # gather_col_stats = false
   ## List of db where collections stats are collected
   ## If empty, all db are concerned
   # col_stats_dbs = ["local"]

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`mysql`].Sample = `
 # Read metrics from one or many mysql servers
 [[inputs.mysql]]
   ## specify servers via a url matching:
   ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
   ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
   ##  e.g.
   ##    servers = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
   ##    servers = ["user@tcp(127.0.0.1:3306)/?tls=false"]
   #
   ## If no servers are specified, then localhost is used as the host.
   servers = ["tcp(127.0.0.1:3306)/"]

   ## Selects the metric output format.
   ##
   ## This option exists to maintain backwards compatibility, if you have
   ## existing metrics do not set or change this value until you are ready to
   ## migrate to the new format.
   ##
   ## If you do not have existing metrics from this plugin set to the latest
   ## version.
   ##
   ## Telegraf >=1.6: metric_version = 2
   ##           <1.6: metric_version = 1 (or unset)
   metric_version = 2

   ## the limits for metrics form perf_events_statements
   perf_events_statements_digest_text_limit  = 120
   perf_events_statements_limit              = 250
   perf_events_statements_time_limit         = 86400
   #
   ## if the list is empty, then metrics are gathered from all databasee tables
   table_schema_databases                    = []
   #
   ## gather metrics from INFORMATION_SCHEMA.TABLES for databases provided above list
   gather_table_schema                       = false
   #
   ## gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST
   gather_process_list                       = true
   #
   ## gather user statistics from INFORMATION_SCHEMA.USER_STATISTICS
   gather_user_statistics                    = true
   #
   ## gather auto_increment columns and max values from information schema
   gather_info_schema_auto_inc               = true
   #
   ## gather metrics from INFORMATION_SCHEMA.INNODB_METRICS
   gather_innodb_metrics                     = true
   #
   ## gather metrics from SHOW SLAVE STATUS command output
   gather_slave_status                       = true
   #
   ## gather metrics from SHOW BINARY LOGS command output
   gather_binary_logs                        = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE
   gather_table_io_waits                     = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS
   gather_table_lock_waits                   = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE
   gather_index_io_waits                     = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS
   gather_event_waits                        = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME
   gather_file_events_stats                  = false
   #
   ## gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST
   gather_perf_events_statements             = false
   #
   ## Some queries we may want to run less often (such as SHOW GLOBAL VARIABLES)
   interval_slow                   = "30m"

   ## Optional TLS Config (will be used if tls=custom parameter specified in server uri)
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`nats`].Sample = `
 # Provides metrics about the state of a NATS server
 [[inputs.nats]]
   ## The address of the monitoring endpoint of the NATS server
   server = "http://localhost:8222"

   ## Maximum time to receive response
   # response_timeout = "5s"`

	TelegrafInputs[`net`].Sample = `
 # Read metrics about network interface usage
 [[inputs.net]]
   ## By default, telegraf gathers stats from any up interface (excluding loopback)
   ## Setting interfaces will tell it to gather these explicit interfaces,
   ## regardless of status.
   ##
   # interfaces = ["eth0"]
   ##
   ## On linux systems telegraf also collects protocol stats.
   ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
   ##
   # ignore_protocol_stats = false`

	TelegrafInputs[`net_response`].Sample = `
 # Collect response time of a TCP or UDP connection
 [[inputs.net_response]]
   ## Protocol, must be "tcp" or "udp"
   ## NOTE: because the "udp" protocol does not respond to requests, it requires
   ## a send/expect string pair (see below).
   protocol = "tcp"
   ## Server address (default localhost)
   address = "localhost:80"

   ## Set timeout
   # timeout = "1s"

   ## Set read timeout (only used if expecting a response)
   # read_timeout = "1s"

   ## The following options are required for UDP checks. For TCP, they are
   ## optional. The plugin will send the given string to the server and then
   ## expect to receive the given 'expect' string back.
   ## string sent to the server
   # send = "ssh"
   ## expected string in answer
   # expect = "ssh"

   ## Uncomment to remove deprecated fields
   # fielddrop = ["result_type", "string_found"]

 # Read TCP metrics such as established, time wait and sockets counts.
 [[inputs.netstat]]
   # no configuration`

	TelegrafInputs[`apache`].Sample = `
 # Read Apache status information (mod_status)
 [[inputs.apache]]
   ## An array of URLs to gather from, must be directed at the machine
   ## readable version of the mod_status page including the auto query string.
   ## Default is "http://localhost/server-status?auto".
   urls = ["http://localhost/server-status?auto"]

   ## Credentials for basic HTTP authentication.
   # username = "myuser"
   # password = "mypassword"

   ## Maximum time to receive response.
   # response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`nginx`].Sample = `
# Read Nginx's basic status information (ngx_http_stub_status_module)
[[inputs.nginx]]
  # An array of Nginx stub_status URI to gather stats.
  urls = ["http://localhost/server_status"]

  ## Optional TLS Config
	#tls_ca = "/etc/telegraf/ca.pem"
  #tls_cert = "/etc/telegraf/cert.cer"
  #tls_key = "/etc/telegraf/key.key"
  ## Use TLS but skip chain & host verification
	#insecure_skip_verify = false

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"

 # Read Nginx Plus' full status information (ngx_http_status_module)
 [[inputs.nginx_plus]]
   ## An array of ngx_http_status_module or status URI to gather stats.
   urls = ["http://localhost/status"]

   # HTTP response timeout (default: 5s)
   response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false


 # Read Nginx Plus Api documentation
 [[inputs.nginx_plus_api]]
   ## An array of API URI to gather stats.
   urls = ["http://localhost/api"]

   # Nginx API version, default: 3
   # api_version = 3

   # HTTP response timeout (default: 5s)
   response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false


 # Read nginx_upstream_check module status information (https://github.com/yaoweibin/nginx_upstream_check_module)
 [[inputs.nginx_upstream_check]]
   ## An URL where Nginx Upstream check module is enabled
   ## It should be set to return a JSON formatted response
   url = "http://127.0.0.1/status?format=json"

   ## HTTP method
   # method = "GET"

   ## Optional HTTP headers
   # headers = {"X-Special-Header" = "Special-Value"}

   ## Override HTTP "Host" header
   # host_header = "check.example.com"

   ## Timeout for HTTP requests
   timeout = "5s"

   ## Optional HTTP Basic Auth credentials
   # username = "username"
   # password = "pa$$word"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false


 # Read Nginx virtual host traffic status module information (nginx-module-vts)
 [[inputs.nginx_vts]]
   ## An array of ngx_http_status_module or status URI to gather stats.
   urls = ["http://localhost/status"]

   ## HTTP response timeout (default: 5s)
   response_timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`postgresql`].Sample = `
 # Read metrics from one or many postgresql servers
 [[inputs.postgresql]]
   ## specify address via a url matching:
   ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
   ##       ?sslmode=[disable|verify-ca|verify-full]
   ## or a simple string:
   ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
   ##
   ## All connection parameters are optional.
   ##
   ## Without the dbname parameter, the driver will default to a database
   ## with the same name as the user. This dbname is just for instantiating a
   ## connection with the server and doesn't restrict the databases we are trying
   ## to grab metrics for.
   ##
   address = "host=localhost user=postgres sslmode=disable"
   ## A custom name for the database that will be used as the "server" tag in the
   ## measurement output. If not specified, a default one generated from
   ## the connection address is used.
   # outputaddress = "db01"

   ## connection configuration.
   ## maxlifetime - specify the maximum lifetime of a connection.
   ## default is forever (0s)
   max_lifetime = "0s"

   ## A  list of databases to explicitly ignore.  If not specified, metrics for all
   ## databases are gathered.  Do NOT use with the 'databases' option.
   # ignored_databases = ["postgres", "template0", "template1"]

   ## A list of databases to pull metrics about. If not specified, metrics for all
   ## databases are gathered.  Do NOT use with the 'ignored_databases' option.
   # databases = ["app_production", "testing"]


 # Read metrics from one or many postgresql servers
 [[inputs.postgresql_extensible]]
   ## specify address via a url matching:
   ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
   ##       ?sslmode=[disable|verify-ca|verify-full]
   ## or a simple string:
   ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
   #
   ## All connection parameters are optional.  #
   ## Without the dbname parameter, the driver will default to a database
   ## with the same name as the user. This dbname is just for instantiating a
   ## connection with the server and doesn't restrict the databases we are trying
   ## to grab metrics for.
   #
   address = "host=localhost user=postgres sslmode=disable"

   ## connection configuration.
   ## maxlifetime - specify the maximum lifetime of a connection.
   ## default is forever (0s)
   max_lifetime = "0s"

   ## A list of databases to pull metrics about. If not specified, metrics for all
   ## databases are gathered.
   ## databases = ["app_production", "testing"]
   #
   ## A custom name for the database that will be used as the "server" tag in the
   ## measurement output. If not specified, a default one generated from
   ## the connection address is used.
   # outputaddress = "db01"
   #
   ## Define the toml config where the sql queries are stored
   ## New queries can be added, if the withdbname is set to true and there is no
   ## databases defined in the 'databases field', the sql query is ended by a
   ## 'is not null' in order to make the query succeed.
   ## Example :
   ## The sqlquery : "SELECT * FROM pg_stat_database where datname" become
   ## "SELECT * FROM pg_stat_database where datname IN ('postgres', 'pgbench')"
   ## because the databases variable was set to ['postgres', 'pgbench' ] and the
   ## withdbname was true. Be careful that if the withdbname is set to false you
   ## don't have to define the where clause (aka with the dbname) the tagvalue
   ## field is used to define custom tags (separated by commas)
   ## The optional "measurement" value can be used to override the default
   ## output measurement name ("postgresql").
   #
   ## Structure :
   ## [[inputs.postgresql_extensible.query]]
   ##   sqlquery string
   ##   version string
   ##   withdbname boolean
   ##   tagvalue string (comma separated)
   ##   measurement string
   [[inputs.postgresql_extensible.query]]
     sqlquery="SELECT * FROM pg_stat_database"
     version=901
     withdbname=false
     tagvalue=""
     measurement=""
   [[inputs.postgresql_extensible.query]]
     sqlquery="SELECT * FROM pg_stat_bgwriter"
     version=901
     withdbname=false
     tagvalue="postgresql.stats"`

	TelegrafInputs[`openldap`].Sample = `
 # OpenLDAP cn=Monitor plugin
 [[inputs.openldap]]
   host = "localhost"
   port = 389

   # ldaps, starttls, or no encryption. default is an empty string, disabling all encryption.
   # note that port will likely need to be changed to 636 for ldaps
   # valid options: "" | "starttls" | "ldaps"
   tls = ""

   # skip peer certificate verification. Default is false.
   insecure_skip_verify = false

   # Path to PEM-encoded Root certificate to use to verify server certificate
   tls_ca = "/etc/ssl/certs.pem"

   # dn/password to bind with. If bind_dn is empty, an anonymous bind is performed.
   bind_dn = ""
   bind_password = ""

   # Reverse metric names so they sort more naturally. Recommended.
   # This defaults to false if unset, but is set to true when generating a new config
   reverse_metric_names = true`

	TelegrafInputs[`kapacitor`].Sample = `
 # Read Kapacitor-formatted JSON metrics from one or more HTTP endpoints
 [[inputs.kapacitor]]
   ## Multiple URLs from which to read Kapacitor-formatted JSON
   ## Default is "http://localhost:9092/kapacitor/v1/debug/vars".
   urls = [
     "http://localhost:9092/kapacitor/v1/debug/vars"
   ]

   ## Time limit for http requests
   timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`phpfpm`].Sample = `
 # Read metrics of phpfpm, via HTTP status page or socket
 [[inputs.phpfpm]]
   ## An array of addresses to gather stats about. Specify an ip or hostname
   ## with optional port and path
   ##
   ## Plugin can be configured in three modes (either can be used):
   ##   - http: the URL must start with http:// or https://, ie:
   ##       "http://localhost/status"
   ##       "http://192.168.130.1/status?full"
   ##
   ##   - unixsocket: path to fpm socket, ie:
   ##       "/var/run/php5-fpm.sock"
   ##      or using a custom fpm status path:
   ##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
   ##
   ##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
   ##       "fcgi://10.0.0.12:9000/status"
   ##       "cgi://10.0.10.12:9001/status"
   ##
   ## Example of multiple gathering from local socket and remote host
   ## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
   urls = ["http://localhost/status"]

   ## Duration allowed to complete HTTP requests.
   # timeout = "5s"

   ## Optional TLS Config
   # tls_ca = "/etc/telegraf/ca.pem"
   # tls_cert = "/etc/telegraf/cert.pem"
   # tls_key = "/etc/telegraf/key.pem"
   ## Use TLS but skip chain & host verification
   # insecure_skip_verify = false`

	TelegrafInputs[`sqlserver`].Sample = `
 # Read metrics from Microsoft SQL Server
 [[inputs.sqlserver]]
   ## Specify instances to monitor with a list of connection strings.
   ## All connection parameters are optional.
   ## By default, the host is localhost, listening on default port, TCP 1433.
   ##   for Windows, the user is the currently running AD user (SSO).
   ##   See https://github.com/denisenkom/go-mssqldb for detailed connection
   ##   parameters.
   servers = [
    "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
   ]

   ## Optional parameter, setting this to 2 will use a new version
   ## of the collection queries that break compatibility with the original
   ## dashboards.
   query_version = 2

   ## If you are using AzureDB, setting this to true will gather resource utilization metrics
   # azuredb = false

   ## If you would like to exclude some of the metrics queries, list them here
   ## Possible choices:
   ## - PerformanceCounters
   ## - WaitStatsCategorized
   ## - DatabaseIO
   ## - DatabaseProperties
   ## - CPUHistory
   ## - DatabaseSize
   ## - DatabaseStats
   ## - MemoryClerk
   ## - VolumeSpace
   ## - PerformanceMetrics
   ## - Schedulers
   ## - AzureDBResourceStats
   ## - AzureDBResourceGovernance
   ## - SqlRequests
   exclude_query = [ 'Schedulers' ]`

	TelegrafInputs[`memcached`].Sample = `
 # Read metrics from one or many memcached servers
 [[inputs.memcached]]
   ## An array of address to gather stats about. Specify an ip on hostname
   ## with optional port. ie localhost, 10.0.0.1:11211, etc.
   servers = ["localhost:11211"]
   # unix_sockets = ["/var/run/memcached.sock"]`

	TelegrafInputs[`ping`].Sample = `
 # Ping given url(s) and return statistics
 [[inputs.ping]]
   ## List of urls to ping
   urls = ["example.org"]

   ## Number of pings to send per collection (ping -c <COUNT>)
   # count = 1

   ## Interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
   # ping_interval = 1.0

   ## Per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
   # timeout = 1.0

   ## Total-ping deadline, in s. 0 == no deadline (ping -w <DEADLINE>)
   # deadline = 10

   ## Interface or source address to send ping from (ping -I[-S] <INTERFACE/SRC_ADDR>)
   # interface = ""

   ## How to ping. "native" doesn't have external dependencies, while "exec" depends on 'ping'.
   # method = "exec"

   ## Specify the ping executable binary, default is "ping"
 	# binary = "ping"

   ## Arguments for ping command. When arguments is not empty, system binary will be used and
   ## other options (ping_interval, timeout, etc) will be ignored.
   # arguments = ["-c", "3"]

   ## Use only ipv6 addresses when resolving hostnames.
   # ipv6 = false`

	TelegrafInputs[`uwsgi`].Sample = `
[[inputs.uwsgi]]
  ## List with urls of uWSGI Stats servers. Url must match pattern:
  ## scheme://address[:port]
  ##
  ## For example:
  ## servers = ["tcp://localhost:5050", "http://localhost:1717", "unix:///tmp/statsock"]
  servers = ["tcp://127.0.0.1:1717"]

  ## General connection timout
  # timeout = "5s"`

	TelegrafInputs[`solr`].Sample = `
 # Read stats from one or more Solr servers or cores
 [[inputs.solr]]
   ## specify a list of one or more Solr servers
   servers = ["http://localhost:8983"]

   ## specify a list of one or more Solr cores (default - all)
   # cores = ["main"]

   ## Optional HTTP Basic Auth Credentials
   # username = "username"
   # password = "pa$$word"`

	TelegrafInputs[`clickhouse`].Sample = `
# Read metrics from one or many ClickHouse servers
[[inputs.clickhouse]]
  ## Username for authorization on ClickHouse server
  ## example: username = "default"
  username = "default"

  ## Password for authorization on ClickHouse server
  ## example: password = "super_secret"

  ## HTTP(s) timeout while getting metrics values
  ## The timeout includes connection time, any redirects, and reading the response body.
  ##   example: timeout = 1s
  # timeout = 5s

  ## List of servers for metrics scraping
  ## metrics scrape via HTTP(s) clickhouse interface
  ## https://clickhouse.tech/docs/en/interfaces/http/
  ##    example: servers = ["http://127.0.0.1:8123","https://custom-server.mdb.yandexcloud.net"]
  servers         = ["http://127.0.0.1:8123"]

  ## If "auto_discovery"" is "true" plugin tries to connect to all servers available in the cluster
  ## with using same "user:password" described in "user" and "password" parameters
  ## and get this server hostname list from "system.clusters" table
  ## see
  ## - https://clickhouse.tech/docs/en/operations/system_tables/#system-clusters
  ## - https://clickhouse.tech/docs/en/operations/server_settings/settings/#server_settings_remote_servers
  ## - https://clickhouse.tech/docs/en/operations/table_engines/distributed/
  ## - https://clickhouse.tech/docs/en/operations/table_engines/replication/#creating-replicated-tables
  ##    example: auto_discovery = false
  # auto_discovery = true

  ## Filter cluster names in "system.clusters" when "auto_discovery" is "true"
  ## when this filter present then "WHERE cluster IN (...)" filter will apply
  ## please use only full cluster names here, regexp and glob filters is not allowed
  ## for "/etc/clickhouse-server/config.d/remote.xml"
  ## <yandex>
  ##  <remote_servers>
  ##    <my-own-cluster>
  ##        <shard>
  ##          <replica><host>clickhouse-ru-1.local</host><port>9000</port></replica>
  ##          <replica><host>clickhouse-ru-2.local</host><port>9000</port></replica>
  ##        </shard>
  ##        <shard>
  ##          <replica><host>clickhouse-eu-1.local</host><port>9000</port></replica>
  ##          <replica><host>clickhouse-eu-2.local</host><port>9000</port></replica>
  ##        </shard>
  ##    </my-onw-cluster>
  ##  </remote_servers>
  ##
  ## </yandex>
  ##
  ## example: cluster_include = ["my-own-cluster"]
  # cluster_include = []

  ## Filter cluster names in "system.clusters" when "auto_discovery" is "true"
  ## when this filter present then "WHERE cluster NOT IN (...)" filter will apply
  ##    example: cluster_exclude = ["my-internal-not-discovered-cluster"]
  # cluster_exclude = []

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false`

	TelegrafInputs[`systemd_units`].Sample = `
[[inputs.systemd_units]]
  # Set timeout for systemctl execution
  timeout = "1s"

  # Filter for a specific unit type, default is "service", other possible
  # values are "socket", "target", "device", "mount", "automount", "swap",
  # "timer", "path", "slice" and "scope ":
  unittype = "service"

	[inputs.systemd_units.tags]
	host = '{{.Hostname}}'
	`

	TelegrafInputs[`influxdb`].Sample = `
# Read InfluxDB-formatted JSON metrics from one or more HTTP endpoints
[[inputs.influxdb]]
	## Works with InfluxDB debug endpoints out of the box,
	## but other services can use this format too.
	## See the influxdb plugin's README for more details.

	## Multiple URLs from which to read InfluxDB-formatted JSON
	## Default is "http://localhost:8086/debug/vars".
	urls = [
		"http://localhost:8086/debug/vars"
	]

	## Username and password to send using HTTP Basic Authentication.
	# username = ""
	# password = ""

	## Optional TLS Config
	# tls_ca = "/etc/telegraf/ca.pem"
	# tls_cert = "/etc/telegraf/cert.pem"
	# tls_key = "/etc/telegraf/key.pem"
	## Use TLS but skip chain & host verification
	# insecure_skip_verify = false

	## http request & header timeout
	timeout = "5s"`

	TelegrafInputs[`consul`].Sample = `
	# Gather health check statuses from services registered in Consul
	[[inputs.consul]]
		# Consul server address
		address = "localhost:8500"

		# URI scheme for the Consul server, one of "http", "https"
		scheme = "http"

		# ACL token used in every request
		token = ""

		# HTTP Basic Authentication username and password.
		username = ""
		password = ""

		# Data center to query the health checks from
		datacenter = ""

		# Optional TLS Config
		#tls_ca = "/etc/telegraf/ca.pem"
		#tls_cert = "/etc/telegraf/cert.pem"
		#tls_key = "/etc/telegraf/key.pem"
		# Use TLS but skip chain & host verification
		#insecure_skip_verify = true

		# Consul checks' tag splitting
		#When tags are formatted like "key:value" with ":" as a delimiter then
		#they will be splitted and reported as proper key:value in Telegraf
		#tag_delimiter = ":"
		`

	TelegrafInputs[`kibana`].Sample = `
		[[inputs.kibana]]
		# Specify a list of one or more Kibana servers
		servers = ["http://localhost:5601"]

		# Timeout for HTTP requests
		timeout = "5s"

		# HTTP Basic Auth credentials
		 username = "username"
		 password = "pa$$word"

		# Optional TLS Config
		#tls_ca = "/etc/telegraf/ca.pem"
		#tls_cert = "/etc/telegraf/cert.pem"
		#tls_key = "/etc/telegraf/key.pem"
		# Use TLS but skip chain & host verification
		#insecure_skip_verify = false
		`

	TelegrafInputs[`modbus`].Sample = `
		[[inputs.modbus]]
		# Connection Configuration
		#
		# The module supports connections to PLCs via MODBUS/TCP or
		# via serial line communication in binary (RTU) or readable (ASCII) encoding
		#
		# Device name
		name = "Device"

		# Slave ID - addresses a MODBUS device on the bus
		# Range: 0 - 255 [0 = broadcast; 248 - 255 = reserved]
		slave_id = 1

		# Timeout for each request
		timeout = "1s"

		# Maximum number of retries and the time to wait between retries
		# when a slave-device is busy.
		 busy_retries = 0
		 busy_retries_wait = "100ms"

		 TCP - connect via Modbus/TCP
		controller = "tcp://localhost:502"

		# Serial (RS485; RS232)
		 controller = "file:///dev/ttyUSB0"
		 baud_rate = 9600
		 data_bits = 8
		 parity = "N"
		 stop_bits = 1
		 transmission_mode = "RTU"


		# Measurements
		#

		# Digital Variables, Discrete Inputs and Coils
		# name    - the variable name
		# address - variable address

		discrete_inputs = [
			{ name = "Start",          address = [0]},
			{ name = "Stop",           address = [1]},
			{ name = "Reset",          address = [2]},
			{ name = "EmergencyStop",  address = [3]},
		]
		coils = [
			{ name = "Motor1-Run",     address = [0]},
			{ name = "Motor1-Jog",     address = [1]},
			{ name = "Motor1-Stop",    address = [2]},
		]

		# Analog Variables, Input Registers and Holding Registers
		# measurement - the (optional) measurement name, defaults to "modbus"
		# name       - the variable name
		# byte_order - the ordering of bytes
		#  |---AB, ABCD   - Big Endian
		#  |---BA, DCBA   - Little Endian
		#  |---BADC       - Mid-Big Endian
		#  |---CDAB       - Mid-Little Endian
		# data_type  - INT16, UINT16, INT32, UINT32, INT64, UINT64, FLOAT32, FLOAT32-IEEE (the IEEE 754 binary representation)
		# scale      - the final numeric variable representation
		# address    - variable address

		holding_registers = [
			{ name = "PowerFactor", byte_order = "AB",   data_type = "FLOAT32", scale=0.01,  address = [8]},
			{ name = "Voltage",     byte_order = "AB",   data_type = "FLOAT32", scale=0.1,   address = [0]},
			{ name = "Energy",      byte_order = "ABCD", data_type = "FLOAT32", scale=0.001, address = [5,6]},
			{ name = "Current",     byte_order = "ABCD", data_type = "FLOAT32", scale=0.001, address = [1,2]},
			{ name = "Frequency",   byte_order = "AB",   data_type = "FLOAT32", scale=0.1,   address = [7]},
			{ name = "Power",       byte_order = "ABCD", data_type = "FLOAT32", scale=0.1,   address = [3,4]},
		]
		input_registers = [
			{ name = "TankLevel",   byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [0]},
			{ name = "TankPH",      byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [1]},
			{ name = "Pump1-Speed", byte_order = "ABCD", data_type = "INT32",   scale=1.0,     address = [3,4]},
		]
		`

	TelegrafInputs[`dotnetclr`].Sample = `
		[[inputs.win_perf_counters]]
		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR Exceptions"

			##(required)
			Counters = ["# of Exceps Thrown / sec"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR Jit"

			##(required)
			Counters = ["% Time in Jit","IL Bytes Jitted / sec"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR Loading"

			##(required)
			Counters = ["% Time Loading"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR LocksAndThreads"

			##(required)
			Counters = ["# of current logical Threads","# of current physical Threads","# of current recognized threads","# of total recognized threads","Queue Length / sec","Total # of Contentions","Current Queue Length"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR Memory"

			##(required)
			Counters = ["% Time in GC","# Bytes in all Heaps","# Gen 0 Collections","# Gen 1 Collections","# Gen 2 Collections","# Induced GC" "Allocated Bytes/sec","Finalization Survivors","Gen 0 heap size","Gen 1 heap size","Gen 2 heap size","Large Object Heap size","# of Pinned Objects"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = ".NET CLR Security"

			##(required)
			Counters = ["% Time in RT checks","Stack Walk Depth","Total Runtime Checks"]

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'dotnetclr'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false
		`

	TelegrafInputs[`aspdotnet`].Sample = `
		[[inputs.win_perf_counters]]
		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'ASP.NET'

			##(required)
			Counters = ['Application Restarts', 'Worker Process Restarts', 'Request Wait Time']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'aspdotnet'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'ASP.NET Applications'

			##(required)
			Counters = ['Requests In Application Queue', 'Requests Executing', 'Requests/Sec', 'Forms Authentication Failure', 'Forms Authentication Success']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'aspdotnet'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false
		`

	TelegrafInputs[`msexchange`].Sample = `
		[[inputs.win_perf_counters]]
		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange ADAccess Domain Controllers'

			##(required)
			Counters = ['LDAP Read Time', 'LDAP Search Time']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			 ##(required)
			ObjectName = 'MSExchange ADAccess Processes'

			 ##(required)
			Counters = ['LDAP Read Time', 'LDAP Search Time']

			 ##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			 ##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'Processor'

			##(required)
			Counters = ['% Processor Time', '% User Time', '% Privileged Time']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['_Total']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=true

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'System'

			##(required)
			Counters = ['Processor Queue Length']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'Memory'

			##(required)
			Counters = ['Available MBytes', '% Committed Bytes In Use']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'Network Interface'

			##(required)
			Counters = ['Packets Outbound Errors']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'TCPv6'

			##(required)
			Counters = ['Connection Failures', 'Connections Reset']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'TCPv4'

			##(required)
			Counters = ['Connections Reset']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'Netlogon'

			##(required)
			Counters = ['Semaphore Waiters', 'Semaphore Holders', 'Semaphore Acquires', 'Semaphore Timeouts', 'Average Semaphore Hold Time']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange Database ==> Instances'

			##(required)
			Counters = ['I/O Database Reads (Attached) Average Latency', 'I/O Database Writes (Attached) Average Latency', 'I/O Log Writes Average Latency', 'I/O Database Reads (Recovery) Average Latency', 'I/O Database Writes (Recovery) Average Latency', 'I/O Database Reads (Attached)/sec', 'I/O Database Writes (Attached)/sec', 'I/O Log Writes/sec']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange RpcClientAccess'

			##(required)
			Counters = ['RPC Averaged Latency', 'RPC Requests', 'Active User Count', 'Connection Count', 'RPC Operations/sec', 'User Count']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange HttpProxy'

			##(required)
			Counters = ['MailboxServerLocator Average Latency', 'Average ClientAccess Server Processing Latency', 'Mailbox Server Proxy Failure Rate', 'Outstanding Proxy Requests', 'Proxy Requests/Sec', 'Requests/Sec']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange ActiveSync'

			##(required)
			Counters = ['Requests/sec', 'Ping Commands Pending', 'Sync Commands/sec']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'Web Service'

			##(required)
			Counters = ['Current Connections', 'Connection Attempts/sec', 'Other Request Methods/sec']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=true

		[[inputs.win_perf_counters.object]]
			##(required)
			ObjectName = 'MSExchange WorkloadManagement Workloads'

			##(required)
			Counters = ['ActiveTasks', 'CompletedTasks', 'QueuedTasks']

			##(required) specify the .net clr instances, use '*' to apply all
			Instances = ['*']

			##(required) all object should use the same name
			Measurement = 'msexchange'

			##(optional)Set to true to include _Total instance when querying for all (*).
			#IncludeTotal=false
			`
}

func init() {
	applyTelegrafSamples()
}
