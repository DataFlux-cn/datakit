// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"database/sql"
	"time"

	"github.com/coreos/go-semver/semver"
	_ "github.com/lib/pq"
)

var (
	V83  = semver.New("8.3.0")
	V90  = semver.New("9.0.0")
	V91  = semver.New("9.1.0")
	V92  = semver.New("9.2.0")
	V94  = semver.New("9.4.0")
	V96  = semver.New("9.6.0")
	V100 = semver.New("10.0.0")
	V120 = semver.New("12.0.0")
	V130 = semver.New("13.0.0")
	V140 = semver.New("14.0.0")
)

type DB interface {
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type SQLService struct {
	Address     string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
	DB          DB
	Open        func(string, string) (DB, error)
}

func (p *SQLService) Start() (err error) {
	open := p.Open
	if open == nil {
		open = func(dbType, connStr string) (DB, error) {
			db, err := sql.Open(dbType, connStr)
			return db, err
		}
	}
	const localhost = "host=localhost sslmode=disable"

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	if p.DB, err = open("postgres", p.Address); err != nil {
		l.Error("connect error: ", p.Address)
		return err
	}

	p.DB.SetMaxOpenConns(p.MaxIdle)
	p.DB.SetMaxIdleConns(p.MaxIdle)
	p.DB.SetConnMaxLifetime(p.MaxLifetime)

	return nil
}

func (p *SQLService) Stop() error {
	if p.DB != nil {
		if err := p.DB.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}
	}
	return nil
}

func (p *SQLService) Query(query string) (Rows, error) {
	rows, err := p.DB.Query(query)

	if err != nil {
		return nil, err
	} else {
		if err := rows.Err(); err != nil {
			l.Errorf("rows.Err: %s", err)
		}

		return rows, nil
	}
}

func (p *SQLService) SetAddress(address string) {
	p.Address = address
}

func (p *SQLService) GetColumnMap(row scanner, columns []string) (map[string]*interface{}, error) {
	var columnVars []interface{}

	columnMap := make(map[string]*interface{})

	for _, column := range columns {
		columnMap[column] = new(interface{})
	}

	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[columns[i]])
	}

	if err := row.Scan(columnVars...); err != nil {
		return nil, err
	}
	return columnMap, nil
}
