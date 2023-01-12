// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const TOKEN string = "123456"

func getResponse(req *http.Request, config *DCAConfig) *httptest.ResponseRecorder {
	dcaConfig = &DCAConfig{}
	if config != nil {
		dcaConfig = config
	}
	dwCfg := &dataway.DataWayCfg{URLs: []string{"http://localhost:9529?token=123456"}}
	dw = &dataway.DataWayDefault{}
	dw.Init(dwCfg) //nolint: errcheck
	router := setupDcaRouter()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func getResponseBody(w *httptest.ResponseRecorder) (*dcaResponse, error) {
	res := &dcaResponse{}
	err := json.Unmarshal(w.Body.Bytes(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func TestCors(t *testing.T) {
	req, _ := http.NewRequest("GET", "/v1/dca/stats", nil)

	w := getResponse(req, nil)
	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Headers"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Credentials"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Methods"))
}

type TestCase struct {
	Title          string
	Method         string
	URL            string
	Header         map[string]string
	IsCorrectToken bool
	IsLoopback     bool
	RemoteAddr     string
	Expected       *dcaResponse
	ExpectedNot    *dcaResponse
	DcaConfg       *DCAConfig
	SubCases       []TestCase
}

func runTestCases(t *testing.T, cases []TestCase) {
	t.Helper()
	for _, tc := range cases {
		title := "test"
		if len(tc.Title) > 0 {
			title = tc.Title
		}
		t.Run(title, func(t *testing.T) {
			if len(tc.SubCases) > 0 {
				runTestCases(t, tc.SubCases)
				return
			}
			method := "GET"
			url := "/"
			if len(tc.Method) > 0 {
				method = tc.Method
			}
			if len(tc.URL) > 0 {
				url = tc.URL
			}

			req, _ := http.NewRequest(method, url, nil)

			for k, v := range tc.Header {
				req.Header.Add(k, v)
			}

			if tc.IsCorrectToken {
				req.Header.Add("X-Token", TOKEN)
			}

			if len(tc.RemoteAddr) > 0 {
				req.RemoteAddr = tc.RemoteAddr
			}

			if tc.IsLoopback {
				req.RemoteAddr = "127.0.0.1:10000"
			}

			dcaConfig := &DCAConfig{}
			if tc.DcaConfg != nil {
				dcaConfig = tc.DcaConfg
			}

			w := getResponse(req, dcaConfig)

			res, _ := getResponseBody(w)

			if tc.Expected != nil {
				if tc.Expected.Code > 0 {
					assert.Equal(t, tc.Expected.Code, res.Code)
				}

				if len(tc.Expected.ErrorCode) > 0 {
					assert.Equal(t, tc.Expected.ErrorCode, res.ErrorCode)
				}
			}

			if tc.ExpectedNot != nil {
				if tc.ExpectedNot.Code > 0 {
					assert.NotEqual(t, tc.ExpectedNot.Code, res.Code)
				}

				if len(tc.ExpectedNot.ErrorCode) > 0 {
					assert.NotEqual(t, tc.ExpectedNot.ErrorCode, res.ErrorCode)
				}
			}
		})
	}
}

func TestDca(t *testing.T) {
	testCases := []TestCase{
		{
			Title:          "test default route",
			URL:            "/invalid/url",
			IsCorrectToken: true,
			IsLoopback:     true,
			Expected:       &dcaResponse{Code: 404, ErrorCode: "route.not.found"},
		},
		{
			Title: "test dcaAuthMiddleware",
			SubCases: []TestCase{
				{
					Title:          "correct token",
					IsCorrectToken: true,
					IsLoopback:     true,
					ExpectedNot:    &dcaResponse{ErrorCode: "auth.failed", Code: 401},
				},
				{
					Title:          "wrong token",
					IsCorrectToken: false,
					IsLoopback:     true,
					Expected:       &dcaResponse{ErrorCode: "auth.failed", Code: 401},
				},
			},
		},
		{
			Title: "white list",
			SubCases: []TestCase{
				{
					Title:    "white list check error",
					DcaConfg: &DCAConfig{WhiteList: []string{"111.111.111.111"}},
					Expected: &dcaResponse{Code: 401, ErrorCode: "whiteList.check.error"},
				},
				{
					Title:       "ignore loopback ip",
					IsLoopback:  true,
					ExpectedNot: &dcaResponse{ErrorCode: "whiteList.check.error"},
				},
				{
					Title:       "client ip in whitelist",
					RemoteAddr:  "111.111.111.111:10000",
					DcaConfg:    &DCAConfig{WhiteList: []string{"111.111.111.111"}},
					ExpectedNot: &dcaResponse{ErrorCode: "whiteList.check.error"},
				},
			},
		},
		{
			Title: "api test",
			SubCases: []TestCase{
				{
					Title: "dcaStats",
					URL:   "/v1/dca/stats",
				},
			},
		},
	}

	runTestCases(t, testCases)
}

func TestDcaStats(t *testing.T) {
	req, _ := http.NewRequest("GET", "/v1/dca/stats", nil)
	req.Header.Add("X-Token", TOKEN)
	hostName := "XXXXX"

	// mock
	dcaAPI.GetStats = func() (*DatakitStats, error) {
		return &DatakitStats{HostName: hostName}, nil
	}

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.Equal(t, 200, res.Code)
	content, ok := res.Content.(map[string]interface{})
	assert.True(t, ok)
	hostNameValue, ok := content["hostname"]
	assert.True(t, ok)
	assert.Equal(t, hostName, hostNameValue)
}

func TestDcaReload(t *testing.T) {
	// reload ok
	dcaAPI.RestartDataKit = func() error {
		return nil
	}

	req, _ := http.NewRequest("GET", "/v1/dca/reload", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.Equal(t, 200, res.Code)

	// reload fail
	dcaAPI.RestartDataKit = func() error {
		return errors.New("restart error")
	}

	w = getResponse(req, nil)
	res, _ = getResponseBody(w)
	assert.Equal(t, 500, res.Code)
	assert.Equal(t, "system.restart.error", res.ErrorCode)
}

func TestDcaSaveConfig(t *testing.T) {
	inputName := "demo-input"
	inputs.ConfigInfo[inputName] = &inputs.Config{}

	confDir, err := ioutil.TempDir("./", "conf")
	if err != nil {
		l.Fatal(err)
	}
	defer os.RemoveAll(confDir) //nolint: errcheck
	datakit.ConfdDir = confDir
	f, err := ioutil.TempFile(confDir, "new-conf*.conf")
	assert.NoError(t, err)

	bodyTemplate := `{"path": "%s","config":"%s", "isNew":%s, "inputName": "%s"}`
	config := "[input]"
	body := strings.NewReader(fmt.Sprintf(bodyTemplate, f.Name(), config, "true", inputName))
	req, _ := http.NewRequest("POST", "/v1/dca/saveConfig", body)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)

	res, _ := getResponseBody(w)

	content, ok := res.Content.(map[string]interface{})
	assert.True(t, ok)
	path, ok := content["path"]
	assert.True(t, ok)
	assert.Equal(t, f.Name(), path)

	confContent, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, config, string(confContent))

	configPaths := inputs.ConfigInfo[inputName].ConfigPaths
	assert.Equal(t, 1, len(configPaths))
	assert.Equal(t, &inputs.ConfigPathStat{Loaded: 2, Path: f.Name()}, configPaths[0])
}

func TestGetConfig(t *testing.T) {
	// no path
	req, _ := http.NewRequest("GET", "/v1/dca/getConfig", nil)
	req.Header.Add("X-Token", TOKEN)
	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.False(t, res.Success)

	// invalid path
	req, _ = http.NewRequest("GET", "/v1/dca/getConfig?path=xxxxxxx.conf", nil)
	req.Header.Add("X-Token", TOKEN)
	w = getResponse(req, nil)
	res, _ = getResponseBody(w)

	assert.False(t, res.Success)
	assert.Equal(t, "params.invalid.path_invalid", res.ErrorCode)

	// get config ok
	confDir, err := ioutil.TempDir("./", "conf")
	if err != nil {
		l.Fatal(err)
	}
	defer os.RemoveAll(confDir) //nolint: errcheck
	datakit.ConfdDir = confDir
	f, err := ioutil.TempFile(confDir, "new-conf*.conf")
	assert.NoError(t, err)
	defer os.Remove(f.Name()) //nolint: errcheck

	config := "[input]"

	err = ioutil.WriteFile(f.Name(), []byte(config), os.ModePerm)
	assert.NoError(t, err)

	req, _ = http.NewRequest("GET", "/v1/dca/getConfig?path="+f.Name(), nil)
	req.Header.Add("X-Token", TOKEN)
	w = getResponse(req, nil)
	res, _ = getResponseBody(w)

	assert.True(t, res.Success)
	assert.Equal(t, config, res.Content)
}

func TestDcaGetPipelines(t *testing.T) {
	pipelineDir, err := ioutil.TempDir("./", "pipeline")
	datakit.PipelineDir = pipelineDir

	defer os.RemoveAll(pipelineDir) //nolint: errcheck
	assert.NoError(t, err)

	f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/v1/dca/pipelines", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	content, ok := res.Content.([]interface{})
	assert.True(t, ok)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, 1, len(content))
	pipelineInfo, ok := content[0].(map[string]interface{})
	assert.True(t, ok)
	fileName, ok := pipelineInfo["fileName"]
	assert.True(t, ok)
	fileDir, ok := pipelineInfo["fileDir"]
	assert.True(t, ok)

	assert.Equal(t, filepath.Clean(f.Name()), filepath.Join(fmt.Sprintf("%v", fileDir), fmt.Sprintf("%v", fileName)))
}

func TestDcaGetPipelinesDetail(t *testing.T) {
	pipelineDir, err := ioutil.TempDir("./", "pipeline")
	datakit.PipelineDir = pipelineDir

	defer os.RemoveAll(pipelineDir) //nolint: errcheck
	assert.NoError(t, err)

	f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
	assert.NoError(t, err)

	pipelineContent := "this is demo pipeline"
	fileName := filepath.Base(f.Name())

	err = ioutil.WriteFile(f.Name(), []byte(pipelineContent), os.ModePerm)
	assert.NoError(t, err)

	testCases := []struct {
		Title    string
		IsOk     bool
		FileName string
	}{
		{
			Title: "no query parameter `fileName`",
		},
		{
			Title:    "invalid `fileName` format",
			FileName: "xxxxxx",
		},
		{
			Title:    "file not exist",
			FileName: "invalid.p",
		},
		{
			Title:    "get pipeline ok",
			FileName: fileName,
			IsOk:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			url := "/v1/dca/pipelines/detail"
			if len(tc.FileName) > 0 {
				url += "?fileName=" + tc.FileName
			}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Add("X-Token", TOKEN)

			w := getResponse(req, nil)
			res, _ := getResponseBody(w)

			assert.Equal(t, tc.IsOk, res.Success)
		})
	}
}

func TestDcaTestPipelines(t *testing.T) {
	testCases := []struct {
		Title        string
		TestPipeline func(string, string) (string, error)
		Body         string
		IsOk         bool
	}{
		{
			Title: "test ok",
			Body: `{
				"script_name": "nginx", 
				"category": "logging",
				"pipeline": {
					"logging": {
						"nginx": "add_key(name, \"test-pipeline\")"
					}
				},
				"data": ["test"]
			}`,
			IsOk: true,
		},
		{
			Title:        "test pipeline failed",
			TestPipeline: func(s1, s2 string) (string, error) { return "", errors.New("pipeline error") },
		},
		{
			Title: "invalid body",
			Body:  "xxxxxxx",
		},
	}
	for _, tc := range testCases {
		t.Logf("test: %s", tc.Title)
		parsedPipeline := "parse text"
		if tc.TestPipeline != nil {
			dcaAPI.TestPipeline = tc.TestPipeline
		} else {
			dcaAPI.TestPipeline = func(s1, s2 string) (string, error) {
				return parsedPipeline, nil
			}
		}
		pipelineDir, err := ioutil.TempDir("./", "pipeline")
		datakit.PipelineDir = pipelineDir

		defer os.RemoveAll(pipelineDir) //nolint: errcheck
		assert.NoError(t, err)

		f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
		assert.NoError(t, err)

		pipelineContent := "this is demo pipeline"

		err = ioutil.WriteFile(f.Name(), []byte(pipelineContent), os.ModePerm)
		assert.NoError(t, err)

		body := strings.NewReader(tc.Body)
		req, _ := http.NewRequest("POST", "/v1/dca/pipelines/test", body)
		req.Header.Add("X-Token", TOKEN)

		w := getResponse(req, nil)
		res, _ := getResponseBody(w)

		if tc.IsOk {
			assert.True(t, res.Success)
		} else {
			assert.False(t, res.Success)
		}
	}
}

func TestDcaCreatePipeline(t *testing.T) {
	testCases := []struct {
		Title          string
		IsOk           bool
		FileName       string
		Category       string
		IsContentEmpty bool
		Body           string
	}{
		{
			Title: "create ok",
			IsOk:  true,
		},
		{
			Title: "invalid body format",
			Body:  "invalid",
		},
		{
			Title:          "content is empty",
			IsContentEmpty: true,
			IsOk:           true,
		},
		{
			Title:    "invalid fileName",
			FileName: "pipeline",
		},
		{
			Title:    "test category",
			FileName: "test.p",
			Category: "logging",
			IsOk:     true,
		},
		{
			Title:    "test invalid category",
			FileName: "test.p",
			Category: "logging-invalid",
			IsOk:     false,
		},
	}

	for _, tc := range testCases {
		t.Logf("testing: %s", tc.Title)
		pipelineDir, err := ioutil.TempDir("./", "pipeline")
		assert.NoError(t, err)
		defer os.RemoveAll(pipelineDir) //nolint: errcheck

		datakit.PipelineDir = pipelineDir

		err = os.Mkdir(filepath.Join(pipelineDir, "logging"), 0o777)
		assert.NoError(t, err)

		pipelineContent := "this is demo pipeline"

		if tc.IsContentEmpty {
			pipelineContent = ""
		}

		fileName := "custom_pipeline.p"

		if len(tc.FileName) > 0 {
			fileName = tc.FileName
		}

		var body *strings.Reader
		if len(tc.Body) > 0 {
			body = strings.NewReader(tc.Body)
		} else {
			bodyTemplate := `{"fileName":"%s", "category": "%s","content": "%s"}`
			body = strings.NewReader(fmt.Sprintf(bodyTemplate, fileName, tc.Category, pipelineContent))
		}
		req, _ := http.NewRequest("POST", "/v1/dca/pipelines", body)
		req.Header.Add("X-Token", TOKEN)

		w := getResponse(req, nil)
		res, _ := getResponseBody(w)

		if tc.IsOk {
			assert.True(t, res.Success)
		} else {
			assert.False(t, res.Success)
		}
	}
}

func TestDcaUploadSourcemap(t *testing.T) {
	dir, err := ioutil.TempDir("./", "tmp")
	if err != nil {
		t.Fatal("create tmp dir eror")
	}
	datakit.DataDir = dir
	datakit.DataRUMDir = filepath.Join(dir, "rum")
	defer os.RemoveAll(dir) //nolint: errcheck
	testCases := []struct {
		title       string
		appId       string
		env         string
		version     string
		fileContent string
		isOk        bool
	}{
		{
			title:       "upload ok",
			appId:       "app_1234",
			env:         "test",
			version:     "0.0.0",
			fileContent: "xxxxxx",
			isOk:        true,
		},
		{
			title:       "param missing",
			env:         "test",
			version:     "0.0.0",
			fileContent: "xxxxxx",
			isOk:        false,
		},
		{
			title:   "file missing",
			appId:   "app_123",
			env:     "test",
			version: "0.0.0",
			isOk:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appId := tc.appId
			env := tc.env
			version := tc.version

			url := fmt.Sprintf("/v1/rum/sourcemap?app_id=%s&env=%s&version=%s&platform=web", appId, env, version)
			var body io.Reader
			data := ""
			boundary := "file"
			if len(tc.fileContent) > 0 {
				data += "--" + boundary + "\n"
				data += "Content-Disposition: form-data; name=\"file\";filename=\"1.zip\"\n\n"
				data += tc.fileContent + "\n"
				data += "--" + boundary + "--"

				body = strings.NewReader(data)
			}

			req, _ := http.NewRequest("POST", url, body)
			req.Header.Add("X-Token", TOKEN)
			if body != nil {
				req.Header.Add("Content-Type", "multipart/form-data;boundary="+boundary)
			}

			w := getResponse(req, nil)
			res, _ := getResponseBody(w)
			fmt.Printf("%+#v", res)

			assert.Equal(t, tc.isOk, res.Success)
		})
	}
}

func TestDcaGetFilter(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pullFilePath := filepath.Join(tmpDir, ".pull")
	pullContent := `{
		"dataways": null,
		"filters": {
		  "logging": [
			"{ source = 'datakit'  and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}"
		  ]
		},
		"pull_interval": 10000000000,
		"remote_pipelines": null
	  }`
	err = ioutil.WriteFile(pullFilePath, []byte(pullContent), os.ModePerm)
	if err != nil {
		t.Fatal("create pull file error", err)
	}

	cases := []struct {
		expected string
		dataDir  string
	}{
		{
			expected: pullContent,
			dataDir:  tmpDir,
		},
		{
			expected: "",
			dataDir:  filepath.Join(tmpDir, "invalid_data_dir"),
		},
	}

	for index, tc := range cases {
		t.Logf("Test #%d: %+v", index, tc)
		datakit.DataDir = tc.dataDir

		req, _ := http.NewRequest("GET", "/v1/filter", nil)
		req.Header.Add("X-Token", TOKEN)

		w := getResponse(req, nil)
		res, _ := getResponseBody(w)

		assert.True(t, res.Success, res)
		if resResult, ok := res.Content.(map[string]interface{}); ok {
			content := resResult["content"]

			assert.Equal(t, tc.expected, content)
		} else {
			t.Fatal("response is not correct", res.Content)
		}
	}
}

func TestDcaGetLogTail(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "log")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("test-log")
	// mock log file
	log = f.Name()
	req, _ := http.NewRequest("GET", "/v1/log/tail", nil)
	req.Header.Add("X-Token", TOKEN)

	// set up dca
	dcaConfig = &DCAConfig{}
	dwCfg := &dataway.DataWayCfg{URLs: []string{"http://localhost:9529?token=123456"}}
	dw = &dataway.DataWayDefault{}
	dw.Init(dwCfg) //nolint: errcheck
	router := setupDcaRouter()
	w := CreateTestResponseRecorder()
	go func() {
		// wait the server
		time.Sleep(1 * time.Second)

		p := make([]byte, 10)
		for {
			n, err := w.Body.Read(p)
			if err != nil && err != io.EOF {
				t.Log("get log response error")
				break
			}
			if n > 0 {
				assert.Equal(t, "test-log", string(p[0:n]))
				break
			}
		}
		w.closeClient()
	}()
	router.ServeHTTP(w, req)
}

func TestDcaDownloadLog(t *testing.T) {
	tmpDir, err := ioutil.TempDir("./", "__tmp")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	f, err := ioutil.TempFile(tmpDir, "log")
	if err != nil {
		t.Fatal(err)
	}
	logStr := "test-download-log"
	f.WriteString(logStr)
	log = f.Name()

	req, _ := http.NewRequest("GET", "/v1/log/download?type=log", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)

	assert.Equal(t, logStr, w.Body.String())
}
