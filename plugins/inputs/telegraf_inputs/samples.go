package telegraf_inputs

var (
	samples = map[string]string{
		"kube_inventory": `
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
	#insecure_skip_verify = false`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"haproxy": `
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
   # insecure_skip_verify = false`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"phpfpm": `
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
   # insecure_skip_verify = false`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"consul": `
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
				#tag_delimiter = ":"`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"dotnetclr": `
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
			#IncludeTotal=false`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"aspdotnet": `
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
			#IncludeTotal=false`,
		/////////////////////////////////////////////////////////////////////////////////////////
		"msexchange": `
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
			#IncludeTotal=false`,
		/////////////////////////////////////////////////////////////////////////////////////////
	}
)
