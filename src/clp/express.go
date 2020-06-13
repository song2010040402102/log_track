package clp

import (
	"errors"
	"fmt"
	"util"
)

const (
	S_MUL = "*"
	S_DIV = "/"
	S_ADD = "+"
	S_SUB = "-"
	S_CMA = ","
)

const (
	OP_NONE int = iota
	OP_MUL
	OP_DIV
	OP_ADD
	OP_SUB
	OP_CMA
	OP_FUNC = 100
)

var g_strOpera []string = []string{S_MUL, S_DIV, S_ADD, S_SUB, S_CMA}

var g_mapOpera map[string]int = map[string]int{
	S_MUL: OP_MUL,
	S_DIV: OP_DIV,
	S_ADD: OP_ADD,
	S_SUB: OP_SUB,
	S_CMA: OP_CMA,
}

var g_opPrior []int = []int{0, 1, 1, 2, 2, 3}

type Express struct {
	exp    string
	values []float64
	opera  int
	childs []*Express
}

func NewExpress() *Express {
	e := &Express{
		opera: OP_NONE,
	}
	return e
}

func (e *Express) GetValues() []float64 {
	return e.values
}

func (e *Express) GetOpera() int {
	return e.opera
}

func (e *Express) IsOpera() bool {
	return e.opera != OP_NONE && len(e.childs) == 0
}

const (
	S_AVER = "aver"
)

type BuiltinFunc func(...[]float64) []float64

type Function struct {
	name string
	pnum int
	sysF BuiltinFunc
}

var g_funcs []*Function = []*Function{
	&Function{S_AVER, -1, SysAver},
}

func findFuncIndex(name string) int {
	for i, v := range g_funcs {
		if v.name == name && v.sysF != nil {
			return i
		}
	}
	return -1
}

func SysAver(values ...[]float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	ret := make([]float64, len(values[0]))
	for i := 0; i < len(ret); i++ {
		for j := 0; j < len(values); j++ {
			ret[i] += values[j][i]
		}
		ret[i] /= float64(len(values))
	}
	return ret
}

func ParseExpress(exp string) (*Express, error) {
	return parseExp(util.RemoveBlank(exp))
}

func RunExpress(exp *Express, handler func(string) ([]float64, error)) error {
	if exp.opera == OP_NONE {
		return errors.New("Operation cannot OP_NONE!")
	}
	return runExp(exp, handler)
}

func runExp(exp *Express, handler func(string) ([]float64, error)) error {
	var vals [][]float64
	for _, child := range exp.childs {
		if child.opera != OP_NONE {
			err := runExp(child, handler)
			if err != nil {
				return err
			}
		} else {
			var err error
			child.values, err = handler(child.exp)
			if err != nil {
				return err
			}
		}
		vals = append(vals, child.values)
	}
	if exp.opera >= OP_FUNC {
		exp.values = g_funcs[exp.opera-OP_FUNC].sysF(vals...)
	} else {
		if len(vals) != 2 {
			return errors.New("Operand num should be 2!")
		}
		v1s, v2s := vals[0], vals[1]
		exp.values = make([]float64, len(v1s))
		for i := 0; i < len(v1s); i++ {
			switch exp.opera {
			case OP_MUL:
				exp.values[i] = v1s[i] * v2s[i]
			case OP_DIV:
				if v2s[i] == 0 {
					exp.values[i] = 0
				} else {
					exp.values[i] = v1s[i] / v2s[i]
				}
			case OP_ADD:
				exp.values[i] = v1s[i] + v2s[i]
			case OP_SUB:
				exp.values[i] = v1s[i] - v2s[i]
			}
		}
	}
	return nil
}

func lexAnalysis(exp string) ([]*Express, error) {
	if len(exp) == 0 {
		return nil, errors.New("Express cannot empty!")
	}
	last := 0
	exps := []*Express{}
	for i := 0; i <= len(exp); {
		s, e, err := handleMatch(exp, i)
		if err != nil {
			return exps, err
		} else {
			i = e
		}
		op := ""
		for _, v := range g_strOpera {
			if i <= len(exp)-len(v) && exp[i:i+len(v)] == v {
				op = v
				break
			}
		}
		if op == "" && i < len(exp) {
			i++
		} else {
			var ep *Express
			if s < len(exp) && exp[s] == '(' {
				if s == last {
					ep, err = parseExp(exp[s+1 : e-1])
				} else {
					ep, err = parseFunc(exp[last:s], exp[s+1:e-1])
				}
				if ep == nil {
					return exps, err
				}
			} else if i > last {
				ep = NewExpress()
				ep.exp = exp[last:i]
			}
			if ep != nil {
				exps = append(exps, ep)
			}
			if op != "" {
				eop := NewExpress()
				eop.opera = g_mapOpera[op]
				exps = append(exps, eop)
			}
			i += len(op)
			last = i
			if last == len(exp) {
				break
			}
		}
	}
	return exps, nil
}

func syntaxAnalysis(exps []*Express) (*Express, error) {
	for {
		min := -1
		for i := 0; i < len(exps); i++ {
			if exps[i].IsOpera() {
				if min == -1 || g_opPrior[exps[i].opera] < g_opPrior[exps[min].opera] {
					min = i
				}
			}
		}
		var new *Express
		if min != -1 {
			new = NewExpress()
			new.opera = exps[min].opera
			if min < 1 || min >= len(exps)-1 {
				return nil, errors.New("Syntax error!")
			}
			new.childs = append(new.childs, exps[min-1])
			new.childs = append(new.childs, exps[min+1])
			exps = updateExps(exps, new, min-1, min+1)
		} else {
			new = exps[0]
		}
		if len(exps) < 2 {
			return new, nil
		}
	}
	return nil, errors.New("Unknown error!")
}

func parseExp(exp string) (*Express, error) {
	exps, err := lexAnalysis(exp)
	if err != nil {
		return nil, err
	}
	return syntaxAnalysis(exps)
}

func parseFunc(name string, paras string) (*Express, error) {
	index := findFuncIndex(name)
	if index == -1 {
		return nil, errors.New(fmt.Sprintf("%s undefined!", name))
	}
	ret := &Express{
		opera: index + OP_FUNC,
	}
	exp, err := parseExp(paras)
	if exp == nil {
		return exp, err
	}
	ret.childs = spliteCommaExp(exp)
	if g_funcs[index].pnum != -1 && len(ret.childs) != g_funcs[index].pnum {
		return nil, errors.New("parameter num diff!")
	}
	return ret, nil
}

func spliteCommaExp(exp *Express) []*Express {
	exps := []*Express{}
	p := exp
	for {
		if p.opera == OP_CMA {
			exps = append(exps, p.childs[1])
		} else {
			exps = append(exps, p)
			break
		}
		p = p.childs[0]
	}
	ret := make([]*Express, 0, len(exps))
	for i := len(exps) - 1; i >= 0; i-- {
		ret = append(ret, exps[i])
	}
	return ret
}

func updateExps(exps []*Express, new *Express, start, end int) []*Express {
	ret := make([]*Express, 0, len(exps)-end+start)
	if start <= 0 && end >= len(exps)-1 {
		ret = append(ret, new)
	} else if start <= 0 {
		ret = append([]*Express{new}, exps[end+1:]...)
	} else if end >= len(exps)-1 {
		ret = append(exps[:start], new)
	} else {
		ret = append(exps[:start], append([]*Express{new}, exps[end+1:]...)...)
	}
	return ret
}

func handleMatch(exp string, cur int) (int, int, error) {
	if cur < len(exp) {
		var c byte = exp[cur]
		var c2 byte
		if c == '\'' || c == '"' {
			c2 = c
		} else if c == '(' {
			c2 = ')'
		} else {
			return cur, cur, nil
		}
		n := 1
		for i := cur + 1; i < len(exp); i++ {
			if c == c2 {
				if exp[i] == c {
					n--
				}
			} else {
				if exp[i] == c {
					n++
				} else if exp[i] == c2 {
					n--
				}
			}
			if n == 0 {
				return cur, i + 1, nil
			}
		}
		return 0, 0, errors.New(fmt.Sprintf("%c not match!", c))
	}
	return cur, cur, nil
}
