// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Duration struct {
	Duration time.Duration
}

func (d *Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	// see if we can directly convert it
	d.Duration, err = time.ParseDuration(string(b))
	if err == nil {
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err == nil {
			return nil
		}
	}

	// First try parsing as integer seconds
	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}
	// Second try parsing as float seconds
	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sF)
		return nil
	}

	return fmt.Errorf("failed to parse duration %s", b)
}

type JSONFlattener struct {
	Fields map[string]interface{}
}

// FlattenJSON flattens nested maps/interfaces into a fields map (ignoring bools and string).
func (f *JSONFlattener) FlattenJSON(
	fieldname string,
	v interface{},
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	return f.FullFlattenJSON(fieldname, v, false, false)
}

// FullFlattenJSON flattens nested maps/interfaces into a fields map (including bools and string).
func (f *JSONFlattener) FullFlattenJSON(
	fieldname string,
	v interface{},
	convertString bool,
	convertBool bool,
) error {
	if f.Fields == nil {
		f.Fields = make(map[string]interface{})
	}

	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			fieldkey := k
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}

			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, v := range t {
			fieldkey := strconv.Itoa(i)
			if fieldname != "" {
				fieldkey = fieldname + "_" + fieldkey
			}
			err := f.FullFlattenJSON(fieldkey, v, convertString, convertBool)
			if err != nil {
				return err
			}
		}
	case float64:
		f.Fields[fieldname] = t
	case string:
		if convertString {
			if x, ok := v.(string); ok {
				f.Fields[fieldname] = x
			}
		} else {
			return nil
		}
	case bool:
		if convertBool {
			if x, ok := v.(bool); ok {
				f.Fields[fieldname] = x
			}
		} else {
			return nil
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("JSON Flattener: got unexpected type %T with value %v (%s)",
			t, t, fieldname)
	}
	return nil
}

func setHostTagIfNotLoopback(tags map[string]string, u string) {
	uu, err := url.Parse(u)
	if err != nil {
		l.Errorf("parse url: %v", err)
		return
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		l.Errorf("split host and port: %v", err)
		return
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		tags["host"] = host
	}
}
