// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1 complement for PSpan conversion.
package v1

var (
	ServiceTypes = map[int32]string{
		// Pinpoint Internal (0 ~ 999)
		-1:  "UNDEFINED",
		1:   "UNKNOWN",
		2:   "USER",
		3:   "UNKNOWN_GROUP",
		5:   "TEST",
		7:   "COLLECTOR",
		100: "ASYNC",
		500: "SDK",
		510: "SDK_ASYNC",
		// Server (1000 ~ 1899)
		1000: "STAND_ALONE",
		1005: "TEST_STAND_ALONE",
		1010: "TOMCAT",
		1011: "TOMCAT_METHOD",
		1020: "RESERVED",
		1021: "RESERVED",
		1030: "JETTY",
		1031: "JETTY_METHOD",
		1040: "JBOSS",
		1041: "JBOSS_METHOD",
		1050: "VERTX",
		1051: "VERTX_INTERNAL",
		1052: "VERTX_HTTP_SERVER",
		1053: "VERTX_HTTP_SERVER_INTERNAL",
		1060: "WEBSPHERE",
		1061: "WEBSPHERE_METHOD",
		1070: "WEBLOGIC",
		1071: "WEBLOGIC_METHOD",
		1080: "RESERVED",
		1081: "RESERVED",
		1100: "THRIFT_SERVER",
		1101: "THRIFT_SERVER_INTERNAL",
		1110: "DUBBO_PROVIDER",
		1120: "UNDERTOW",
		1121: "UNDERTOW_METHOD",
		1126: "UNDERTOW_SERVLET_METHOD",
		1130: "GRPC_SERVER",
		1140: "REACTOR_NETTY",
		1141: "REACTOR_NETTY_INTERNAL",
		1300: "C_CPP",
		1301: "C_CPP_METHOD",
		1400: "NODE",
		1401: "NODE_METHOD",
		1500: "PHP",
		1501: "PHP_METHOD",
		1550: "ENVOY",
		1620: "OPENWHISK_INTERNAL",
		1621: "OPENWHISK_CONTROLLER",
		1622: "OPENWHISK_INVOKER",
		1700: "PYTHON",
		1701: "PYTHON_METHOD",
		1702: "CELERY",
		1703: "CELERY-WORKER",
		1800: "GO",
		1801: "GO_FUNCTION",
		// Server Sandbox (1900 ~ 1999)
		// Database (2000 ~ 2899)
		2050: "UNKNOWN_DB",
		2051: "UNKNOWN_DB_EXECUTE_QUERY",
		2100: "MYSQL",
		2101: "MYSQL_EXECUTE_QUERY",
		2102: "R2DBC_MYSQL",
		2103: "R2DBC_MYSQL_EXECUTE_QUERY",
		2150: "MARIADB",
		2151: "MARIADB_EXECUTE_QUERY",
		2152: "R2DBC_MARIADB",
		2153: "R2DBC_MARIADB_EXECUTE_QUERY",
		2200: "MSSQL",
		2201: "MSSQL_EXECUTE_QUERY",
		2250: "MSSQL_JDBC",
		2251: "MSSQL_JDBC_QUERY",
		2252: "R2DBC_MSSQL_JDBC",
		2253: "R2DBC_MSSQL_JDBC_QUERY",
		2300: "ORACLE",
		2301: "ORACLE_EXECUTE_QUERY",
		2302: "R2DBC_ORACLE",
		2303: "R2DBC_ORACLE_EXECUTE_QUERY",
		2400: "CUBRID",
		2401: "CUBRID_EXECUTE_QUERY",
		2410: "NBASET",
		2411: "NBASET_EXECUTE_QUERY",
		2412: "NBASET_INTERNAL",
		2450: "INFORMIX",
		2451: "INFORMIX_EXECUTE_QUERY",
		2500: "POSTGRESQL",
		2501: "POSTGRESQL_EXECUTE_QUERY",
		2502: "R2DBC_POSTGRESQL",
		2503: "R2DBC_POSTGRESQL_EXECUTE_QUERY",
		2600: "CASSANDRA",
		2601: "CASSANDRA_EXECUTE_QUERY",
		2650: "MONGO",
		2651: "MONGO_EXECUTE_QUERY",
		2652: "MONGO_REACTIVE",
		2700: "COUCHDB",
		2701: "COUCHDB_EXECUTE_QUERY",
		2750: "H2",
		2751: "H2_EXECUTE_QUERY",
		2752: "R2DBC_H2",
		2753: "R2DBC_H2_EXECUTE_QUERY",
		// Database Sandbox (2900 ~ 2999)
		// RESERVED (3000 ~ 4999)
		// Library (5000 ~ 7499)
		5000: "INTERNAL_METHOD",
		5005: "JSP",
		5010: "GSON",
		5011: "JACKSON",
		5012: "JSON-LIB",
		5013: "FASTJSON",
		5020: "JDK_FUTURE",
		5050: "SPRING",
		5051: "SPRING_MVC",
		5052: "SPRING_ASYNC",
		5053: "SPRING_WEBFLUX",
		5061: "RESERVED",
		5071: "SPRING_BEAN",
		5500: "IBATIS",
		5501: "IBATIS-SPRING",
		5510: "MYBATIS",
		6001: "THREAD_ASYNC",
		6005: "PROCESS",
		6050: "DBCP",
		6052: "DBCP2",
		6060: "HIKARICP",
		6062: "DRUID",
		6500: "RXJAVA",
		6510: "REACTOR",
		6600: "EXPRESS",
		6610: "KOA",
		6620: "HAPI",
		6630: "RESTIFY",
		6640: "SPRING_DATA_R2DBC",
		7010: "USER_INCLUDE",
		// Library Sandbox (7500 ~ 7999)
		// Cache & File Library (8000 ~ 8899) Fast Histogram
		8050: "MEMCACHED",
		8051: "MEMCACHED_FUTURE_GET",
		8100: "ARCUS",
		8101: "ARCUS_FUTURE_GET",
		8102: "ARCUS_EHCACHE_FUTURE_GET",
		8103: "ARCUS_INTERNAL",
		8200: "REDIS",
		8201: "REDIS_LETTUCE",
		8202: "IOREDIS",
		8203: "REDIS_REDISSON",
		8204: "REDIS_REDISSON_INTERNAL",
		8250: "RESERVED",
		8251: "RESERVED",
		8260: "RESERVED",
		8280: "ETCD",
		8300: "RABBITMQ",
		8310: "ACTIVEMQ_CLIENT",
		8311: "ACTIVEMQ_CLIENT_INTERNAL",
		8660: "KAFKA_CLIENT",
		8661: "KAFKA_CLIENT_INTERNAL",
		8800: "HBASE_CLIENT",
		8801: "HBASE_CLIENT_ADMIN",
		8802: "HBASE_CLIENT_TABLE",
		8803: "HBASE_ASYNC_CLIENT",
		// Cache Library Sandbox (8900 ~ 8999) Histogram type: Fast
		// RPC (9000 ~ 9899)
		9050: "HTTP_CLIENT_3",
		9051: "HTTP_CLIENT_3_INTERNAL",
		9052: "HTTP_CLIENT_4",
		9053: "HTTP_CLIENT_4_INTERNAL",
		9054: "GOOGLE_HTTP_CLIENT_INTERNAL",
		9055: "JDK_HTTPURLCONNECTOR",
		9056: "ASYNC_HTTP_CLIENT",
		9057: "ASYNC_HTTP_CLIENT_INTERNAL",
		9058: "OK_HTTP_CLIENT",
		9059: "OK_HTTP_CLIENT_INTERNAL",
		9060: "RESERVED",
		9070: "RESERVED",
		9080: "APACHE_CXF_CLIENT",
		9081: "APACHE_CXF_SERVICE_INVOKER",
		9082: "APACHE_CXF_MESSAGE_SENDER",
		9083: "APACHE_CXF_LOGGING_IN",
		9084: "APACHE_CXF_LOGGING_OUT",
		9100: "THRIFT_CLIENT",
		9101: "THRIFT_CLIENT_INTERNAL",
		9110: "DUBBO_CONSUMER",
		9120: "HYSTRIX_COMMAND",
		9130: "VERTX_HTTP_CLIENT",
		9131: "VERTX_HTTP_CLIENT_INTERNAL",
		9140: "REST_TEMPLATE",
		9150: "NETTY",
		9151: "NETTY_INTERNAL",
		9152: "NETTY_HTTP",
		9153: "SPRING_WEBFLUX_CLIENT",
		9154: "REACTOR_NETTY_CLIENT",
		9155: "REACTOR_NETTY_CLIENT_INTERNAL",
		9160: "GRPC",
		9161: "GRPC_INTERNAL",
		9162: "GRPC_SERVER_INTERNAL",
		9201: "ElasticsearchBBoss @Deprecated",
		9202: "ElasticsearchBBossExecutor @Deprecated",
		9203: "ELASTICSEARCH",
		9204: "ELASTICSEARCH_HIGHLEVEL_CLIENT",
		9205: "ELASTICSEARCH8",
		9206: "ELASTICSEARCH8_CLIENT",
		9301: "ENVOY_INGRESS",
		9302: "ENVOY_EGRESS",
		9401: "GO_HTTP_CLIENT",
		9622: "OPENWHISK_CLIENT",
		9700: "PHP_REMOTE_METHOD",
		9800: "C_CPP_REMOTE_METHOD",
		9900: "PYTHON_REMOTE_METHOD",
		// RPC Sandbox (9900 ~ 9999)
	}
	Annotations = map[int32]string{
		-1:       "args[0]",
		-2:       "args[1]",
		-3:       "args[2]",
		-4:       "args[3]",
		-5:       "args[4]",
		-6:       "args[5]",
		-7:       "args[6]",
		-8:       "args[7]",
		-9:       "args[8]",
		-10:      "args[9]",
		-11:      "args[N]",
		-30:      "cached_args[0]",
		-31:      "cached_args[1]",
		-32:      "cached_args[2]",
		-33:      "cached_args[3]",
		-34:      "cached_args[4]",
		-35:      "cached_args[5]",
		-36:      "cached_args[6]",
		-37:      "cached_args[7]",
		-38:      "cached_args[8]",
		-39:      "cached_args[9]",
		-40:      "cached_args[N]",
		-50:      "Exception",
		-51:      "ExceptionClass",
		-100:     "Asynchronous Invocation",
		-9999:    "UNKNOWN",
		12:       "API",
		13:       "API_METADATA",
		14:       "RETURN_DATA",
		15:       "RESERVED",
		16:       "RESERVED",
		17:       "RESERVED",
		20:       "SQL-ID",
		21:       "SQL",
		22:       "SQL-METADATA",
		23:       "SQL-PARAM",
		24:       "SQL-BindValue",
		30:       "STRING_ID",
		40:       "http.url",
		41:       "http.param",
		42:       "http.entity",
		45:       "http.cookie",
		46:       "http.status.code",
		48:       "http.internal.display",
		49:       "http.io",
		50:       "arcus.command",
		55:       "http.resp.header",
		60:       "RESERVED",
		61:       "RESERVED",
		62:       "RESERVED",
		70:       "RESERVED",
		71:       "RESERVED",
		72:       "RESERVED",
		73:       "RESERVED",
		80:       "thrift.url",
		81:       "thrift.args",
		82:       "thrift.result",
		90:       "dubbo.args",
		91:       "dubbo.result",
		100:      "message.queue.url",
		110:      "hystrix.command",
		111:      "hystrix.command.execution",
		112:      "hystrix.command.fallback.cause",
		113:      "hystrix.command.exception",
		115:      "hystrix.command.key",
		116:      "hystrix.command.group.key",
		117:      "hystrix.thread.pool.key",
		118:      "hystrix.collapser.key",
		120:      "netty.address",
		130:      "rabbitmq.properties",
		131:      "rabbitmq.body",
		132:      "rabbitmq.properties",
		133:      "rabbitmq.body",
		140:      "kafka.topic",
		141:      "kafka.partition",
		142:      "kafka.offset",
		150:      "mongo.json.data",
		151:      "mongo.collection.info",
		152:      "mongo.collection.option",
		153:      "mongo.json",
		154:      "mongo.json.bindvalue",
		160:      "grpc.status",
		171:      "es.args",
		172:      "es.url",
		173:      "es.dsl",
		174:      "es.action",
		175:      "es.responseHandle",
		176:      "es.version",
		177:      "es.dsl",
		178:      "es.version",
		180:      "process.command",
		181:      "process.pid",
		200:      "cxf.operation",
		201:      "cxf.args",
		203:      "cxf.address",
		204:      "cxf.response.code",
		205:      "cxf.encoding",
		206:      "cxf.http.method",
		207:      "cxf.content.type",
		208:      "cxf.headers",
		209:      "cxf.messages",
		210:      "cxf.payload",
		300:      "PROXY_HTTP_HEADER",
		310:      "REDIS.IO",
		320:      "hbase.client.params",
		321:      "hbase.table.name",
		330:      "memory.usage",
		923:      "marker.message",
		9000:     "gson.json.length",
		9001:     "jackson.json.length",
		9002:     "json-lib.json.length",
		9003:     "fastjson.json.length",
		10015:    "API-TAG",
		10000010: "API-METADATA-ERROR",
		10000011: "API-METADATA-AGENT-INFO-NOT-FOUND",
		10000012: "API-METADATA-IDENTIFIER-CHECK_ERROR",
		10000013: "API-METADATA-NOT-FOUND",
		10000014: "API-METADATA-DID-COLLSION",
	}
)

func getServiceType(serviceType int32) string {
	service, ok := ServiceTypes[serviceType]
	if ok {
		return service
	} else {
		return "Unknown"
	}
}