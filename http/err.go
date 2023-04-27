// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"errors"
	"net/http"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
)

var OK = newErr(nil, http.StatusOK)

var (
	ErrBadReq = newErr(errors.New("bad request"), http.StatusBadRequest)

	ErrInvalidRequest  = newErr(errors.New("invalid request"), http.StatusBadRequest)
	ErrInvalidCategory = newErr(errors.New("invalid category"), http.StatusBadRequest)
	ErrInvalidPipeline = newErr(errors.New("invalid pipeline"), http.StatusBadRequest)
	ErrInvalidData     = newErr(errors.New("invalid data"), http.StatusBadRequest)
	ErrCompiledFailed  = newErr(errors.New("pipeline compile failed"), http.StatusBadRequest)

	ErrInvalidPrecision       = newErr(errors.New("invalid precision"), http.StatusBadRequest)
	ErrHTTPReadErr            = newErr(errors.New("HTTP read error"), http.StatusInternalServerError)
	ErrEmptyBody              = newErr(errors.New("empty body"), http.StatusBadRequest)
	ErrNoPoints               = newErr(errors.New("no points"), http.StatusBadRequest)
	ErrReloadDatakitFailed    = newErr(errors.New("reload datakit failed"), http.StatusInternalServerError)
	ErrUploadFileErr          = newErr(errors.New("upload file failed"), http.StatusInternalServerError)
	ErrInvalidToken           = newErr(errors.New("invalid token"), http.StatusForbidden)
	ErrUnknownRUMMeasurement  = newErr(errors.New("unknown RUM measurement"), http.StatusBadRequest)
	ErrRUMAppIDNotInWhiteList = newErr(errors.New("RUM app_id not in the white list"), http.StatusForbidden)
	ErrInvalidAPIHandler      = newErr(errors.New("invalid API handler"), http.StatusInternalServerError)
	ErrPublicAccessDisabled   = newErr(errors.New("public access disabled"), http.StatusForbidden)
	ErrReachLimit             = newErr(errors.New("reach max API limit"), http.StatusTooManyRequests)

	// write body error.
	ErrInvalidJSONPoint = newErr(errors.New("invalid json point"), http.StatusBadRequest)
	ErrInvalidLinePoint = newErr(errors.New("invalid line point"), http.StatusBadRequest)
)

func newErr(err error, code int) *uhttp.HttpError {
	return uhttp.NewNamespaceErr(err, code, "datakit")
}
