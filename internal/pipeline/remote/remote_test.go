// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package remote

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//------------------------------------------------------------------------------

var (
	writeFileData          *FileDataStruct
	readFileData           []byte
	isFileExist            bool
	readDirResult          []fs.FileInfo
	pullPipelineUpdateTime int64
	pullRelationUpdate     bool
	pullRelationUpdateAt   int64

	errGeneral                   = fmt.Errorf("test_specific_error")
	errMarshal                   error
	errUnMarshal                 error
	errReadFile                  error
	errWriteFile                 error
	errReadDir                   error
	errPullPipeline              error
	errRemove                    error
	errGetNamespacePipelineFiles error
	errReadTarToMap              error
	errWriteTarFromMap           error
)

func resetVars() {
	writeFileData = nil
	readFileData = []byte{}
	isFileExist = false
	readDirResult = []fs.FileInfo{}
	pullPipelineUpdateTime = 0
	pullRelationUpdate = false
	pullPipelineUpdateTime = 0

	errMarshal = nil
	errUnMarshal = nil
	errReadFile = nil
	errWriteFile = nil
	errReadDir = nil
	errPullPipeline = nil
	errRemove = nil
	errGetNamespacePipelineFiles = nil
	errReadTarToMap = nil
	errWriteTarFromMap = nil
}

type fileInfoStruct struct{}

func (fileInfoStruct) Name() string {
	return "useless"
}

func (fileInfoStruct) Size() int64 {
	return 0
}

func (fileInfoStruct) Mode() fs.FileMode {
	return 0
}

func (fileInfoStruct) ModTime() time.Time {
	return time.Time{}
}

func (fileInfoStruct) IsDir() bool {
	return false
}

func (fileInfoStruct) Sys() interface{} {
	return nil
}

// Make sure pipelineRemoteMockerTest implements the IPipelineRemote interface
var _ IPipelineRemote = new(pipelineRemoteMockerTest)

type pipelineRemoteMockerTest struct{}

type FileDataStruct struct {
	FileName string
	Bytes    []byte
}

func (*pipelineRemoteMockerTest) FileExist(filename string) bool {
	return isFileExist
}

func (*pipelineRemoteMockerTest) Marshal(v interface{}) ([]byte, error) {
	if errMarshal != nil {
		return nil, errMarshal
	}

	return json.Marshal(v)
}

func (*pipelineRemoteMockerTest) Unmarshal(data []byte, v interface{}) error {
	if errUnMarshal != nil {
		return errUnMarshal
	}

	return json.Unmarshal(data, v)
}

func (*pipelineRemoteMockerTest) ReadFile(filename string) ([]byte, error) {
	if errReadFile != nil {
		return nil, errReadFile
	}

	return readFileData, nil
}

func (*pipelineRemoteMockerTest) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	if errWriteFile != nil {
		return errWriteFile
	}

	writeFileData = &FileDataStruct{
		FileName: filename,
		Bytes:    data,
	}
	return nil
}

func (*pipelineRemoteMockerTest) ReadDir(dirname string) ([]fs.FileInfo, error) {
	if errReadDir != nil {
		return nil, errReadDir
	}

	return readDirResult, nil
}

func (*pipelineRemoteMockerTest) PullPipeline(ts, relationTS int64) (mFiles, plRelation map[string]map[string]string,
	defaultPl map[string]string, updateTime int64, relationUpdateAt int64, err error,
) {
	if errPullPipeline != nil {
		return nil, nil, nil, 0, 0, errPullPipeline
	}

	return map[string]map[string]string{
			"logging": {
				"123.p": "text123",
				"456.p": "text456",
			},
		}, map[string]map[string]string{
			"logging": {
				"123": "123.p",
				"234": "123.p",
			},
		}, map[string]string{
			"logging": "123.p",
		}, pullPipelineUpdateTime, relationUpdateAt, nil
}

func (*pipelineRemoteMockerTest) GetTickerDurationAndBreak() (time.Duration, bool) {
	return time.Second, true
}

func (*pipelineRemoteMockerTest) Remove(name string) error {
	return errRemove
}

func (*pipelineRemoteMockerTest) FeedLastError(inputName string, err string) {}

func (*pipelineRemoteMockerTest) GetNamespacePipelineFiles(namespace string) ([]string, error) {
	return nil, errGetNamespacePipelineFiles
}

func (*pipelineRemoteMockerTest) ReadTarToMap(srcFile string) (map[string]string, error) {
	return nil, errReadTarToMap
}

func (*pipelineRemoteMockerTest) WriteTarFromMap(data map[string]string, dest string) error {
	return errWriteTarFromMap
}

//------------------------------------------------------------------------------

// go test -v -timeout 30s -run ^TestPullMain$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestPullMain(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name           string
		fileExist      bool
		urls           []string
		pathConfig     string
		siteURL        string
		configContent  []byte
		failedReadFile error
		expectError    error
	}{
		{
			name:          "normal",
			urls:          []string{"https://openway.guance.com?token=tkn_123"},
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name: "urls_zero",
		},
		{
			name:           "do_pull_failed",
			urls:           []string{"https://openway.guance.com?token=tkn_123"},
			fileExist:      true,
			failedReadFile: errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			isFileExist = tc.fileExist
			errReadFile = tc.failedReadFile

			err := pullMain(tc.urls, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "pullMain found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestDoPull$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestDoPull(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"
	const relationPath = "/usr/local/datakit/pipeline_remote/.relation_fake_dump.json"

	cases := []struct {
		name                       string
		fileExist                  bool
		pathConfig                 string
		siteURL                    string
		configContent              []byte
		testPullPipelineUpdateTime int64
		testPullRelationUpdate     bool
		testPullRelationUpdateAt   int64
		testReadDirResult          []fs.FileInfo
		failedMarshal              error
		failedReadFile             error
		failedReadDir              error
		failedPullPipeline         error
		failedRemove               error
		expectError                error
	}{
		{
			name:          "update",
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:           "getPipelineRemoteConfig_fail",
			fileExist:      true,
			failedReadFile: errGeneral,
			expectError:    errGeneral,
		},
		{
			name:               "PullPipeline_fail",
			failedPullPipeline: errGeneral,
			expectError:        errGeneral,
		},
		{
			name:          "alread_up_to_date",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:                       "dumpfile_fail",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
			failedReadDir:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "updatePipelineRemoteConfig_fail",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
			failedMarshal:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "updatePipelineRemoteConfig_pass",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 123,
		},
		{
			name:                       "deleteAll_nil",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
		},
		{
			name:                       "deleteAll_error",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
			failedReadDir:              errGeneral,
			expectError:                errGeneral,
		},
		{
			name:                       "removeLocalRemote_continue",
			pathConfig:                 configPath,
			siteURL:                    dwURL,
			configContent:              []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			testPullPipelineUpdateTime: 1,
			testReadDirResult:          []fs.FileInfo{&fileInfoStruct{}},
			failedRemove:               errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("TestDoPull: tc.name = %s\n", tc.name)

			resetVars()
			readFileData = tc.configContent
			isFileExist = tc.fileExist
			errMarshal = tc.failedMarshal
			errReadFile = tc.failedReadFile
			errReadDir = tc.failedReadDir
			errPullPipeline = tc.failedPullPipeline
			pullPipelineUpdateTime = tc.testPullPipelineUpdateTime
			pullRelationUpdate = tc.testPullRelationUpdate
			pullRelationUpdateAt = tc.testPullRelationUpdateAt
			if len(tc.testReadDirResult) > 0 {
				readDirResult = tc.testReadDirResult
			}
			errRemove = tc.failedRemove

			err := doPull(tc.pathConfig, relationPath, tc.siteURL, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "doPull found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestDumpFiles$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestDumpFiles(t *testing.T) {
	cases := []struct {
		name                  string
		files                 map[string]map[string]string
		readDir               []fs.FileInfo
		failedReadDir         error
		failedWriteTarFromMap error
		expectError           error
	}{
		{
			name: "normal",
			files: map[string]map[string]string{
				"logging": {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
		},
		{
			name:          "read_dir_fail",
			failedReadDir: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:                  "WriteTarFromMap_fail",
			failedWriteTarFromMap: errGeneral,
			files: map[string]map[string]string{
				"logging": {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
			expectError: errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			errReadDir = tc.failedReadDir
			errWriteTarFromMap = tc.failedWriteTarFromMap

			err := dumpFiles(tc.files, nil, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "dumpFiles found error: %v", err)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetPipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestGetPipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"

	cases := []struct {
		name               string
		fileExist          bool
		pathConfig         string
		siteURL            string
		configContent      []byte
		failedUnMarshal    error
		failedReadFile     error
		failedRemove       error
		failedReadDir      error
		failedReadTarToMap error
		expectError        error
		expect             int64
	}{
		{
			name:          "normal",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			expect:        0,
		},
		{
			name:       "config_not_exist",
			pathConfig: "",
		},
		{
			name:           "read_file_fail",
			fileExist:      true,
			pathConfig:     configPath,
			failedReadFile: errGeneral,
			expectError:    errGeneral,
		},
		{
			name:            "json_unmarshal_fail",
			fileExist:       true,
			pathConfig:      configPath,
			failedUnMarshal: errGeneral,
			expectError:     errGeneral,
		},
		{
			name:          "token_changed",
			fileExist:     true,
			pathConfig:    configPath,
			siteURL:       dwURL,
			configContent: []byte(`{"SiteURL":"http://127.0.0.1:9528?token=tkn_123","UpdateTime":1644318398}`),
		},
		{
			name:               "ReadTarToMap_failed",
			fileExist:          true,
			pathConfig:         configPath,
			siteURL:            dwURL,
			configContent:      []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			failedReadTarToMap: errGeneral,
			expect:             0,
		},
		{
			name:          "remove_error",
			fileExist:     true,
			pathConfig:    configPath,
			configContent: []byte(`{"SiteURL":"https://openway.guance.com?token=tkn_123","UpdateTime":1644318398}`),
			failedRemove:  errGeneral,
			failedReadDir: errGeneral,
			expect:        0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isFirst = true // variable from package remote

			resetVars()
			readFileData = tc.configContent
			isFileExist = tc.fileExist
			errUnMarshal = tc.failedUnMarshal
			errReadFile = tc.failedReadFile
			errRemove = tc.failedRemove
			errReadDir = tc.failedReadDir
			errReadTarToMap = tc.failedReadTarToMap

			n, err := getPipelineRemoteConfig(tc.pathConfig, tc.siteURL, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "getPipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, n, "getPipelineRemoteConfig not equal!")
		})
	}
}

// go test -v -timeout 30s -run ^TestUpdatePipelineRemoteConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestUpdatePipelineRemoteConfig(t *testing.T) {
	const dwURL = "https://openway.guance.com?token=tkn_123"
	const configPath = "/usr/local/datakit/pipeline_remote/.config_fake"
	const ts = 1644820678

	cases := []struct {
		name            string
		pathConfig      string
		siteURL         string
		latestTime      int64
		failedMarshal   error
		failedWriteFile error
		expectError     error
		expect          *FileDataStruct
	}{
		{
			name:       "normal",
			pathConfig: configPath,
			siteURL:    dwURL,
			latestTime: ts,
			expect: &FileDataStruct{
				FileName: configPath,
				Bytes: func() []byte {
					cf := pipelineRemoteConfig{
						SiteURL:    dwURL,
						UpdateTime: ts,
					}
					bys, err := json.Marshal(cf)
					if err != nil {
						panic(err)
					}
					return bys
				}(),
			},
		},
		{
			name:          "json_fail",
			failedMarshal: errGeneral,
			expectError:   errGeneral,
		},
		{
			name:            "write_fail",
			failedWriteFile: errGeneral,
			expectError:     errGeneral,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetVars()
			errMarshal = tc.failedMarshal
			errWriteFile = tc.failedWriteFile

			err := updatePipelineRemoteConfig(tc.pathConfig, tc.siteURL, tc.latestTime, &pipelineRemoteMockerTest{})
			assert.Equal(t, tc.expectError, err, "updatePipelineRemoteConfig found error: %v", err)
			assert.Equal(t, tc.expect, writeFileData, "updatePipelineRemoteConfig not equal!")
		})
	}
}

// go test -v -timeout 30s -run ^TestConvertContentMapToThreeMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestConvertContentMapToThreeMap(t *testing.T) {
	cases := []struct {
		name   string
		in     map[string]string
		expect map[string]map[string]string
	}{
		{
			name: "new",
			in: map[string]string{
				"metric/123.p":  "text123",
				"logging/456.p": "text456",
			},
			expect: map[string]map[string]string{
				"metric": {
					"123.p": "text123",
				},
				"logging": {
					"456.p": "text456",
				},
			},
		},
		{
			name: "old",
			in: map[string]string{
				"123.p": "text123",
				"456.p": "text456",
			},
			expect: map[string]map[string]string{
				".": {
					"123.p": "text123",
					"456.p": "text456",
				},
			},
		},
		{
			name: "append",
			in: map[string]string{
				"metric/123.p":   "text123",
				"logging/456.p":  "text456",
				"metric/1234.p":  "text1234",
				"logging/123.p":  "text123",
				"metric/12345.p": "text12345",
			},
			expect: map[string]map[string]string{
				"metric": {
					"123.p":   "text123",
					"1234.p":  "text1234",
					"12345.p": "text12345",
				},
				"logging": {
					"456.p": "text456",
					"123.p": "text123",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ConvertContentMapToThreeMap(tc.in)
			assert.Equal(t, tc.expect, out)
		})
	}
}

// go test -v -timeout 30s -run ^TestConvertThreeMapToContentMap$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/remote
func TestConvertThreeMapToContentMap(t *testing.T) {
	cases := []struct {
		name      string
		in        map[string]map[string]string
		inDefault map[string]string
		expect    map[string]string
	}{
		{
			name: "normal",
			in: map[string]map[string]string{
				"logging": {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				"metric": {
					"456.p": "text456",
				},
			},
			inDefault: map[string]string{
				"logging": "123.p",
			},
			expect: map[string]string{
				"logging/123.p":         "text123",
				"logging/1234.p":        "text1234",
				"metric/456.p":          "text456",
				"category_default.json": "{\"logging\":\"123.p\"}",
			},
		},
		{
			name: "normal1",
			in: map[string]map[string]string{
				"logging": {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				"metric": {
					"456.p": "text456",
				},
			},
			expect: map[string]string{
				"logging/123.p":  "text123",
				"logging/1234.p": "text1234",
				"metric/456.p":   "text456",
			},
		},
		{
			name: "normal2",
			in: map[string]map[string]string{
				"logging": {
					"123.p":  "text123",
					"1234.p": "text1234",
				},
				"metric": {
					"456.p": "text456",
				},
			},
			inDefault: map[string]string{},
			expect: map[string]string{
				"logging/123.p":         "text123",
				"logging/1234.p":        "text1234",
				"metric/456.p":          "text456",
				"category_default.json": "{}",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := convertThreeMapToContentMap(tc.in, tc.inDefault)
			assert.Equal(t, tc.expect, out)
		})
	}
}
