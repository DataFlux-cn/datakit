package httpstat

const (
	pluginName = "httpstat"

	description = `stat http protocol request time, contain dnsLookup, tcpConnection, tlsHandshake,
	serverProcessing, contentTransfer, and total time`
	httpstatConfigSample = `
#    [[inputs.httpstat]]
#    ##if empty, use "httpstat"
#    metricName = ''
#    timeout = ''
#    ## default is 10s
#    interval = '10s'
#    [[httpstat.action]]
#    url = ""
#    method = ""
#    playload = ""
#    kAlive = true
#    tlsSkipVerify = true 
#    compress = true
`
)

type Httpstat struct {
	MetricName string    `toml:"metricName"`
	Timeout    string    `toml:"timeout"`
	Interval   string    `toml:"interval"`
	Actions    []*Action `toml:"action"`
	httpPing   []*httpPing
}

type Action struct {
	Url           string `toml:"url"`
	Method        string `toml:"method"`
	Playload      string `toml:"playload"`
	KAlive        bool   `toml:"kAlive"`
	TLSSkipVerify bool   `toml:"tlsSkipVerify`
	Compress      bool   `toml:"compress`
}
