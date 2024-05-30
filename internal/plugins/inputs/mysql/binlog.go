// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"strconv"
)

func binlogMetrics(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	res := map[string]interface{}{}
	defer closeRows(r)

	var usage uint64

	for r.Next() {
		var key string
		var val sql.RawBytes
		if n, err := r.Columns(); err != nil {
			l.Warnf("Columns(): %s, ignored", err.Error())
			continue
		} else {
			length := len(n)
			switch length {
			case 3:
				var encrypted string
				if err := r.Scan(&key, &val, &encrypted); err != nil {
					l.Warnf("Scan(): %s, ignored", err.Error())
					continue
				}
			default: // 2
				if err := r.Scan(&key, &val); err != nil {
					l.Warnf("Scan(): %s, ignored", err.Error())
					continue
				}
			}
		}

		raw := string(val)

		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			usage += v
		} else {
			l.Warnf("invalid binlog usage: (%s: %s), ignored", key, raw)
		}
	}

	res["Binlog_space_usage_bytes"] = usage
	return res
}
