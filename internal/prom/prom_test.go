// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

const promURL = "http://127.0.0.1:9100/metrics"

const caContent = `-----BEGIN CERTIFICATE-----
MIICqDCCAZACCQC27UZHg8A/CjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAt0
b255YmFpLmNvbTAeFw0yMTExMjUwMTU3MzBaFw0zNTA4MDQwMTU3MzBaMBYxFDAS
BgNVBAMMC3RvbnliYWkuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAozWNMKEeVVKRg5QuOPv9bmuGOShRWaMxmyLnfzvV5tS/Odg63jEecE3K/HHa
OUrTwHKl2NSfwfUZPfCf1gVYHBzozX66XXXYR+qV2aeg+GsMg+o8foH8mmBL9cW+
fvbpNNv9k9G4W0zX9YdWmXt8KHKr5KThSUq46KN8qUCUPqIBnPMKfDJuEjLMPuxi
hliehoeHY32YcglLKSSAYMos2SWUA/D81wyfZKeH8KQu7lPKdCEXKLJBS4+HxUFx
gwDV84m+H8v9bf8PIeKrUGmzSuCYUCxrQyoiaIawB6iY7BJAQeaqEr+W6bS9BU2f
p6KHG8yEHDfz3gFuNR3vCLzGDQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQA8DiJy
o6BHVTumAbBv9+Q0FsXKQRH1YVwR7spHLdbzbqJadghTGPrXGzwYBaGiLTHHaLXX
Ksbdc8T/C7pRIVXc9Jbx1EzCQFlaDBk89okAG/cWcbr0P5sMDJ96UrapBo2PYKNq
QvSQhSjvKTVB19wwSoD7zbOqITXWQKcv1d10yd3X5Q2PComjMMuhWKAOtJvIvEru
/3WiDpYgGGz/XN1YRFnNvRsXEVa6T0Q7lOi/7Lfv+N96R643Zv5fcyAFMGIiQ0na
vfqe/FB05Gl89x+Bb7xti8bzAlsFy1byeIfFKU3Gmvb8INRJyH5wRWVu29poXl1N
g/pAjggcs8zy5GxR
-----END CERTIFICATE-----`

type transportMock struct {
	statusCode int
	body       string
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{statusCode: http.StatusOK, body: body}
}

func TestCollect(t *testing.T) {
	testcases := []struct {
		in     *Option
		name   string
		fail   bool
		expect []string
	}{
		{
			name: "nil option",
			fail: true,
		},
		{
			name: "empty option",
			in:   &Option{},
			fail: true,
		},
		{
			name: "ok",
			expect: []string{
				`gogo gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`gogo,quantile=0 gc_duration_seconds=0`,
				`gogo,quantile=0.25 gc_duration_seconds=0`,
				`gogo,quantile=0.5 gc_duration_seconds=0`,
			},
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{},
				Measurements: []Rule{
					{
						Pattern: `^go_.*`,
						Name:    "gogo",
						Prefix:  "go_",
					},
				},
				MetricNameFilter: []string{"go"},
			},
		},

		{
			name: "option-only-URL",
			in:   &Option{URL: promURL},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				"http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=+Inf,method=GET,status_code=403 request_duration_seconds_bucket=1i",
				`http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				"http,method=GET,status_code=403 request_duration_seconds_count=0,request_duration_seconds_sum=0",
				`promhttp metric_handler_requests_in_flight=1`,
				`promhttp,cause=encoding metric_handler_errors_total=0`,
				`promhttp,cause=gathering metric_handler_errors_total=0`,
				`promhttp,code=200 metric_handler_requests_total=15143`,
				`promhttp,code=500 metric_handler_requests_total=0`,
				`promhttp,code=503 metric_handler_requests_total=0`,
				`up up=1`,
			},
		},

		{
			name: "option-ignore-tag-kv",
			in: &Option{
				URL: promURL,
				IgnoreTagKV: IgnoreTagKeyValMatch{
					"le":          []*regexp.Regexp{regexp.MustCompile("0.*")},
					"status_code": []*regexp.Regexp{regexp.MustCompile("403")},
				},
			},
			expect: []string{
				`go gc_duration_seconds_count=0,gc_duration_seconds_sum=0`,
				`go,quantile=0 gc_duration_seconds=0`,
				`go,quantile=0.25 gc_duration_seconds=0`,
				`go,quantile=0.5 gc_duration_seconds=0`,
				`http,le=1.2,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i`,
				`http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013`,
				`promhttp metric_handler_requests_in_flight=1`,
				`promhttp,cause=encoding metric_handler_errors_total=0`,
				`promhttp,cause=gathering metric_handler_errors_total=0`,
				`promhttp,code=200 metric_handler_requests_total=15143`,
				`promhttp,code=500 metric_handler_requests_total=0`,
				`promhttp,code=503 metric_handler_requests_total=0`,
				`up up=1`,
			},
		},
	}

	mockBody := `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.03",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.1",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.3",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.5",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="10",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.2",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="403",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true

			pts, err := p.CollectFromHTTP(p.opt.URL)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var arr []string
			for _, pt := range pts {
				arr = append(arr, pt.String())
			}

			sort.Strings(arr)
			sort.Strings(tc.expect)

			for i := range arr {
				assert.Equal(t, strings.HasPrefix(arr[i], tc.expect[i]), true)
				t.Logf(">>>\n%s\n%s", arr[i], tc.expect[i])
			}
		})
	}
}

func Test_BearerToken(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck
	f, err := ioutil.TempFile(tmpDir, "token")
	assert.NoError(t, err)
	token_file := f.Name()
	defer os.Remove(token_file) // nolint:errcheck
	testcases := []struct {
		auth    map[string]string
		url     string
		isError bool
	}{
		{
			auth:    map[string]string{},
			url:     "http://localhost",
			isError: true,
		},
		{
			auth:    map[string]string{"token": "xxxxxxxxxx"},
			url:     "http://localhost",
			isError: false,
		},
		{
			auth:    map[string]string{"token_file": "invalid_file"},
			url:     "http://localhost",
			isError: true,
		},
		{
			auth:    map[string]string{"token_file": token_file},
			url:     "http://localhost",
			isError: false,
		},
	}

	for _, tc := range testcases {
		r, err := BearerToken(tc.auth, tc.url)

		if tc.isError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			authHeader, ok := r.Header["Authorization"]
			assert.True(t, ok)
			assert.Equal(t, len(authHeader), 1)
			assert.Contains(t, authHeader[0], "Bearer")
		}
	}
}

func Test_Tls(t *testing.T) {
	t.Run("enable tls", func(t *testing.T) {
		p, err := NewProm(&Option{
			URL:     "http://127.0.0.1:9100",
			TLSOpen: true,
		})
		assert.NoError(t, err)
		transport, ok := p.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, transport.TLSClientConfig.InsecureSkipVerify, true)
	})

	t.Run("tls with ca", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir("./", "__tmp")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir) // nolint:errcheck
		f, err := ioutil.TempFile(tmpDir, "ca.crt")
		assert.NoError(t, err)
		_, err = f.WriteString(caContent)
		assert.NoError(t, err)
		caFile := f.Name()
		defer os.Remove(caFile) // nolint:errcheck
		p, err := NewProm(&Option{
			URL:        "http://127.0.0.1:9100",
			TLSOpen:    true,
			CacertFile: caFile,
		})

		assert.NoError(t, err)
		transport, ok := p.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, transport.TLSClientConfig.InsecureSkipVerify, false)
	})
}

func Test_Auth(t *testing.T) {
	p, err := NewProm(&Option{
		URL: promURL,
		Auth: map[string]string{
			"type":  "bearer_token",
			"token": ".....",
		},
	})
	assert.NoError(t, err)
	r, err := p.GetReq(p.opt.URL)
	assert.NoError(t, err)
	authHeader, ok := r.Header["Authorization"]
	assert.True(t, ok)

	assert.Equal(t, len(authHeader), 1)
	assert.Contains(t, authHeader[0], "Bearer ")
}

func Test_Option(t *testing.T) {
	o := Option{
		Disable: true,
	}
	assert.True(t, o.IsDisable(), o.Disable)

	// GetSource
	assert.Equal(t, o.GetSource("p"), "p")
	assert.Equal(t, o.GetSource(), "prom")
	o.Source = "p1"
	assert.Equal(t, o.GetSource("p"), "p1")

	// GetIntervalDuration
	assert.Equal(t, o.GetIntervalDuration(), defaultInterval)
	o.interval = 1 * time.Second
	assert.Equal(t, o.GetIntervalDuration(), 1*time.Second)
	o.interval = 0
	o.Interval = "10s"
	assert.Equal(t, o.GetIntervalDuration(), 10*time.Second)
	assert.Equal(t, o.interval, 10*time.Second)
}

func Test_WriteFile(t *testing.T) {
	mockBody := `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.03",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.1",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.3",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.5",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="10",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="404",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // nolint:errcheck
	f, err := ioutil.TempFile(tmpDir, "output")
	assert.NoError(t, err)
	outputFile, err := filepath.Abs(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	p, err := NewProm(&Option{
		URL:         promURL,
		Output:      outputFile,
		MaxFileSize: 100000,
	})

	assert.NoError(t, err)
	p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
	err = p.WriteMetricText2File(p.opt.URL)

	assert.NoError(t, err)

	fileContent, err := ioutil.ReadFile(outputFile)

	assert.NoError(t, err)

	assert.Equal(t, string(fileContent), mockBody)
}

func TestIgnoreReqErr(t *testing.T) {
	testCases := []struct {
		name string
		in   *Option
		fail bool
	}{
		{
			name: "ignore url request error",
			in:   &Option{IgnoreReqErr: true, URL: "127.0.0.1:999999"},
			fail: false,
		},
		{
			name: "do not ignore url request error",
			in:   &Option{IgnoreReqErr: false, URL: "127.0.0.1:999999"},
			fail: true,
		},
	}
	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			_, err = p.CollectFromHTTP(p.opt.URL)
			if err != nil {
				if tc.fail {
					t.Logf("[%d] returned an error as expected: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			// Expect to fail but it didn't.
			if tc.fail {
				t.Errorf("[%d] expected to fail but it didn't", idx)
			}
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestProm(t *testing.T) {
	testCases := []struct {
		name     string
		in       *Option
		fail     bool
		expected []string
	}{
		{
			name: "counter metric type only",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"counter"},
			},
			fail: false,
			expected: []string{
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "histogram metric type only",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"histogram"},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
			},
		},

		{
			name: "default metric types",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "all metric types",
			in: &Option{
				URL:         promURL,
				MetricTypes: []string{"histogram", "gauge", "counter", "summary", "untyped"},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "metric name filtering",
			in: &Option{
				URL:              promURL,
				MetricNameFilter: []string{"http"},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "regex metric name filtering",
			in: &Option{
				URL:              promURL,
				MetricNameFilter: []string{"promht+p_metric_han[a-z]ler_req[^abcd]ests_total?"},
			},
			fail: false,
			expected: []string{
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
			},
		},

		{
			name: "measurement name prefix",
			in: &Option{
				URL:               promURL,
				MeasurementPrefix: "prefix_",
			},
			fail: false,
			expected: []string{
				"prefix_go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"prefix_go,quantile=0 gc_duration_seconds=0",
				"prefix_go,quantile=0.25 gc_duration_seconds=0",
				"prefix_go,quantile=0.5 gc_duration_seconds=0",
				"prefix_http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"prefix_http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"prefix_promhttp metric_handler_requests_in_flight=1",
				"prefix_promhttp,cause=encoding metric_handler_errors_total=0",
				"prefix_promhttp,cause=gathering metric_handler_errors_total=0",
				"prefix_promhttp,code=200 metric_handler_requests_total=15143",
				"prefix_promhttp,code=500 metric_handler_requests_total=0",
				"prefix_promhttp,code=503 metric_handler_requests_total=0",
				"prefix_up up=1",
			},
		},

		{
			name: "measurement name",
			in: &Option{
				URL:             promURL,
				MeasurementName: "measurement_name",
			},
			fail: false,
			expected: []string{
				"measurement_name go_gc_duration_seconds_count=0,go_gc_duration_seconds_sum=0",
				"measurement_name promhttp_metric_handler_requests_in_flight=1",
				"measurement_name up=1",
				"measurement_name,cause=encoding promhttp_metric_handler_errors_total=0",
				"measurement_name,cause=gathering promhttp_metric_handler_errors_total=0",
				"measurement_name,code=200 promhttp_metric_handler_requests_total=15143",
				"measurement_name,code=500 promhttp_metric_handler_requests_total=0",
				"measurement_name,code=503 promhttp_metric_handler_requests_total=0",
				"measurement_name,le=+Inf,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.003,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.03,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.1,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=0.3,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=1.5,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,le=10,method=GET,status_code=404 http_request_duration_seconds_bucket=1i",
				"measurement_name,method=GET,status_code=404 http_request_duration_seconds_count=1,http_request_duration_seconds_sum=0.002451013",
				"measurement_name,quantile=0 go_gc_duration_seconds=0",
				"measurement_name,quantile=0.25 go_gc_duration_seconds=0",
				"measurement_name,quantile=0.5 go_gc_duration_seconds=0",
			},
		},

		{
			name: "tags filtering",
			in: &Option{
				URL:        promURL,
				TagsIgnore: []string{"status_code", "method"},
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"http,le=+Inf request_duration_seconds_bucket=1i",
				"http,le=0.003 request_duration_seconds_bucket=1i",
				"http,le=0.03 request_duration_seconds_bucket=1i",
				"http,le=0.1 request_duration_seconds_bucket=1i",
				"http,le=0.3 request_duration_seconds_bucket=1i",
				"http,le=1.5 request_duration_seconds_bucket=1i",
				"http,le=10 request_duration_seconds_bucket=1i",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
			},
		},

		{
			name: "rename-measurement",
			in: &Option{
				URL: promURL,
				Measurements: []Rule{
					{
						Prefix: "go_",
						Name:   "with_prefix_go",
					},
					{
						Prefix: "request_",
						Name:   "with_prefix_request",
					},
				},
			},
			fail: false,
			expected: []string{
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
				"with_prefix_go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"with_prefix_go,quantile=0 gc_duration_seconds=0",
				"with_prefix_go,quantile=0.25 gc_duration_seconds=0",
				"with_prefix_go,quantile=0.5 gc_duration_seconds=0",
			},
		},

		{
			name: "custom tags",
			in: &Option{
				URL:  promURL,
				Tags: map[string]string{"some_tag": "some_value", "more_tag": "some_other_value"},
			},
			fail: false,
			expected: []string{
				"go,more_tag=some_other_value,quantile=0,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,quantile=0.25,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,quantile=0.5,some_tag=some_value gc_duration_seconds=0",
				"go,more_tag=some_other_value,some_tag=some_value gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"http,le=+Inf,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,more_tag=some_other_value,some_tag=some_value,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp,cause=encoding,more_tag=some_other_value,some_tag=some_value metric_handler_errors_total=0",
				"promhttp,cause=gathering,more_tag=some_other_value,some_tag=some_value metric_handler_errors_total=0",
				"promhttp,code=200,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=15143",
				"promhttp,code=500,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=0",
				"promhttp,code=503,more_tag=some_other_value,some_tag=some_value metric_handler_requests_total=0",
				"promhttp,more_tag=some_other_value,some_tag=some_value metric_handler_requests_in_flight=1",
				"up,more_tag=some_other_value,some_tag=some_value up=1",
			},
		},

		{
			name: "multiple urls",
			in: &Option{
				URLs:         []string{"localhost:1234", "localhost:5678"},
				IgnoreReqErr: true,
			},
			fail: false,
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=0",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=+Inf,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.03,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.1,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.3,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=1.5,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=10,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"http,method=GET,status_code=404 request_duration_seconds_count=1,request_duration_seconds_sum=0.002451013",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=200 metric_handler_requests_total=15143",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=500 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"promhttp,code=503 metric_handler_requests_total=0",
				"up up=1",
				"up up=1",
			},
		},
	}

	mockBody := `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.03",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.1",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.3",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.5",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="10",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="404",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

	for idx, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if err != nil {
				t.Errorf("[%d] failed to init prom: %s", idx, err)
			}
			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true
			var points []*point.Point
			for _, u := range p.opt.URLs {
				pts, err := p.CollectFromHTTP(u)
				if err != nil {
					break
				}
				points = append(points, pts...)
			}
			if err != nil {
				if tc.fail {
					t.Logf("[%d] returned an error as expected: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			// Expect to fail but it didn't.
			if tc.fail {
				t.Errorf("[%d] expected to fail but it didn't", idx)
			}

			var got []string
			for _, p := range points {
				s := p.String()
				// remove timestamp
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestCollectFromFile(t *testing.T) {
	mockBody := `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.03",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.1",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="0.3",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="1.5",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="10",status_code="404",method="GET"} 1
http_request_duration_seconds_bucket{le="+Inf",status_code="404",method="GET"} 1
http_request_duration_seconds_sum{status_code="404",method="GET"} 0.002451013
http_request_duration_seconds_count{status_code="404",method="GET"} 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
`

	f, err := os.CreateTemp("./", "test_collect_from_file_")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer os.Remove(f.Name()) //nolint:errcheck,gosec

	if _, err := f.WriteString(mockBody); err != nil {
		t.Errorf("fail to write mock body to temporary file: %v", err)
	}
	if err := f.Sync(); err != nil {
		t.Errorf("fail to flush data to disk: %v", err)
	}
	option := Option{
		URLs: []string{f.Name()},
	}
	p, err := NewProm(&option)
	if err != nil {
		t.Errorf("failed to init prom: %s", err)
	}
	if _, err := p.CollectFromFile(f.Name()); err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetTimestampS(t *testing.T) {
	var (
		ts        int64 = 1647959040488
		startTime       = time.Unix(1600000000, 0)
	)
	m1 := dto.Metric{
		TimestampMs: &ts,
	}
	m2 := dto.Metric{}
	tu.Equals(t, int64(1647959040000000000), getTimestampS(&m1, startTime).UnixNano())
	tu.Equals(t, int64(1600000000000000000), getTimestampS(&m2, startTime).UnixNano())
}

func TestRenameTag(t *testing.T) {
	tm := time.Now()
	cases := []struct {
		name     string
		opt      *Option
		promdata string
		expect   []*point.Point
	}{
		{
			name: "rename-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: false,
					Mapping: map[string]string{
						"status_code":    "StatusCode",
						"tag-not-exists": "do-nothing",
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						map[string]string{"le": "0.003", "StatusCode": "404", "method": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
			`,
		},

		{
			name: "rename-overwrite-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						// method key removed, it's value overwrite tag_exists's value
						map[string]string{"le": "0.003", "StatusCode": "404", "tag_exists": "GET"},
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",tag_exists="yes",status_code="404",method="GET"} 1
			`,
		},

		{
			name: "rename-tags-disable-overwrite",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: false, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						map[string]string{"le": "0.003", "tag_exists": "yes", "StatusCode": "404", "method": "GET"}, // overwrite not work on method
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket{le="0.003",status_code="404",tag_exists="yes", method="GET"} 1
			`,
		},

		{
			name: "empty-tags",
			opt: &Option{
				URL: "http://not-set",
				RenameTags: &RenameTags{
					OverwriteExistTags: true, // enable overwrite
					Mapping: map[string]string{
						"status_code": "StatusCode",
						"method":      "tag_exists", // rename `method` to a exists tag key
					},
				},
			},
			expect: []*point.Point{
				func() *point.Point {
					pt, err := point.NewPoint("http",
						nil,
						map[string]interface{}{"request_duration_seconds_bucket": 1.0},
						&point.PointOption{Category: datakit.Metric, Time: tm})
					if err != nil {
						t.Errorf("NewPoint: %s", err)
						return nil
					}
					return pt
				}(),
			},
			promdata: `
http_request_duration_seconds_bucket 1
			`,
		},
	}

	for _, tc := range cases {
		p, err := NewProm(tc.opt)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run(tc.name, func(t *testing.T) {
			pts, err := p.text2Metrics(bytes.NewBufferString(tc.promdata), "")
			if err != nil {
				t.Error(err)
				return
			}

			for idx, pt := range pts {
				tu.Equals(t, tc.expect[idx].PrecisionString("m"), pt.PrecisionString("m"))
				t.Log(tc.expect[idx].PrecisionString("m"))
			}
		})
	}
}

func TestSetHeaders(t *testing.T) {
	testcases := []struct {
		name string
		opt  *Option
	}{
		{
			name: "add custom http header",
			opt: &Option{
				URL: "dummy_url",
				HTTPHeaders: map[string]string{
					"Root":    "passwd",
					"Michael": "12345",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.opt)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := p.GetReq(p.opt.URL)
			assert.NoError(t, err)
			for k, v := range p.opt.HTTPHeaders {
				assert.Equal(t, v, req.Header.Get(k))
			}
		})
	}
}

func TestInfoTag(t *testing.T) {
	testcases := []struct {
		in       *Option
		name     string
		fail     bool
		expected []string
	}{
		{
			name: "type-info",
			in:   &Option{URL: promURL},
			expected: []string{
				"process,host_arch=amd64,host_name=DESKTOP-3JJLRI8,os_description=Windows\\ 11\\ 10.0,os_type=windows,otel_scope_name=otlp-server,pool=mapped\\ -\\ 'non-volatile\\ memory',process_command_line=D:\\software_installer\\java\\jdk-17\\bin\\java.exe\\ -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar\\ -Dotel.traces.exporter\\=otlp\\ -Dotel.exporter.otlp.endpoint\\=http://localhost:4317\\ -Dotel.resource.attributes\\=service.name\\=server\\,username\\=liu\\ -Dotel.metrics.exporter\\=otlp\\ -Dotel.propagators\\=b3\\ -Dotel.metrics.exporter\\=prometheus\\ -Dotel.exporter.prometheus.port\\=10086\\ -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled\\=true\\ -XX:TieredStopAtLevel\\=1\\ -Xverify:none\\ -Dspring.output.ansi.enabled\\=always\\ -Dcom.sun.management.jmxremote\\ -Dspring.jmx.enabled\\=true\\ -Dspring.liveBeansView.mbeanDomain\\ -Dspring.application.admin.enabled\\=true\\ -javaagent:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\lib\\idea_rt.jar\\=55275:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\bin\\ -Dfile.encoding\\=UTF-8,process_executable_path=D:\\software_installer\\java\\jdk-17\\bin\\java.exe,process_pid=23592,process_runtime_description=Oracle\\ Corporation\\ Java\\ HotSpot(TM)\\ 64-Bit\\ Server\\ VM\\ 17.0.6+9-LTS-190,process_runtime_name=Java(TM)\\ SE\\ Runtime\\ Environment,process_runtime_version=17.0.6+9-LTS-190,service_name=server,telemetry_auto_version=1.24.0-SNAPSHOT,telemetry_sdk_language=java,telemetry_sdk_name=opentelemetry,telemetry_sdk_version=1.23.1,username=liu runtime_jvm_buffer_count=0",
			},
		},
	}
	mockBody := `# TYPE target info
# HELP target Target metadata
target_info{host_arch="amd64",host_name="DESKTOP-3JJLRI8",os_description="Windows 11 10.0",os_type="windows",process_command_line="D:\\software_installer\\java\\jdk-17\\bin\\java.exe -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar -Dotel.traces.exporter=otlp -Dotel.exporter.otlp.endpoint=http://localhost:4317 -Dotel.resource.attributes=service.name=server,username=liu -Dotel.metrics.exporter=otlp -Dotel.propagators=b3 -Dotel.metrics.exporter=prometheus -Dotel.exporter.prometheus.port=10086 -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled=true -XX:TieredStopAtLevel=1 -Xverify:none -Dspring.output.ansi.enabled=always -Dcom.sun.management.jmxremote -Dspring.jmx.enabled=true -Dspring.liveBeansView.mbeanDomain -Dspring.application.admin.enabled=true -javaagent:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\lib\\idea_rt.jar=55275:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\bin -Dfile.encoding=UTF-8",process_executable_path="D:\\software_installer\\java\\jdk-17\\bin\\java.exe",process_pid="23592",process_runtime_description="Oracle Corporation Java HotSpot(TM) 64-Bit Server VM 17.0.6+9-LTS-190",process_runtime_name="Java(TM) SE Runtime Environment",process_runtime_version="17.0.6+9-LTS-190",service_name="server",telemetry_auto_version="1.24.0-SNAPSHOT",telemetry_sdk_language="java",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.23.1",username="liu"} 1
# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
# TYPE otel_scope_info info
# HELP otel_scope_info Scope metadata
otel_scope_info{otel_scope_name="otlp-server"} 1
`

	for idx, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true

			points, err := p.CollectFromHTTP(p.opt.URL)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var got []string
			for _, p := range points {
				s := p.String()
				// remove timestamp
				got = append(got, s[:strings.LastIndex(s, " ")])
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}

func TestDuplicateComment(t *testing.T) {
	testcases := []struct {
		in       *Option
		name     string
		fail     bool
		expected []string
	}{
		{
			name: "duplicate-comment",
			in:   &Option{URL: promURL},
			expected: []string{
				"go gc_duration_seconds_count=0,gc_duration_seconds_sum=0",
				"go,quantile=0 gc_duration_seconds=0",
				"go,quantile=0 gc_duration_seconds=1",
				"go,quantile=0.25 gc_duration_seconds=0",
				"go,quantile=0.25 gc_duration_seconds=1",
				"go,quantile=0.5 gc_duration_seconds=0",
				"go,quantile=0.5 gc_duration_seconds=1",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=1i",
				"http,le=0.003,method=GET,status_code=404 request_duration_seconds_bucket=2i",
				"http,method=GET,status_code=404 request_duration_seconds_count=0,request_duration_seconds_sum=0",
				"promhttp metric_handler_requests_in_flight=1",
				"promhttp metric_handler_requests_in_flight=2",
				"promhttp,cause=encoding metric_handler_errors_total=0",
				"promhttp,cause=encoding metric_handler_errors_total=1",
				"promhttp,cause=gathering metric_handler_errors_total=0",
				"promhttp,cause=gathering metric_handler_errors_total=1",
				"up up=0",
				"up up=1",
			},
		},
	}
	mockBody := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 1
promhttp_metric_handler_errors_total{cause="gathering"} 1
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 2
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 1
go_gc_duration_seconds{quantile="0.25"} 1
go_gc_duration_seconds{quantile="0.5"} 1
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 1
# HELP http_request_duration_seconds duration histogram of http responses labeled with: status_code, method
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.003",status_code="404",method="GET"} 2
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 1
# HELP up 1 = up, 0 = not up
# TYPE up untyped
up 0
`

	for idx, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProm(tc.in)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			p.SetClient(&http.Client{Transport: newTransportMock(mockBody)})
			p.opt.DisableInstanceTag = true

			points, err := p.CollectFromHTTP(p.opt.URL)
			if tc.fail && assert.Error(t, err) {
				return
			} else {
				assert.NoError(t, err)
			}

			var got []string
			for _, p := range points {
				s := p.String()
				// remove timestamp
				got = append(got, s[:strings.LastIndex(s, " ")])
			}
			sort.Strings(got)
			tu.Equals(t, strings.Join(tc.expected, "\n"), strings.Join(got, "\n"))
			t.Logf("[%d] PASS", idx)
		})
	}
}
