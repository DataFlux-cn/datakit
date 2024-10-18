// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// DialtestingSender used for dialtesting collector.
type DialtestingSender struct {
	ep *endPoint
}

type DialtestingSenderOpt struct {
	HTTPTimeout time.Duration
}

func (d *DialtestingSender) Init(opt *DialtestingSenderOpt) error {
	d.ep = &endPoint{}
	if opt != nil {
		withHTTPTimeout(opt.HTTPTimeout)(d.ep)
	}
	return d.ep.setupHTTP()
}

func (d *DialtestingSender) WriteData(url string, pts []*point.Point) error {
	// TODO: can not set content encoding here, default use line-protocol

	// return write error or build error
	var writeError error
	w := getWriter(WithPoints(pts),
		WithDynamicURL(url),
		WithCategory(point.DynamicDWCategory),
		WithBodyCallback(func(w *writer, b *body) error {
			err := d.ep.writePointData(w, b)
			if err != nil {
				writeError = err
			}

			return err
		}),
		WithHTTPHeader("X-Sub-Category", "dialtesting"))
	defer putWriter(w)

	if d.ep == nil {
		return fmt.Errorf("endpoint is not set correctly")
	}

	buildErr := w.buildPointsBody()

	if buildErr != nil {
		return buildErr
	}

	return writeError
}

// CheckToken checks if token is valid based on the specified scheme and host.
func (d *DialtestingSender) CheckToken(token, scheme, host string) (bool, error) {
	if d.ep == nil {
		return false, fmt.Errorf("no endpoint available")
	}

	reqURL := fmt.Sprintf("%s://%s%s/%s", scheme, host, datakit.TokenCheck, token)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := d.ep.sendReq(req)
	if err != nil {
		return false, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return false, err
	}

	defer resp.Body.Close() //nolint:errcheck

	result := checkTokenResult{}

	if resp.StatusCode == 200 {
		return true, nil
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("invalid JSON body content(%s): %w", body, err)
	}

	if result.Code == 200 || len(result.ErrorCode) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}
