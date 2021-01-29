package mysqlmonitor

import (
	"database/sql"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	_ "github.com/go-sql-driver/mysql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	defaultTimeout                             = 5 * time.Second
	defaultPerfEventsStatementsDigestTextLimit = 120
	defaultPerfEventsStatementsLimit           = 250
	defaultPerfEventsStatementsTimeLimit       = 86400
	defaultGatherGlobalVars                    = true
)

var (
	l             *logger.Logger
	name          = "mysqlMonitor"
	mariaDBMetric = "mariaDBMonitor"
)

func (_ *MysqlMonitor) Catalog() string {
	return "db"
}

func (_ *MysqlMonitor) SampleConfig() string {
	return configSample
}

func (m *MysqlMonitor) Run() {
	l = logger.SLogger("mysqlMonitor")
	l.Info("mysqlMonitor input started...")

	m.checkCfg()

	tick := time.NewTicker(m.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			m.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (m *MysqlMonitor) checkCfg() {
	// 采集频度
	m.IntervalDuration = 10 * time.Minute

	if m.Interval != "" {
		du, err := time.ParseDuration(m.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.Interval, err.Error())
		} else {
			m.IntervalDuration = du
		}
	}

	// 指标集名称
	if m.MetricName == "" {
		m.MetricName = name
	}

	if m.Product == "MariaDB" {
		m.MetricName = mariaDBMetric
	}
}

func (m *MysqlMonitor) handle() {
	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, server := range m.Servers {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			m.gatherServer(s)
		}(server)
	}

	wg.Wait()
}

func (m *MysqlMonitor) gatherServer(serv string) error {
	serv, err := dsnAddTimeout(serv)
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", serv)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
		return err
	}

	defer db.Close()

	if m.GatherGlobalStatus {
		err = m.gatherGlobalStatuses(db, serv)
		if err != nil {
			l.Warnf("gatherGlobalStatuses error, %v", err)
		}
	}

	if m.GatherGlobalVars {
		// Global Variables may be gathered less often
		err = m.gatherGlobalVariables(db, serv)
		if err != nil {
			l.Warnf("gatherGlobalVariables error, %v", err)
		}
	}

	if m.GatherBinaryLogs {
		err = m.gatherBinaryLogs(db, serv)
		if err != nil {
			l.Warnf("gatherBinaryLogs error, %v", err)
		}
	}

	if m.GatherProcessList {
		err = m.GatherProcessListStatuses(db, serv)
		if err != nil {
			l.Warnf("GatherProcessListStatuses error, %v", err)
		}
	}

	if m.GatherUserStatistics {
		err = m.GatherUserStatisticsStatuses(db, serv)
		if err != nil {
			l.Warnf("gatherUserStatisticsStatuses error, %v", err)
		}
	}

	if m.GatherSlaveStatus {
		err = m.gatherSlaveStatuses(db, serv)
		if err != nil {
			l.Warnf("gatherSlaveStatuses error, %v", err)
		}
	}

	if m.GatherInfoSchemaAutoInc {
		err = m.gatherInfoSchemaAutoIncStatuses(db, serv)
		if err != nil {
			l.Warnf("gatherInfoSchemaAutoIncStatuses error, %v", err)
		}
	}

	if m.GatherInnoDBMetrics {
		err = m.gatherInnoDBMetrics(db, serv)
		if err != nil {
			l.Warnf("gatherInnoDBMetrics error, %v", err)
		}
	}

	if m.GatherTableIOWaits {
		err = m.gatherPerfTableIOWaits(db, serv)
		if err != nil {
			l.Warnf("gatherPerfTableIOWaits error, %v", err)
		}
	}

	if m.GatherIndexIOWaits {
		err = m.gatherPerfIndexIOWaits(db, serv)
		if err != nil {
			l.Warnf("gatherPerfIndexIOWaits error, %v", err)
		}
	}

	if m.GatherTableLockWaits {
		err = m.gatherPerfTableLockWaits(db, serv)
		if err != nil {
			l.Warnf("gatherPerfTableLockWaits error, %v", err)
		}
	}

	if m.GatherEventWaits {
		err = m.gatherPerfEventWaits(db, serv)
		if err != nil {
			l.Warnf("gatherPerfEventWaits error, %v", err)
		}
	}

	if m.GatherFileEventsStats {
		err = m.gatherPerfFileEventsStatuses(db, serv)
		if err != nil {
			l.Warnf("gatherPerfFileEventsStatuses error, %v", err)
		}
	}

	if m.GatherPerfEventsStatements {
		err = m.gatherPerfEventsStatements(db, serv)
		if err != nil {
			l.Warnf("gatherPerfEventsStatements error, %v", err)
		}
	}

	if m.GatherTableSchema {
		err = m.gatherTableSchema(db, serv)
		if err != nil {
			l.Warnf("gatherTableSchema error, %v", err)
		}
	}
	return nil
}

func (m *MysqlMonitor) Test() (*inputs.TestResult, error) {
	m.test = true
	m.resData = nil

	m.handle()

	res := &inputs.TestResult{
		Result: m.resData,
		Desc:   "success!",
	}

	return res, nil
}

func init() {
	inputs.Add(name, func() inputs.Input {
		return &MysqlMonitor{}
	})
}
