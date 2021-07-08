package mysql

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type tbMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *tbMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *tbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_table_schema",
		Fields: map[string]interface{}{
			// status
			"data_free": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows. Some storage engines, such as MyISAM, store the exact count. For other storage engines, such as InnoDB, this value is an approximation, and may vary from the actual value by as much as 40% to 50%. In such cases, use SELECT COUNT(*) to obtain an accurate count.",
			},
			// status
			"data_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "For InnoDB, DATA_LENGTH is the approximate amount of space allocated for the clustered index, in bytes. Specifically, it is the clustered index size, in pages, multiplied by the InnoDB page size",
			},
			// status
			"index_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "For InnoDB, INDEX_LENGTH is the approximate amount of space allocated for non-clustered indexes, in bytes. Specifically, it is the sum of non-clustered index sizes, in pages, multiplied by the InnoDB page size",
			},
			// status
			"table_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of rows. Some storage engines, such as MyISAM, store the exact count. For other storage engines, such as InnoDB, this value is an approximation, and may vary from the actual value by as much as 40% to 50%. In such cases, use SELECT COUNT(*) to obtain an accurate count.",
			},
		},
		Tags: map[string]interface{}{
			"engine": &inputs.TagInfo{
				Desc: "The storage engine for the table. See The InnoDB Storage Engine, and Alternative Storage Engines.",
			},
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"table_name": &inputs.TagInfo{
				Desc: "The name of the table.",
			},
			"table_schema": &inputs.TagInfo{
				Desc: "The name of the schema (database) to which the table belongs.",
			},
			"table_type": &inputs.TagInfo{
				Desc: "BASE TABLE for a table, VIEW for a view, or SYSTEM VIEW for an INFORMATION_SCHEMA table.",
			},
			"version": &inputs.TagInfo{
				Desc: "The version number of the table's .frm file.",
			},
		},
	}
}

// 数据源获取数据
func (i *Input) getTableSchema() ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	var tableSchemaSql = `
	SELECT
        TABLE_SCHEMA,
        TABLE_NAME,
        TABLE_TYPE,
        ifnull(ENGINE, 'NONE') as ENGINE,
        ifnull(VERSION, '0') as VERSION,
        ifnull(ROW_FORMAT, 'NONE') as ROW_FORMAT,
        ifnull(TABLE_ROWS, '0') as TABLE_ROWS,
        ifnull(DATA_LENGTH, '0') as DATA_LENGTH,
        ifnull(INDEX_LENGTH, '0') as INDEX_LENGTH,
        ifnull(DATA_FREE, '0') as DATA_FREE
    FROM information_schema.tables
    WHERE TABLE_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
	`

	if len(i.Tables) > 0 {
		var arr []string
		for _, table := range i.Tables {
			arr = append(arr, fmt.Sprintf("'%s'", table))
		}

		filterStr := strings.Join(arr, ",")
		tableSchemaSql = fmt.Sprintf("%s and TABLE_NAME in (%s);", tableSchemaSql, filterStr)
	}

	// run query
	l.Info("tableSchema sql,", tableSchemaSql)
	rows, err := i.db.Query(tableSchemaSql)
	if err != nil {
		l.Errorf("query %s error %v", tableSchemaSql, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := &tbMeasurement{
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

		m.name = "mysql_table_schema"

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		var (
			tableSchema string
			tableName   string
			tableType   string
			engine      string
			version     int
			rowFormat   string
			tableRows   int
			dataLength  int
			indexLength int
			dataFree    int
		)

		err = rows.Scan(
			&tableSchema,
			&tableName,
			&tableType,
			&engine,
			&version,
			&rowFormat,
			&tableRows,
			&dataLength,
			&indexLength,
			&dataFree,
		)

		if err != nil {
			return nil, err
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		m.tags["table_schema"] = tableSchema
		m.tags["table_name"] = tableName
		m.tags["table_type"] = tableType
		m.tags["engine"] = engine
		m.tags["version"] = fmt.Sprintf("%d", version)

		m.fields["table_rows"] = tableRows
		m.fields["data_length"] = dataLength
		m.fields["index_length"] = indexLength
		m.fields["data_free"] = dataFree

		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}
