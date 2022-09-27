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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	platformWeb     = "web"
	platformAndroid = "android"
	platformIOS     = "ios"
)

var dcaErrorMessage = map[string]string{
	"server.error": "server error",
}

func getBody(c *gin.Context, data interface{}) error {
	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close() //nolint:errcheck
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, data)

	if err != nil {
		return err
	}

	return nil
}

func dcaGetMessage(errCode string) string {
	if errMsg, ok := dcaErrorMessage[errCode]; ok {
		return errMsg
	} else {
		return "server error"
	}
}

type dcaResponse struct {
	Code      int         `json:"code"`
	Content   interface{} `json:"content"`
	ErrorCode string      `json:"errorCode"`
	ErrorMsg  string      `json:"errorMsg"`
	Success   bool        `json:"success"`
}

type dcaError struct {
	Code      int
	ErrorCode string
	ErrorMsg  string
}

type dcaContext struct {
	c    *gin.Context
	data interface{}
}

func (d *dcaContext) send(response dcaResponse) {
	body, err := json.Marshal(response)
	if err != nil {
		d.fail(dcaError{})
		return
	}

	status := d.c.Writer.Status()

	d.c.Data(status, "application/json", body)
}

func (d *dcaContext) success(datas ...interface{}) {
	var data interface{}

	if len(datas) > 0 {
		data = datas[0]
	}

	if data == nil {
		data = d.data
	}

	response := dcaResponse{
		Code:    200,
		Content: data,
		Success: true,
	}

	d.send(response)
}

func (d *dcaContext) fail(errors ...dcaError) {
	var e dcaError
	if len(errors) > 0 {
		e = errors[0]
	} else {
		e = dcaError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "server.error",
			ErrorMsg:  "",
		}
	}

	code := e.Code
	errorCode := e.ErrorCode
	errorMsg := e.ErrorMsg

	if code == 0 {
		code = http.StatusInternalServerError
	}

	if errorCode == "" {
		errorCode = "server.error"
	}

	if errorMsg == "" {
		errorMsg = dcaGetMessage(errorCode)
	}

	response := dcaResponse{
		Code:      code,
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
		Success:   false,
	}

	d.send(response)
}

// dca reload.
func dcaReload(c *gin.Context) {
	context := getContext(c)
	if err := dcaAPI.RestartDataKit(); err != nil {
		l.Error("restartDataKit: %s", err)
		context.fail(dcaError{ErrorCode: "system.restart.error", ErrorMsg: "restart datakit error"})
		return
	}

	context.success()
}

func restartDataKit() error {
	bin := "datakit"
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	program := filepath.Join(datakit.InstallDir, bin)
	l.Info("apiRestart", program)
	cmd := exec.Command(program, "--api-restart") //nolint:gosec
	return cmd.Start()
}

type dcastats struct {
	*DatakitStats
	ConfigInfo map[string]*inputs.Config `json:"config_info"`
}

func dcaStats(c *gin.Context) {
	s, err := dcaAPI.GetStats()
	context := dcaContext{c: c}
	stats := &dcastats{
		DatakitStats: s,
		ConfigInfo:   inputs.ConfigInfo,
	}
	if err != nil {
		context.fail()
		return
	}

	context.success(stats)
}

func dcaStatsByType(c *gin.Context) {
	var stat interface{}

	context := dcaContext{c: c}

	statType := c.Param("type")
	if statType == "" {
		statType = StatInfoType
	}

	if statType == StatInfoType {
		stat = getStatInfo()
	} else if statType == StatMetricType {
		stat = getStatMetric()
	}

	if stat == nil {
		context.fail(dcaError{
			Code:      400,
			ErrorCode: "param.invalid",
			ErrorMsg:  fmt.Sprintf("invalid type, which should be '%s' or '%s'", StatInfoType, StatMetricType),
		})
		return
	}

	context.success(stat)
}

func dcaDefault(c *gin.Context) {
	context := dcaContext{c: c}
	context.c.Status(404)
	context.fail(dcaError{Code: 404, ErrorCode: "route.not.found", ErrorMsg: "route not found"})
}

type saveConfigParam struct {
	Path      string `json:"path"`
	Config    string `json:"config"`
	IsNew     bool   `json:"isNew"`
	InputName string `json:"inputName"`
}

// auth middleware.
func dcaAuthMiddleware(c *gin.Context) {
	fullPath := c.FullPath()
	for _, uri := range ignoreAuthURI {
		if uri == fullPath {
			c.Next()
			return
		}
	}
	tokens := c.Request.Header["X-Token"]
	context := &dcaContext{c: c}
	if len(tokens) == 0 {
		context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
		c.Abort()
		return
	}

	token := tokens[0]
	localTokens := dw.GetTokens()
	if len(token) == 0 || len(localTokens) == 0 || (token != localTokens[0]) {
		context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
		c.Abort()
		return
	}
	c.Next()
}

func dcaGetConfig(c *gin.Context) {
	context := getContext(c)
	path := c.Query("path")

	if errMsg, err := checkPath(path); err != nil {
		context.fail(errMsg)
		return
	}

	content, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf("Read config file %s error: %s", path, err.Error())
		context.fail(dcaError{ErrorCode: "invalid.path", ErrorMsg: "invalid path"})
		return
	}
	context.success(string(content))
}

// save config.
func dcaSaveConfig(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)

	defer c.Request.Body.Close() //nolint:errcheck

	context := &dcaContext{c: c}
	if err != nil {
		l.Errorf("Read request body error: %s", err.Error())
		context.fail()
		return
	}

	param := saveConfigParam{}

	if err := json.Unmarshal(body, &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		context.fail()
		return
	}

	if errMsg, err := checkPath(param.Path); err != nil {
		context.fail(errMsg)
		return
	}

	configContent := []byte(param.Config)

	// add new config
	if param.IsNew {
		if _, err := os.Stat(param.Path); err == nil { // exist
			var content []byte
			var err error

			if content, err = ioutil.ReadFile(param.Path); err != nil {
				l.Errorf("Read file %s error: %s", param.Path, err.Error())
				context.fail()
				return
			}
			configContent = append(content, configContent...)
		}
	}

	// check toml
	_, err = toml.Parse(configContent)
	if err != nil {
		l.Errorf("parse toml %s failed", string(configContent))
		context.fail(dcaError{ErrorCode: "toml.format.error", ErrorMsg: "toml format error"})
		return
	}

	// create and save
	err = ioutil.WriteFile(param.Path, configContent, datakit.ConfPerm)
	if err != nil {
		l.Errorf("Write file %s failed: %s", param.Path, err.Error())
		context.fail()
		return
	}

	// update configInfo
	if len(param.InputName) > 0 {
		if c, ok := inputs.ConfigInfo[param.InputName]; ok { // update
			existed := false
			for _, configPath := range c.ConfigPaths {
				if configPath.Path == param.Path {
					existed = true
					configPath.Loaded = int8(2)
					break
				}
			}
			if !existed {
				c.ConfigPaths = append(c.ConfigPaths, &inputs.ConfigPathStat{
					Loaded: int8(2),
					Path:   param.Path,
				})
			}
		} else if creator, ok := inputs.Inputs[param.InputName]; ok { // add new info
			inputs.ConfigInfo[param.InputName] = &inputs.Config{
				ConfigPaths: []*inputs.ConfigPathStat{
					{
						Loaded: int8(2),
						Path:   param.Path,
					},
				},
				SampleConfig: creator().SampleConfig(),
				Catalog:      creator().Catalog(),
				ConfigDir:    datakit.ConfdDir,
			}
		}
	}

	context.success(map[string]string{
		"path": param.Path,
	})
}

func dcaInputDoc(c *gin.Context) {
	context := getContext(c)
	inputName := c.Query("inputName")
	l.Debug("api_doc", inputName)
	if inputName == "" {
		context.fail()
		return
	}

	md, err := dcaAPI.GetMarkdownContent(inputName)
	if err != nil {
		l.Warn(err)
		// context.fail(dcaError{ErrorCode: "record.not.exist", ErrorMsg: "record not exist", Code: http.StatusNotFound})
		context.success("")
		return
	}

	context.success(string(md))
}

func getContext(c *gin.Context) dcaContext {
	return dcaContext{c: c}
}

func checkPath(path string) (dcaError, error) {
	err := errors.New("invalid conf")

	// path should under conf.d
	dir := filepath.Dir(path)

	pathReg := regexp.MustCompile(`\.conf$`)

	if pathReg == nil {
		return dcaError{}, err
	}

	// check path
	if !strings.Contains(path, datakit.ConfdDir) || !pathReg.Match([]byte(path)) {
		return dcaError{ErrorCode: "params.invalid.path_invalid", ErrorMsg: "invalid param 'path'"}, err
	}

	// check dir
	if _, err := os.Stat(dir); err != nil {
		return dcaError{ErrorCode: "params.invalid.dir_not_exist", ErrorMsg: "dir not exist"}, err
	}

	return dcaError{}, nil
}

func isValidPipelineFileName(name string) bool {
	pipelineFileRegxp := regexp.MustCompile(`.+\.p$`)

	return pipelineFileRegxp.Match([]byte(name))
}

func isValidCustomPipelineName(name string) bool {
	pipelineFileRegxp := regexp.MustCompile(`^custom_.+\.p$`)

	return pipelineFileRegxp.Match([]byte(name))
}

// filter info.
type filterInfo struct {
	Content  string `json:"content"`  // file content string
	FilePath string `json:"filePath"` // file path
}

// dcaGetFilter return filter file content, which is located at data/.pull
// if the file not existed, return empty content.
func dcaGetFilter(c *gin.Context) {
	context := getContext(c)
	dataDir := datakit.DataDir
	pullFilePath := filepath.Join(dataDir, ".pull")
	pullFileBytes, err := ioutil.ReadFile(pullFilePath) //nolint: gosec
	if err != nil {
		context.success(filterInfo{Content: "", FilePath: ""})
		return
	}

	context.success(filterInfo{Content: string(pullFileBytes), FilePath: pullFilePath})
}

func dcaDownloadLog(c *gin.Context) {
	logType := "log"
	context := getContext(c)
	logFile := log
	if c.Query("type") == "gin.log" {
		logFile = ginLog
		logType = "gin.log"
	}

	file, err := os.Open(logFile)
	if err != nil {
		l.Errorf("DCA open log file %s failed: %s", logFile, err.Error())
		c.Status(400)
		context.fail(dcaError{ErrorCode: "dca.log.file.invalid", ErrorMsg: "datakit log file is not valid"})
		return
	}
	defer file.Close() //nolint: errcheck,gosec
	var fileSize int64 = 0
	if info, err := file.Stat(); err == nil {
		fileSize = info.Size()
	}

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`"attachment; filename="%s"`, logType),
	}

	c.DataFromReader(http.StatusOK, fileSize, "application/octet-stream", file, extraHeaders)
}

func dcaGetLogTail(c *gin.Context) {
	l.Info("dcaGetLogTail start")
	logFile := log
	context := getContext(c)

	if c.Query("type") == "gin.log" {
		logFile = ginLog
	}

	file, err := os.Open(logFile)
	if err != nil {
		l.Errorf("DCA open log file %s failed: %s", logFile, err.Error())
		c.Status(400)
		context.fail(dcaError{ErrorCode: "dca.log.file.invalid", ErrorMsg: "datakit log file is not valid"})
		return
	}
	defer file.Close() //nolint: errcheck,gosec

	var fileSize int64 = 0
	if info, err := file.Stat(); err == nil {
		fileSize = info.Size()
	}

	var fileSeek int64
	if fileSize >= 2000 {
		fileSeek = -2000
	} else {
		fileSeek = -1 * fileSize
	}

	buffer := make([]byte, 1024)

	_, err = file.Seek(fileSeek, io.SeekEnd)

	if err != nil {
		l.Errorf("Seek offset %v error: %s", fileSeek, err.Error())
		c.Status(400)
		context.fail(dcaError{ErrorCode: "dca.log.file.seek.error", ErrorMsg: "seek datakit log file error"})
		return
	}
	c.Status(202)
	c.Stream(func(w io.Writer) bool {
		n, err := file.Read(buffer)
		if errors.Is(err, io.EOF) {
			time.Sleep(5 * time.Second)
			return true
		}
		if err != nil {
			return false
		}
		if n > 0 {
			_, err = w.Write(buffer[0:n])
			if err != nil {
				return false
			}
		} else {
			time.Sleep(5 * time.Second)
		}

		return true
	})

	l.Info("dcaGetLogTail end")
}

type pipelineInfo struct {
	FileName string `json:"fileName"`
	FileDir  string `json:"fileDir"`
	Content  string `json:"content"`
}

func dcaGetPipelines(c *gin.Context) {
	context := getContext(c)

	allFiles, err := ioutil.ReadDir(datakit.PipelineDir)
	if err != nil {
		context.fail()
		return
	}

	pipelines := []pipelineInfo{}

	// filter pipeline file
	for _, file := range allFiles {
		if !file.IsDir() {
			name := file.Name()
			if isValidPipelineFileName(name) {
				pipelines = append(pipelines, pipelineInfo{FileName: name, FileDir: datakit.PipelineDir})
			}
		} else if file.Name() == "logging" {
			allFiles, err := ioutil.ReadDir(filepath.Join(datakit.PipelineDir, "logging"))
			if err != nil {
				context.fail()
				return
			}
			for _, file := range allFiles {
				if !file.IsDir() {
					name := file.Name()
					if isValidPipelineFileName(name) {
						pipelines = append(pipelines, pipelineInfo{FileName: "logging/" + name, FileDir: datakit.PipelineDir})
					}
				}
			}
		}
	}

	context.success(pipelines)
}

func dcaGetPipelinesDetail(c *gin.Context) {
	context := getContext(c)
	fileName := c.Query("fileName")
	if len(fileName) == 0 {
		context.fail(dcaError{ErrorCode: "params.required", ErrorMsg: fmt.Sprintf("param %s is required", "fileName")})
		return
	}

	if !isValidPipelineFileName(fileName) {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	path := filepath.Join(datakit.PipelineDir, fileName)

	contentBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf("Read pipeline file %s failed: %s", path, err.Error())
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	context.success(string(contentBytes))
}

func dcaCreatePipeline(c *gin.Context) {
	dcaSavePipeline(c, false)
}

func dcaUpdatePipeline(c *gin.Context) {
	dcaSavePipeline(c, true)
}

func dcaSavePipeline(c *gin.Context, isUpdate bool) {
	var filePath string
	var fileName string
	context := getContext(c)

	pipeline := pipelineInfo{}
	err := getBody(c, &pipeline)
	if err != nil {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "make sure parameter is in correct format",
		})
		return
	}

	fileName = pipeline.FileName
	filePath = filepath.Join(datakit.PipelineDir, fileName)

	if len(pipeline.Content) == 0 {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "'content' should not be empty",
		})
		return
	}

	// update pipeline
	if isUpdate {
		if !path.IsFileExists(filePath) {
			context.fail(dcaError{ErrorCode: "file.not.exist", ErrorMsg: "file not exist"})
			return
		}
	} else { // new pipeline
		if path.IsFileExists(filePath) {
			context.fail(dcaError{
				ErrorCode: "param.invalid.duplicate",
				ErrorMsg:  fmt.Sprintf("current name '%s' is duplicate", pipeline.FileName),
			})
			return
		}
	}

	// only save or update custom pipeline!
	if !isValidCustomPipelineName(fileName) {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "fileName is not valid custom pipeline name",
		})
		return
	}

	err = ioutil.WriteFile(filePath, []byte(pipeline.Content), datakit.ConfPerm)

	if err != nil {
		l.Errorf("Write pipeline file %s failed: %s", filePath, err.Error())
		context.fail()
		return
	}
	pipeline.FileDir = datakit.PipelineDir
	context.success(pipeline)
}

func pipelineTest(pipelineFile string, text string) (string, error) {
	// TODO
	pl, err := pipeline.NewPipelineFromFile(datakit.Logging, filepath.Join(datakit.PipelineDir, pipelineFile))
	if err != nil {
		return "", err
	}
	opt := &point.PointOption{
		Category: datakit.Logging,
		Time:     time.Now(),
	}
	pt, err := point.NewPoint("default", nil, map[string]interface{}{pipeline.FieldMessage: text}, opt)
	if err != nil {
		return "", err
	}

	pt, dropFlag, err := pl.Run(pt, nil, opt)
	if err != nil {
		return "", err
	}

	if pt == nil {
		l.Debug("No data extracted from pipeline")
		return "", nil
	}

	fields := pt.Fields
	tags := pt.Tags

	if dropFlag {
		l.Debug("the current log has been dropped by the pipeline script")
		return "", nil
	}

	res := pipeline.Result{
		Output: &pipeline.Output{
			Drop:        dropFlag,
			Measurement: pt.Name,
			Time:        pt.Time,
			Tags:        tags,
			Fields:      fields,
		},
	}
	if j, err := json.Marshal(res); err != nil {
		return "", err
	} else {
		return string(j), nil
	}
}

func dcaTestPipelines(c *gin.Context) {
	context := getContext(c)

	body := map[string]string{}

	if err := getBody(c, &body); err != nil {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter format error"})
		return
	}

	text, ok := body["text"]
	if !ok {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter 'text' is required"})
		return
	}

	fileName, ok := body["fileName"]
	if !ok {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter 'fileName' is required"})
		return
	}

	parsedText, err := dcaAPI.TestPipeline(fileName, text)
	if err != nil {
		l.Errorf("Test pipeline error: %s", err.Error())
		context.fail(dcaError{ErrorCode: "pipeline.parse.error", ErrorMsg: err.Error()})
		return
	}

	context.success(parsedText)
}

type rumQueryParam struct {
	ApplicationID string `form:"app_id"`
	Env           string `form:"env"`
	Version       string `form:"version"`
	Platform      string `form:"platform"`
}

// upload sourcemap
// curl -X POST 'http://localhost:9531/v1/rum/sourcemap?app_id=app_xxxx&env=release&version=1.0.1'
// 			-F "file=@/tmp/code.zip"
// 			-H "Content-Type: multipart/form-data".
func dcaUploadSourcemap(c *gin.Context) {
	context := getContext(c)

	var param rumQueryParam

	if c.ShouldBindQuery(&param) != nil {
		context.fail(dcaError{ErrorCode: "query.parse.error", ErrorMsg: "query string parse error"})
		return
	}

	if (len(param.ApplicationID) == 0) || (len(param.Env) == 0) || (len(param.Version) == 0) {
		context.fail(dcaError{ErrorCode: "query.param.required", ErrorMsg: "app_id, env, version required"})
		return
	}

	if param.Platform == "" {
		param.Platform = platformWeb
	}

	if param.Platform != platformWeb && param.Platform != platformAndroid && param.Platform != platformIOS {
		l.Errorf("platform [%s] not supported", param.Platform)
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  fmt.Sprintf("platform [%s] not supported, please use web, android or ios", param.Platform),
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		l.Errorf("get file failed: %s", err.Error())
		context.fail(dcaError{ErrorCode: "upload.file.required", ErrorMsg: "make sure the file was uploaded correctly"})
		return
	}

	fileName := GetSourcemapZipFileName(param.ApplicationID, param.Env, param.Version)
	rumDir := filepath.Join(GetRumSourcemapDir(), param.Platform)
	if !path.IsDir(rumDir) {
		if err := os.MkdirAll(rumDir, datakit.ConfPerm); err != nil {
			context.fail(dcaError{
				ErrorCode: "dir.create.failed",
				ErrorMsg:  "rum dir created failed",
			})
			return
		}
	}
	dst := filepath.Clean(filepath.Join(rumDir, fileName))

	// check filename
	if !strings.HasPrefix(dst, rumDir) {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "invalid param, should not contain illegal char, such as  '../, /'",
		})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		l.Errorf("save upload file error: %s", err.Error())
		context.fail(dcaError{ErrorCode: "upload.file.error", ErrorMsg: "upload failed"})
		return
	}
	updateSourcemapCache(dst)
	context.success(fmt.Sprintf("uploaded to %s!", fileName))
}

func dcaDeleteSourcemap(c *gin.Context) {
	context := getContext(c)

	var param rumQueryParam

	if c.ShouldBindQuery(&param) != nil {
		context.fail(dcaError{ErrorCode: "query.parse.error", ErrorMsg: "query string parse error"})
		return
	}

	if (len(param.ApplicationID) == 0) || (len(param.Env) == 0) || (len(param.Version) == 0) {
		context.fail(dcaError{ErrorCode: "query.param.required", ErrorMsg: "app_id, env, version required"})
		return
	}

	if param.Platform == "" {
		param.Platform = platformWeb
	}

	if param.Platform != platformWeb && param.Platform != platformAndroid && param.Platform != platformIOS {
		l.Errorf("platform [%s] not supported", param.Platform)
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  fmt.Sprintf("platform [%s] not supported, please use web, android or ios", param.Platform),
		})
		return
	}

	fileName := GetSourcemapZipFileName(param.ApplicationID, param.Env, param.Version)
	rumDir := filepath.Join(GetRumSourcemapDir(), param.Platform)
	zipFilePath := filepath.Clean(filepath.Join(rumDir, fileName))

	// check filename
	if !strings.HasPrefix(zipFilePath, rumDir) {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "invalid param, should not contain illegal char, such as  '../, /'",
		})
		return
	}

	if err := os.Remove(zipFilePath); err != nil {
		l.Errorf("delete zip file failed: %s, %s", zipFilePath, err.Error())
		context.fail(dcaError{
			ErrorCode: "delete.error",
			ErrorMsg:  "delete sourcemap file failed.",
		})
		return
	}
	deleteSourcemapCache(zipFilePath)
	context.success("delete file successfully")
}
