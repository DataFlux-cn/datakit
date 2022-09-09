// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"database/sql"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	sample = `
[[inputs.sqlserver]]
  ## your sqlserver host ,example ip:port
  host = ""

  ## your sqlserver user,password
  user = ""
  password = ""

  ## (optional) collection interval, default is 10s
  interval = "10s"

  ## by default, support TLS 1.2 and above.
  ## set to true if server side uses TLS 1.0 or TLS 1.1
  allow_tls10 = false

  ## Set true to enable election
  election = true

  ## configure db_filter to filter out metrics from certain databases according to their database_name tag.
  ## If leave blank, no metric from any database is filtered out.
  # db_filter = ["some_db_instance_name", "other_db_instance_name"]

  # [inputs.sqlserver.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "sqlserver.p"

  [inputs.sqlserver.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	pipeline = `
grok(_,"%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\\s+%{GREEDYDATA:msg}")
default_time(time)
`

	inputName    = `sqlserver`
	catalogName  = "db"
	l            = logger.DefaultSLogger(inputName)
	collectCache []*point.Point
	minInterval  = time.Second * 5
	maxInterval  = time.Second * 30
	query        = []string{
		sqlServerPerformanceCounters,
		sqlServerWaitStatsCategorized,
		sqlServerDatabaseIO,
		sqlServerProperties,
		sqlServerSchedulers,
		sqlServerVolumeSpace,
	}
)

type Input struct {
	Host       string            `toml:"host"`
	User       string            `toml:"user"`
	Password   string            `toml:"password"`
	Interval   datakit.Duration  `toml:"interval"`
	Tags       map[string]string `toml:"tags"`
	Log        *sqlserverlog     `toml:"log"`
	AllowTLS10 bool              `toml:"allow_tls10,omitempty"`

	QueryVersionDeprecated int      `toml:"query_version,omitempty"`
	ExcludeQuery           []string `toml:"exclude_query,omitempty"`

	DBFilter    []string `toml:"db_filter,omitempty"`
	dbFilterMap map[string]struct{}

	lastErr error
	tail    *tailer.Tailer
	start   time.Time
	db      *sql.DB

	Election bool `toml:"election"`
	pauseCh  chan bool
	pause    bool

	semStop *cliutils.Sem // start stop signal
}

type sqlserverlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newTimeFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func newByteFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newBoolFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Bool,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}
