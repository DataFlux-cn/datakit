package influxdb

const (
	sampleConfig = `
[[inputs.influxdb]]
  ## (optional) collect interval, default is 10 seconds
  interval = '10s'
  
  url = "http://localhost:8086/debug/vars"

  ## Username and password to send using HTTP Basic Authentication.
  # username = ""
  # password = ""

  ## http request & header timeout
  timeout = "5s"

  ## (Optional) TLS connection config
  # [inputs.influxdb.tlsconf]
    # ca_certs = ["/path/to/ca.pem"]
    # cert = "/path/to/cert.pem"
    # cert_key = "/path/to/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

  [inputs.influxdb.log]
    files = []
    ## grok pipeline script path
    pipeline = "influxdb.p"

  [inputs.influxdb.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
	pipelineCfg = ``
)
