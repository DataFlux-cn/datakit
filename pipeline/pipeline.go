package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	influxm "github.com/influxdata/influxdb1-client/models"
	conv "github.com/spf13/cast"
	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/patterns"
)

type Pipeline struct {
	Content  string
	Output   map[string]interface{}
	lastErr  error
	patterns map[string]string //存放自定义patterns
	nodes    []parser.Node
	grok     *vgrok.Grok
}

var (
	l = logger.DefaultSLogger("process")
)

func NewPipelineByScriptPath(path string) (*Pipeline, error) {

	scriptPath := filepath.Join(datakit.PipelineDir, path)
	data, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(data))
}

func NewPipeline(script string) (*Pipeline, error) {
	p := &Pipeline{
		Output: make(map[string]interface{}),
		grok:   grokCfg,
	}

	if err := p.parseScript(script); err != nil {
		return p, err
	}

	return p, nil
}

func NewPipelineFromFile(filename string) (*Pipeline, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(b))
}

// PointToJSON, line protocol point to pipeline JSON
func (p *Pipeline) RunPoint(point influxm.Point) *Pipeline {
	defer func() {
		r := recover()
		if r != nil {
			p.lastErr = fmt.Errorf("%v", r)
		}
	}()

	m := map[string]interface{}{"measurement": string(point.Name())}

	if tags := point.Tags(); len(tags) > 0 {
		m["tags"] = map[string]string{}
		for _, tag := range tags {
			m["tags"].(map[string]string)[string(tag.Key)] = string(tag.Value)
		}
	}

	fields, err := point.Fields()
	if err != nil {
		p.lastErr = err
		return p
	}

	for k, v := range fields {
		m[k] = v
	}

	m["time"] = point.UnixNano()

	j, err := json.Marshal(m)
	if err != nil {
		p.lastErr = err
		return p
	}

	return p.Run(string(j))
}

func (p *Pipeline) Run(data string) *Pipeline {
	defer func() {
		r := recover()
		if r != nil {
			p.lastErr = fmt.Errorf("%v", r)
		}
	}()

	var err error

	p.Content = data
	p.Output = make(map[string]interface{})
	p.Output["message"] = data

	//防止脚本解析错误
	if p.lastErr != nil {
		return p
	}

	for _, node := range p.nodes {
		switch v := node.(type) {
		case *parser.FuncExpr:
			fn := strings.ToLower(v.Name)
			f, ok := funcsMap[fn]
			if !ok {
				err := fmt.Errorf("unsupported func: %v", v.Name)
				l.Error(err)
				p.lastErr = err
				return p
			}

			_, err = f(p, node)
			if err != nil {
				l.Errorf("Run func %v: %v", v.Name, err)
				p.lastErr = err
				return p
			}

		default:
			p.lastErr = fmt.Errorf("%v not function", v.String())
		}
	}
	return p
}

func (p *Pipeline) Result() (map[string]interface{}, error) {
	return p.Output, p.lastErr
}

func (p *Pipeline) LastError() error {
	return p.lastErr
}

func (p *Pipeline) getContent(key string) interface{} {
	if key == "_" {
		return p.Content
	}

	if v, ok := p.Output[key]; ok {
		return v
	}

	var m interface{}
	var nm interface{}

	m = p.Output
	keys := strings.Split(key, ".")
	for _, k := range keys {
		switch m.(type) {
		case map[string]interface{}:
			v := m.(map[string]interface{})
			nm = v[k]
			m = nm
		default:
			nm = nil
		}
	}

	return nm
}

func (p *Pipeline) getContentStr(key string) string {
	return conv.ToString(p.getContent(key))
}

func (p *Pipeline) getContentStrByCheck(key string) (string, bool) {
	v := p.getContent(key)
	if v == nil {
		return "", false
	}

	return conv.ToString(v), true
}

func (p *Pipeline) setContent(k string, v interface{}) {
	if p.Output == nil {
		p.Output = make(map[string]interface{})
	}

	if v == nil {
		return
	}

	p.Output[k] = v
}

func (pl *Pipeline) parseScript(script string) error {

	nodes, err := parser.ParseFuncExpr(script)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		switch v := node.(type) {
		case *parser.FuncExpr:
			debugNodesHelp(v, "")
		default:
			return fmt.Errorf("should not been here")
		}
	}

	pl.nodes = nodes
	return nil
}

func debugNodesHelp(f *parser.FuncExpr, prev string) {
	l.Debugf("%v%v", prev, f.Name)

	for _, node := range f.Param {
		switch v := node.(type) {
		case *parser.FuncExpr:
			debugNodesHelp(v, prev+"    ")
		default:
			l.Debugf("%v%v", prev+"    |", node)
		}
	}
}

func Init() error {
	if err := patterns.InitPatternsFile(); err != nil {
		return err
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	return nil
}
