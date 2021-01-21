package pipeline

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

type PipelineFunc func(p *Pipeline, node parser.Node) (*Pipeline, error)

var (
	funcsMap = map[string]PipelineFunc{
		"add_key":          Addkey,
		"add_pattern":      AddPattern,
		"cast":             Cast,
		"datetime":         DateTime,
		"default_time":     DefaultTime,
		"drop_key":         Dropkey,
		"drop_origin_data": DropOriginData,
		"expr":             Expr,
		"geoip":            GeoIp,
		"grok":             Grok,
		"group_between":    Group,
		"group_in":         GroupIn,
		"json":             Json,
		"json_all":         JsonAll,
		"lowercase":        Lowercase,
		"nullif":           NullIf,
		"rename":           Rename,
		"strfmt":           Strfmt,
		"uppercase":        Uppercase,
		"url_decode":       UrlDecode,
		"user_agent":       UserAgent,
	}
)

func Json(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 3 {
		return p, fmt.Errorf("func %s expected 2 or 3 args", funcExpr.Name)
	}

	var key, old parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		old = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	newkey := old
	if len(funcExpr.Param) == 3 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v, err := GsonGet(cont, old)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	err = p.setContent(newkey, v)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func JsonAll(p *Pipeline, node parser.Node) (*Pipeline, error) {
	out := JsonParse(p.Content)
	p.Output = out

	return p, nil
}

func Rename(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var old, new parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		new = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		old = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	v, err := p.getContent(old)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	err = p.setContent(new, v)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	delete(p.Output, old.String())

	return p, nil
}

func UserAgent(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	dic := UserAgentHandle(cont)

	for k, val := range dic {
		p.setContent(k, val)
	}

	return p, nil
}

func UrlDecode(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)

	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if v, err := UrldecodeHandle(cont); err != nil {
		return p, err
	} else {
		p.setContent(key, v)
	}

	return p, nil
}

func GeoIp(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func `%s' expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if dic, err := GeoIpHandle(cont); err != nil {
		return p, err
	} else {
		for k, v := range dic {
			p.setContent(k, v)
		}
	}

	return p, nil
}

func DateTime(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 3 {
		return p, fmt.Errorf("func `%s' expected 3 args", funcExpr.Name)
	}

	var key parser.Node
	var precision, fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		precision = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	switch v := funcExpr.Param[2].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[2]).String())
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if v, err := DateFormatHandle(cont, precision, fmts); err != nil {
		return p, err
	} else {
		p.setContent(key, v)
	}

	return p, nil
}

func Expr(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var expr *parser.BinaryExpr

	switch v := funcExpr.Param[0].(type) {
	case *parser.BinaryExpr:
		expr = v
	default:
		return p, fmt.Errorf("expect BinaryExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.Identifier, *parser.AttrExpr, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if v, err := Calc(expr, p); err != nil {
		l.Warn(err)
		return p, nil
	} else {
		err = p.setContent(key, v)
		if err != nil {
			l.Warn(err)
			return p, nil
		}
	}

	return p, nil
}

func Strfmt(p *Pipeline, node parser.Node) (*Pipeline, error) {
	outdata := make([]interface{}, 0)

	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 2 {
		return p, fmt.Errorf("func `%s' expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	var fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Identifier:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.AttrExpr:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.NumberLiteral:
			if v.IsInt {
				outdata = append(outdata, v.Int)
			} else {
				outdata = append(outdata, v.Float)
			}
		default:
			outdata = append(outdata, v)
		}
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	err := p.setContent(key, strfmt)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Cast(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var castType string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		castType = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	val := cast(cont, castType)
	err = p.setContent(key, val)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Group(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return p, fmt.Errorf("func `%s' expected 3 or 4 args", funcExpr.Name)
	}

	set := funcExpr.Param[1].(parser.FuncArgList)
	value := funcExpr.Param[2]

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	newkey := key
	var start, end float64

	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	if len(set) != 2 {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	}

	if v, ok := set[0].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			start = float64(v.Int)
		} else {
			start = v.Float
		}
	}

	if v, ok := set[1].(*parser.NumberLiteral); !ok {
		return p, fmt.Errorf("range value `%v' is not expected", set)
	} else {
		if v.IsInt {
			end = float64(v.Int)
		} else {
			end = v.Float
		}

		if start > end {
			return p, fmt.Errorf("range value start %v must le end %v", start, end)
		}
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if GroupHandle(cont, start, end) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				p.setContent(newkey, v.Int)
			} else {
				p.setContent(newkey, v.Float)
			}
		case *parser.StringLiteral:
			p.setContent(newkey, v.Val)
		case *parser.BoolLiteral:
			p.setContent(newkey, v.Val)
		}
	}

	return p, nil
}

func GroupIn(p *Pipeline, node parser.Node) (*Pipeline, error) {
	setdata := make([]interface{}, 0)
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 3 || len(funcExpr.Param) > 4 {
		return nil, fmt.Errorf("func %s expected 3 or 4 args", funcExpr.Name)
	}

	set := funcExpr.Param[1].(parser.FuncArgList)
	value := funcExpr.Param[2]

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	newkey := key
	if len(funcExpr.Param) == 4 {
		switch v := funcExpr.Param[3].(type) {
		case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
			newkey = v
		default:
			return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[3]).String())
		}
	}

	for _, node := range set {
		switch v := node.(type) {
		case *parser.Identifier:
			cont, err := p.getContent(v.Name)
			if err != nil {
				l.Warnf("key `%v' not exist", key)
				return p, nil
			}
			setdata = append(setdata, cont)
		case *parser.NumberLiteral:
			if v.IsInt {
				setdata = append(setdata, v.Int)
			} else {
				setdata = append(setdata, v.Float)
			}
		case *parser.BoolLiteral:
			setdata = append(setdata, v.Val)
		case *parser.StringLiteral:
			setdata = append(setdata, v.Val)
		default:
			setdata = append(setdata, v)
		}
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if GroupInHandle(cont, setdata) {
		switch v := value.(type) {
		case *parser.NumberLiteral:
			if v.IsInt {
				p.setContent(newkey, v.IsInt)
			} else {
				p.setContent(newkey, v.Float)
			}
		case *parser.StringLiteral:
			p.setContent(newkey, v.Val)
		case *parser.BoolLiteral:
			p.setContent(newkey, v.Val)
		}
	}

	return p, nil
}

func DefaultTime(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	if v, err := TimestampHandle(cont); err != nil {
		// l.Warnf("time convert fail error %v", err)
		return p, fmt.Errorf("time convert fail error %v", err)
	} else {
		p.setContent(key, v)
	}

	return p, nil
}

func Uppercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v := strings.ToUpper(cont)
	err = p.setContent(key, v)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func Lowercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v := strings.ToLower(cont)
	err = p.setContent(key, v)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func NullIf(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warnf("key `%v' not exist", key)
		return p, nil
	}

	// todo key string
	if reflect.DeepEqual(cont, val) {
		var k string

		switch t := key.(type) {
		case *parser.Identifier:
			k = t.String()
		case *parser.AttrExpr:
			k = t.String()
		case *parser.StringLiteral:
			k = t.Val
		default:
			l.Warnf("unsupported %v get", reflect.TypeOf(key).String())
			return p, nil
		}

		delete(p.Output, k)
	}

	return p, nil
}

func Dropkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	delete(p.Output, key.String())

	return p, nil
}

func DropOriginData(p *Pipeline, node parser.Node) (*Pipeline, error) {
	delete(p.Output, "message")
	return p, nil
}

func Addkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	err := p.setContent(key, val)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	return p, nil
}

func getGjsonResult(data, id string) interface{} {
	g := gjson.Get(data, id)
	switch g.Type {
	case gjson.Null:
		return nil

	case gjson.False:
		return false

	case gjson.Number:
		if strings.Contains(g.Raw, ".") {
			return g.Float()
		} else {
			return g.Int()
		}

	case gjson.String:
		return g.String()

	case gjson.True:
		return true

	case gjson.JSON:
		return g.Raw

	default:
		return nil
	}
}
