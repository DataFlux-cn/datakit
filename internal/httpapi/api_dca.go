// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var dcaErrorMessage = map[string]string{
	"server.error": "server error",
}

func getBody(c *gin.Context, data interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
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

var errDCAReloadError = dcaError{
	ErrorCode: "system.reload.error",
	ErrorMsg:  "reload datakit error",
}

// dca reload.
func dcaReload(c *gin.Context) {
	dcaCtx := getContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := dcaAPI.ReloadDataKit(ctx); err != nil {
		l.Error("reloadDataKit: %s", err)
		dcaCtx.fail(errDCAReloadError)
		return
	}

	dcaCtx.success()
}

// ReloadDataKit will reload datakit modules wihout restart datakit process.
func ReloadDataKit(ctx context.Context) error {
	round := 0 // 循环次数
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("reload timeout")

		default:
			switch round {
			case 0:
				l.Info("before ReloadCheckInputCfg")

				_, err := config.ReloadCheckInputCfg()
				if err != nil {
					l.Errorf("ReloadCheckInputCfg failed: %v", err)
					return err
				}

				l.Info("before ReloadCheckPipelineCfg")

			case 1:
				l.Info("before StopInputs")

				if err := inputs.StopInputs(); err != nil {
					l.Errorf("StopInputs failed: %v", err)
					return err
				}

			case 2:
				l.Info("before ReloadInputConfig")

				if err := config.ReloadInputConfig(); err != nil {
					l.Errorf("ReloadInputConfig failed: %v", err)
					return err
				}

			case 3:
				l.Info("before set pipelines")
				if managerwkr, ok := plval.GetManager(); ok && managerwkr != nil {
					manager.LoadScripts2StoreFromPlStructPath(managerwkr,
						manager.GitRepoScriptNS,
						filepath.Join(datakit.GitReposRepoFullPath, "pipeline"), nil)
				}

			case 4:
				l.Info("before RunInputs")

				CleanHTTPHandler()
				if err := inputs.RunInputs(); err != nil {
					l.Errorf("RunInputs failed: %v", err)
					return err
				}

			case 5:
				l.Info("before ReloadTheNormalServer")

				ReloadTheNormalServer(
					WithAPIConfig(config.Cfg.HTTPAPI),
					WithDCAConfig(config.Cfg.DCAConfig),
					WithGinLog(config.Cfg.Logging.GinLog),
					WithGinRotateMB(config.Cfg.Logging.Rotate),
					WithGinReleaseMode(strings.ToLower(config.Cfg.Logging.Level) != "debug"),
					WithDataway(config.Cfg.Dataway),
					WithPProf(config.Cfg.EnablePProf),
					WithPProfListen(config.Cfg.PProfListen),
				)
			}
		}

		round++
		if round > 6 {
			return nil
		}
	}
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
func dcaAuthMiddleware(tkns []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := c.FullPath()
		for _, uri := range ignoreAuthURI {
			if uri == fullPath {
				c.Next()
				return
			}
		}

		tokens := c.Request.Header["X-Token"]

		l.Debugf("request tokens: %+#v, local tokens: %+#v", tokens, tkns)

		context := &dcaContext{c: c}
		if len(tokens) == 0 {
			context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
			c.Abort()
			return
		}

		token := tokens[0]
		if len(token) == 0 || len(tkns) == 0 || (token != tkns[0]) {
			context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func dcaGetConfig(c *gin.Context) {
	context := getContext(c)
	path := c.Query("path")

	if errMsg, err := checkPath(path); err != nil {
		context.fail(errMsg)
		return
	}

	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf("Read config file %s error: %s", path, err.Error())
		context.fail(dcaError{ErrorCode: "invalid.path", ErrorMsg: "invalid path"})
		return
	}
	context.success(string(content))
}

func dcaDeleteConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close() //nolint:errcheck

	context := &dcaContext{c: c}
	if err != nil {
		l.Errorf("Read request body error: %s", err.Error())
		context.fail()
		return
	}

	param := &struct {
		Path      string `json:"path"`
		InputName string `json:"inputName"`
	}{}

	if err := json.Unmarshal(body, &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		context.fail(dcaError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if errMsg, err := checkPath(param.Path); err != nil {
		context.fail(errMsg)
		return
	}

	if !path.IsFileExists(param.Path) {
		context.fail(dcaError{Code: 400, ErrorCode: "file.path.invalid", ErrorMsg: "The file to be deleted is not existed!"})
		return
	}

	if err := os.Remove(param.Path); err != nil {
		context.fail(dcaError{Code: 500, ErrorCode: "file.delete.failed", ErrorMsg: "Fail to delete conf file"})
		l.Errorf("Delete conf file [%s] failed, %s", param.Path, err.Error())
	} else {
		inputs.DeleteConfigInfoPath(param.InputName, param.Path)
		context.success()
	}
}

// save config.
func dcaSaveConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)

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

			if content, err = os.ReadFile(param.Path); err != nil {
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
	err = os.WriteFile(param.Path, configContent, datakit.ConfPerm)
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
	pullFileBytes, err := os.ReadFile(pullFilePath) //nolint: gosec
	if err != nil {
		context.success(filterInfo{Content: "", FilePath: ""})
		return
	}

	context.success(filterInfo{Content: string(pullFileBytes), FilePath: pullFilePath})
}

func dcaDownloadLog(c *gin.Context) {
	logType := "log"
	context := getContext(c)
	logFile := config.Cfg.Logging.Log
	if c.Query("type") == "gin.log" {
		logFile = config.Cfg.Logging.GinLog
		logType = "gin.log"
	}

	file, err := os.Open(filepath.Clean(logFile))
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
	logFile := config.Cfg.Logging.Log
	context := getContext(c)

	if c.Query("type") == "gin.log" {
		logFile = config.Cfg.Logging.GinLog
	}

	file, err := os.Open(filepath.Clean(logFile))
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
	Category string `json:"category"`
}

func dcaGetPipelines(c *gin.Context) {
	context := getContext(c)

	allFiles, err := os.ReadDir(datakit.PipelineDir)
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
		} else {
			pipelines = append(pipelines, pipelineInfo{Category: file.Name()})
			allFiles, err := os.ReadDir(filepath.Join(datakit.PipelineDir, file.Name()))
			if err != nil {
				context.fail()
				return
			}
			for _, subFile := range allFiles {
				if !subFile.IsDir() {
					name := subFile.Name()
					if isValidPipelineFileName(name) {
						pipelines = append(pipelines,
							pipelineInfo{
								FileName: name,
								FileDir:  filepath.Join(datakit.PipelineDir, file.Name()),
								Category: file.Name(),
							},
						)
					}
				}
			}
		}
	}

	context.success(pipelines)
}

type pipelineDetailResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func dcaGetPipelinesDetail(c *gin.Context) {
	context := getContext(c)
	fileName := c.Query("fileName")
	if len(fileName) == 0 {
		context.fail(dcaError{ErrorCode: "params.required", ErrorMsg: fmt.Sprintf("param %s is required", "fileName")})
		return
	}

	category := c.Query("category")

	if !isValidPipelineFileName(fileName) {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	path := filepath.Join(datakit.PipelineDir, category, fileName)

	contentBytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf("Read pipeline file %s failed: %s", path, err.Error())
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	context.success(pipelineDetailResponse{
		Path:    path,
		Content: string(contentBytes),
	})
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

	if pipeline.Category == "default" {
		pipeline.Category = ""
	}

	fileName = pipeline.FileName
	category := pipeline.Category
	filePath = filepath.Join(datakit.PipelineDir, category, fileName)

	if isUpdate && len(pipeline.Content) == 0 {
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

	if !isValidPipelineFileName(fileName) {
		context.fail(dcaError{
			ErrorCode: "param.invalid",
			ErrorMsg:  "fileName is not valid pipeline name",
		})
		return
	}

	err = os.WriteFile(filePath, []byte(pipeline.Content), datakit.ConfPerm)

	if err != nil {
		l.Errorf("Write pipeline file %s failed: %s", filePath, err.Error())
		context.fail()
		return
	}
	pipeline.FileDir = datakit.PipelineDir
	context.success(pipeline)
}

func dcaDeletePipelines(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close() //nolint:errcheck

	context := &dcaContext{c: c}
	if err != nil {
		l.Errorf("Read request body error: %s", err.Error())
		context.fail()
		return
	}

	param := &struct {
		Category string `json:"category"`
		FileName string `json:"fileName"`
	}{}

	if err := json.Unmarshal(body, &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		context.fail(dcaError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if param.Category == "default" {
		param.Category = ""
	}

	fp := filepath.Join(datakit.PipelineDir, param.Category, param.FileName)

	if !path.IsFileExists(fp) {
		context.fail(dcaError{Code: 400, ErrorCode: "file.path.invalid", ErrorMsg: "The file to be deleted is not existed!"})
		return
	}

	if err := os.Remove(fp); err != nil {
		context.fail(dcaError{Code: 500, ErrorCode: "file.delete.failed", ErrorMsg: "Fail to delete conf file"})
		l.Errorf("Delete pipeline file [%s] failed, %s", fp, err.Error())
	} else {
		context.success()
	}
}

func pipelineTest(pipelineFile string, text string) (string, error) {
	// TODO
	pl, err := pipeline.NewPlScriptSampleFromFile(point.Logging, filepath.Join(datakit.PipelineDir, pipelineFile))
	if err != nil {
		return "", err
	}

	kvs := point.NewTags(datakit.GlobalHostTags())
	kvs = append(kvs, point.NewKVs(map[string]interface{}{pipeline.FieldMessage: text})...)
	opt := point.DefaultLoggingOptions()
	pt := point.NewPointV2("default", kvs, opt...)

	plpt := ptinput.WrapPoint(point.Logging, pt)
	err = pl.Run(plpt, nil, nil)
	if err != nil {
		return "", err
	}

	if plpt == nil {
		l.Debug("No data extracted from pipeline")
		return "", nil
	}

	if plpt.Dropped() {
		l.Debug("the current log has been dropped by the pipeline script")
		return "", nil
	}

	plpt.KeyTime2Time()

	res := pipeline.Result{
		Output: &pipeline.Output{
			Drop:        plpt.Dropped(),
			Measurement: plpt.GetPtName(),
			Time:        plpt.PtTime(),
			Tags:        plpt.Tags(),
			Fields:      plpt.Fields(),
		},
	}
	if j, err := json.Marshal(res); err != nil {
		return "", err
	} else {
		return string(j), nil
	}
}

type dcaTestParam struct {
	Pipeline   map[string]map[string]string `json:"pipeline"`
	ScriptName string                       `json:"script_name"`
	Category   string                       `json:"category"`
	Data       []string                     `json:"data"`
}

func dcaTestPipelines(c *gin.Context) {
	context := getContext(c)

	body := dcaTestParam{}
	if err := getBody(c, &body); err != nil {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter format error"})
		return
	}

	if len(body.Category) == 0 {
		body.Category = "logging"
	}

	// deal with default
	if body.Category == "default" {
		body.Pipeline["logging"] = body.Pipeline["default"]
		body.Category = "logging"
	}

	category := point.CatString(body.Category)

	pls, errs := pipeline.NewPipelineMulti(category, body.Pipeline[body.Category], nil, nil)
	if err, ok := errs[body.ScriptName]; ok && err != nil {
		context.fail(dcaError{ErrorCode: "400", ErrorMsg: fmt.Sprintf("pipeline parse error: %s", err.Error())})
		return
	}

	pl, ok := pls[body.ScriptName]

	if !ok {
		context.fail(dcaError{ErrorCode: "400", ErrorMsg: "pipeline is not valid"})
		return
	}

	var pts []*point.Point
	var pointErr error

	dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
	defer point.PutDecoder(dec)

	for _, data := range body.Data {
		switch category {
		case point.Logging:
			kvs := point.NewTags(datakit.GlobalHostTags())
			kvs = append(kvs, point.NewKVs(map[string]interface{}{
				pipeline.FieldMessage: data,
			})...)
			pts = append(pts, point.NewPointV2(
				body.ScriptName, kvs, point.DefaultLoggingOptions()...))

		case point.CustomObject,
			point.DynamicDWCategory,
			point.KeyEvent,
			point.MetricDeprecated,
			point.Metric,
			point.Network,
			point.Object,
			point.Profiling,
			point.RUM,
			point.Security,
			point.Tracing,
			point.UnknownCategory:

			arr, err := dec.Decode([]byte(data))
			if err != nil {
				l.Warnf("make point error: %s", err.Error())
				pointErr = err
				break
			}
			pts = append(pts, arr...)
		}
	}

	if pointErr != nil {
		context.fail(dcaError{ErrorCode: "400", ErrorMsg: fmt.Sprintf("invalid sample: %s", pointErr.Error())})
		return
	}

	var runResult []*pipelineResult

	for _, pt := range pts {
		plpt := ptinput.WrapPoint(category, pt)
		err := pl.Run(plpt, newPlTestSingal(), nil)
		if err != nil {
			plerr, ok := err.(*errchain.PlError) //nolint:errorlint
			if !ok {
				plerr = errchain.NewErr(body.ScriptName+".p", token.LnColPos{
					Pos: 0,
					Ln:  1,
					Col: 1,
				}, err.Error())
			}
			runResult = append(runResult, &pipelineResult{
				RunError: plerr,
			})
		} else {
			dropFlag := plpt.Dropped()

			plpt.KeyTime2Time()
			runResult = append(runResult, &pipelineResult{
				Point: &PlRetPoint{
					Dropped: dropFlag,
					Name:    plpt.GetPtName(),
					Tags:    plpt.Tags(),
					Fields:  plpt.Fields(),
					Time:    plpt.PtTime().Unix(),
					TimeNS:  int64(plpt.PtTime().Nanosecond()),
				},
			})
		}
	}
	context.success(&runResult)
}
