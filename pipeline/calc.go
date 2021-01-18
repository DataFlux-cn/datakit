package pipeline

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func Calc(expr interface{}, p *Pipeline) (interface{}, error) {
	if expr == nil {
		return nil, nil
	}

	switch v := expr.(type) {
	case *parser.ParenExpr:
		return Calc(v.Param, p)

	case *parser.BinaryExpr:
		lv, err := Calc(v.LHS, p)
		if err != nil {
			return nil, err
		}

		rv, err := Calc(v.RHS, p)
		if err != nil {
			return nil, err
		}

		return binaryOp(lv, rv, int(v.Op))

	case *parser.NumberLiteral:
		if v.IsInt {
			return v.Int, nil
		} else {
			return v.Float, nil
		}

	case *parser.Identifier:
		return p.getContent(v)

	case *parser.AttrExpr:
		return p.getContent(v)

	default:
		return nil, fmt.Errorf("unsupported Expr %v", v)
	}
}

func binaryOp(lv, rv interface{}, opCode int) (val interface{}, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	switch opCode {
	//四则运算
	case parser.ADD:
		val, err = binaryAdd(lv, rv)
	case parser.SUB:
		val, err = binarySub(lv, rv)
	case parser.MUL:
		val, err = binaryMul(lv, rv)
	case parser.DIV:
		val, err = binaryDiv(lv, rv)
	case parser.MOD:
		val, err = binaryMod(lv, rv)

	//关系运算
	case parser.GTE:
		val, err = binaryGte(lv, rv)
	case parser.GT:
		val, err = binaryGt(lv, rv)
	case parser.LTE:
		val, err = binaryLte(lv, rv)
	case parser.LT:
		val, err = binaryLt(lv, rv)
	case parser.EQ:
		val, err = binaryEq(lv, rv)
	case parser.NEQ:
		val, err = binaryNeq(lv, rv)

		//逻辑运算
	case parser.AND:
		val, err = binaryAnd(lv, rv)
	case parser.OR:
		val, err = binaryOr(lv, rv)

	default:
		err = fmt.Errorf("opcode %v unsupported", opCode)
	}
	return
}

func binaryAdd(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T + %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 + v2, nil
		case uint64:
			return v1 + int64(v2), nil
		case float64:
			return float64(v1) + v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) + v2, nil
		case uint64:
			return v1 + v2, nil
		case float64:
			return float64(v1) + v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 + float64(v2), nil
		case uint64:
			return v1 + float64(v2), nil
		case float64:
			return v1 + v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binarySub(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T - %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 - v2, nil
		case uint64:
			return v1 - int64(v2), nil
		case float64:
			return float64(v1) - v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) - v2, nil
		case uint64:
			return v1 - v2, nil
		case float64:
			return float64(v1) - v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 - float64(v2), nil
		case uint64:
			return v1 - float64(v2), nil
		case float64:
			return v1 - v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryMul(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T * %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 * v2, nil
		case uint64:
			return v1 * int64(v2), nil
		case float64:
			return float64(v1) * v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) * v2, nil
		case uint64:
			return v1 * v2, nil
		case float64:
			return float64(v1) * v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 * float64(v2), nil
		case uint64:
			return v1 * float64(v2), nil
		case float64:
			return v1 * v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryDiv(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T / %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 / v2, nil
		case uint64:
			return v1 / int64(v2), nil
		case float64:
			return float64(v1) / v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) / v2, nil
		case uint64:
			return v1 / v2, nil
		case float64:
			return float64(v1) / v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 / float64(v2), nil
		case uint64:
			return v1 / float64(v2), nil
		case float64:
			return v1 / v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryMod(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T %% %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 % v2, nil
		case uint64:
			return v1 % int64(v2), nil
		case float64:
			return v1 % int64(v2), nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) % v2, nil
		case uint64:
			return v1 % v2, nil
		case float64:
			return int64(v1) % int64(v2), nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return int64(v1) % v2, nil
		case uint64:
			return int64(v1) % int64(v2), nil
		case float64:
			return int64(v1) % int64(v2), nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryGte(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T >= %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 >= v2, nil
		case uint64:
			return uint64(v1) >= v2, nil
		case float64:
			return float64(v1) >= v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 >= uint64(v2), nil
		case uint64:
			return v1 >= v2, nil
		case float64:
			return float64(v1) >= v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 >= float64(v2), nil
		case uint64:
			return v1 >= float64(v2), nil
		case float64:
			return v1 >= v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryGt(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T > %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 > v2, nil
		case uint64:
			return uint64(v1) > v2, nil
		case float64:
			return float64(v1) > v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 > uint64(v2), nil
		case uint64:
			return v1 > v2, nil
		case float64:
			return float64(v1) > v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 > float64(v2), nil
		case uint64:
			return v1 > float64(v2), nil
		case float64:
			return v1 > v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryLte(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T <= %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 <= v2, nil
		case uint64:
			return uint64(v1) <= v2, nil
		case float64:
			return float64(v1) <= v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 <= uint64(v2), nil
		case uint64:
			return v1 <= v2, nil
		case float64:
			return float64(v1) <= v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 <= float64(v2), nil
		case uint64:
			return v1 <= float64(v2), nil
		case float64:
			return v1 <= v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryLt(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T < %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 < v2, nil
		case uint64:
			return uint64(v1) < v2, nil
		case float64:
			return float64(v1) < v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 < uint64(v2), nil
		case uint64:
			return v1 < v2, nil
		case float64:
			return float64(v1) < v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 < float64(v2), nil
		case uint64:
			return v1 < float64(v2), nil
		case float64:
			return v1 < v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryEq(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T == %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 == v2, nil
		case uint64:
			return uint64(v1) == v2, nil
		case float64:
			return float64(v1) == v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 == uint64(v2), nil
		case uint64:
			return v1 == v2, nil
		case float64:
			return float64(v1) == v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 == float64(v2), nil
		case uint64:
			return v1 == float64(v2), nil
		case float64:
			return v1 == v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryNeq(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T != %T", lv, rv)
	switch v1 := lv.(type) {
	case int64:
		switch v2 := rv.(type) {
		case int64:
			return v1 != v2, nil
		case uint64:
			return uint64(v1) != v2, nil
		case float64:
			return float64(v1) != v2, nil
		default:
			return nil, err
		}

	case uint64:
		switch v2 := rv.(type) {
		case int64:
			return v1 != uint64(v2), nil
		case uint64:
			return v1 != v2, nil
		case float64:
			return float64(v1) != v2, nil
		default:
			return nil, err
		}

	case float64:
		switch v2 := rv.(type) {
		case int64:
			return v1 != float64(v2), nil
		case uint64:
			return v1 != float64(v2), nil
		case float64:
			return v1 != v2, nil
		default:
			return nil, err
		}

	default:
		return nil, err
	}
}

func binaryAnd(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T &&  %T", lv, rv)
	switch v1 := lv.(type) {
	case bool:
		switch v2 := rv.(type) {
		case bool:
			return v1 && v2, nil
		default:
			return nil, err
		}
	default:
		return nil, err
	}
}

func binaryOr(lv, rv interface{}) (interface{}, error) {
	err := fmt.Errorf("unsupported %T || %T", lv, rv)
	switch v1 := lv.(type) {
	case bool:
		switch v2 := rv.(type) {
		case bool:
			return v1 || v2, nil
		default:
			return nil, err
		}
	default:
		return nil, err
	}
}
