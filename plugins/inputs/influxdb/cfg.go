// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package influxdb

const sampleConfig = `
[[inputs.influxdb]]
  url = "http://localhost:8086/debug/vars"

  ## (optional) collect interval, default is 10 seconds
  interval = '10s'
  
  ## Username and password to send using HTTP Basic Authentication.
  # username = ""
  # password = ""

  ## http request & header timeout
  timeout = "5s"

  ## Set true to enable election
  election = true

  ## (Optional) TLS connection config
  # [inputs.influxdb.tlsconf]
  # ca_certs = ["/path/to/ca.pem"]
  # cert = "/path/to/cert.pem"
  # cert_key = "/path/to/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  # [inputs.influxdb.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "influxdb.p"

  [inputs.influxdb.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
