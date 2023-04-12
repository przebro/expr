package expr

import (
	"fmt"
	"strconv"
	"strings"
)

type valueT uint8

const (
	boolValue   valueT = 0
	intValue    valueT = 1
	stringValue valueT = 2
)

type exprNode interface {
	evaluate() (bool, error)
	isValue() valueT
}

type valueNode interface {
	value(out interface{}) error
}

type boolValueExpr struct {
	val bool
}

func (ex *boolValueExpr) evaluate() (bool, error) {
	return ex.val, nil
}
func (ex *boolValueExpr) isValue() valueT { return boolValue }

type stringValueExpr struct {
	val string
}

func (ex *stringValueExpr) evaluate() (bool, error) {
	return false, newEvaluateError("can't evaluate string")
}
func (ex *stringValueExpr) isValue() valueT { return stringValue }
func (ex *stringValueExpr) value(out interface{}) error {

	if v, ok := out.(*string); ok {
		*v = ex.val
		return nil
	}

	return newParserError("can't cast value to string")
}

type intValueExpr struct {
	val int
}

func (ex *intValueExpr) evaluate() (bool, error) {
	return false, newEvaluateError("can't evaluate")
}
func (ex *intValueExpr) isValue() valueT { return intValue }
func (ex *intValueExpr) value(out interface{}) error {

	if v, ok := out.(*int); ok {
		*v = ex.val
		return nil
	}

	return newParserError("can't cast value to int")
}

type negValueExpr struct {
	expR exprNode
}

func (ex *negValueExpr) evaluate() (bool, error) {

	if ex.expR.isValue() == boolValue {
		right, _ := ex.expR.evaluate()
		return !right, nil
	}

	return false, newEvaluateError("can't evaluate expression")

}
func (ex *negValueExpr) isValue() valueT { return boolValue }

type compareOperExpr struct {
	exprL exprNode
	exprR exprNode
}

func (ex *compareOperExpr) evaluate() (bool, error) {

	var left, right bool

	if ex.exprL.isValue() == ex.exprR.isValue() && ex.exprL.isValue() == boolValue {
		left, _ = ex.exprL.evaluate()
		right, _ = ex.exprR.evaluate()

		return left == right, nil
	} else if ex.exprL.isValue() == ex.exprR.isValue() {
		if ex.exprL.isValue() == intValue {
			lvalue := 0
			rvalue := 0
			ex.exprL.(valueNode).value(&lvalue)
			ex.exprR.(valueNode).value(&rvalue)

			return lvalue == rvalue, nil
		}
		if ex.exprL.isValue() == stringValue {

			lvalue := ""
			rvalue := ""
			ex.exprL.(valueNode).value(&lvalue)
			ex.exprR.(valueNode).value(&rvalue)

			return lvalue == rvalue, nil

		}
	}

	return false, newEvaluateError("can't evaluate left == right")

}
func (ex *compareOperExpr) isValue() valueT { return boolValue }

type orOperExpr struct {
	exprL exprNode
	exprR exprNode
}

func (ex *orOperExpr) evaluate() (bool, error) {

	var left, right bool

	if ex.exprL.isValue() == ex.exprR.isValue() && ex.exprL.isValue() == boolValue {
		left, _ = ex.exprL.evaluate()
		right, _ = ex.exprR.evaluate()

		return left || right, nil
	}

	return false, newEvaluateError("can't evaluate expression left || right")
}
func (ex *orOperExpr) isValue() valueT { return boolValue }

type andOperExpr struct {
	exprL exprNode
	exprR exprNode
}

func (ex *andOperExpr) evaluate() (bool, error) {

	var left, right bool

	if ex.exprL.isValue() == ex.exprR.isValue() && ex.exprL.isValue() == boolValue {
		left, _ = ex.exprL.evaluate()
		right, _ = ex.exprR.evaluate()

		return left && right, nil
	}

	return false, newEvaluateError("can't evaluate expression")

}
func (ex *andOperExpr) isValue() valueT { return boolValue }

type notOperExpr struct {
	exprL exprNode
	exprR exprNode
}

func (ex *notOperExpr) evaluate() (bool, error) {

	var left, right bool

	if ex.exprL.isValue() == ex.exprR.isValue() && ex.exprL.isValue() == boolValue {
		left, _ = ex.exprL.evaluate()
		right, _ = ex.exprR.evaluate()

		return left != right, nil
	} else if ex.exprL.isValue() == ex.exprR.isValue() {
		if ex.exprL.isValue() == intValue {
			lvalue := 0
			rvalue := 0
			ex.exprL.(valueNode).value(&lvalue)
			ex.exprR.(valueNode).value(&rvalue)

			return lvalue != rvalue, nil
		}
		if ex.exprL.isValue() == stringValue {

			lvalue := ""
			rvalue := ""
			ex.exprL.(valueNode).value(&lvalue)
			ex.exprR.(valueNode).value(&rvalue)

			return lvalue != rvalue, nil
		}
	}

	return false, newEvaluateError("can't evaluate expression")
}
func (ex *notOperExpr) isValue() valueT { return boolValue }

type parser struct {
	tstream   []ParserToken
	current   exprNode
	variables map[string]interface{}
}

func (p *parser) peek() ParserToken {

	if len(p.tstream) == 0 {
		return ParserToken{tokenType: tokenT_END}
	}
	return p.tstream[0]
}

func (p *parser) pop() ParserToken {

	if len(p.tstream) == 0 {
		return ParserToken{tokenType: tokenT_END}
	}
	var value ParserToken
	value, p.tstream = p.tstream[0], p.tstream[1:]

	return value
}

func parse(tstream []ParserToken, variables map[string]interface{}) (exprNode, error) {

	p := parser{tstream: tstream, current: nil, variables: variables}

	var err error

	t := p.peek()
	fn := parseExprN

	for t.tokenType != tokenT_END {
		fn, err = fn(&p)
		if err != nil {
			return nil, err
		}
		t = p.peek()

	}

	return p.current, nil
}

type parserFunc func(p *parser) (parserFunc, error)

func parseExprN(p *parser) (parserFunc, error) {

	next := p.peek()

	if next.tokenType == tokenT_IDENT {

		token := p.pop()

		if v, ok := p.variables[token.value]; ok {
			p.current = createValueExprNode(v) // ::TODO
		} else {
			return nil, newParserError(fmt.Sprintf("undefined variable:%s", token.value))
		}

		return parseExprExpr, nil
	}

	if next.tokenType == tokenT_CONS {

		token := p.pop()
		p.current = &boolValueExpr{val: token.value == "true"}

		return parseExprExpr, nil
	}

	if next.tokenType == tokenT_NUMBER {
		token := p.pop()
		val, err := strconv.Atoi(token.value)
		if err != nil {
			return nil, newLexerError("unexpected value, expected int")
		}
		p.current = &intValueExpr{val: val}
		return parseExprExpr, nil
	}

	if next.tokenType == tokenT_STRVAL {
		token := p.pop()

		p.current = createValueExprNode(token.value)
		return parseExprExpr, nil

	}

	if next.tokenType == tokenT_LOPER {

		expr, err := branchLOperatorExpr(p)
		if err != nil {
			return nil, err
		}
		p.current = expr

		return parseExprExpr, nil
	}

	if next.tokenType == tokenT_LPAR {

		expr, err := branchExpr(p)
		if err != nil {
			return nil, err
		}
		p.current = expr

		return parseExprExpr, nil
	}

	return nil, newLexerError("unexpected token")
}

func parseExprExpr(p *parser) (parserFunc, error) {

	next := p.peek()

	if next.tokenType == tokenT_OPER {
		return parseOperatorExpr, nil
	}

	return nil, newParserError(fmt.Sprintf("unexpected token:%s,line:%d,pos:%d", next.value, next.line, next.pos))
}

func branchLOperatorExpr(p *parser) (exprNode, error) {
	p.pop()
	next := p.peek()

	if next.tokenType == tokenT_IDENT {

		p.pop()

		if v, ok := p.variables[next.value]; ok {
			return &negValueExpr{expR: createValueExprNode(v)}, nil //::TODO
		}
		return nil, newParserError(fmt.Sprintf("undefined variable:%s", next.value))
	}

	if next.tokenType == tokenT_LPAR {
		expr, err := branchExpr(p)
		return &negValueExpr{expR: expr}, err
	}

	return nil, newParserError(fmt.Sprintf("unexpected token:%s,line:%d,pos:%d", next.value, next.line, next.pos))
}

func branchExpr(p *parser) (exprNode, error) {

	p.pop()
	depth := 1

	tsream := []ParserToken{}

	for depth != 0 {
		token := p.pop()
		if token.tokenType == tokenT_LPAR {
			depth++
		}
		if token.tokenType == tokenT_RPAR {
			depth--
		}

		if depth != 0 {
			tsream = append(tsream, token)
		}

	}

	return parse(tsream, p.variables)
}

func parseOperatorExpr(p *parser) (parserFunc, error) {

	token := p.pop()
	next := p.peek()

	if next.tokenType == tokenT_IDENT {
		var right exprNode

		if v, ok := p.variables[next.value]; ok {
			right = createValueExprNode(v) // ::TODO
		} else {
			return nil, newParserError(fmt.Sprintf("undefined variable:%s", next.value))
		}

		p.current = produce(p.current, right, token)

		p.pop()
		return parseExprExpr, nil
	}

	if next.tokenType == tokenT_CONS {
		right := &boolValueExpr{val: next.value == "true"}

		p.current = produce(p.current, right, token)

		p.pop()
		return parseExprExpr, nil

	}

	if next.tokenType == tokenT_STRVAL {

		right := createValueExprNode(next.value)
		p.current = produce(p.current, right, token)

		p.pop()
		return parseExprExpr, nil

	}

	if next.tokenType == tokenT_NUMBER {

		val, err := strconv.Atoi(next.value)
		if err != nil {
			return nil, newLexerError("unexpected value, expected int")
		}

		right := &intValueExpr{val: val}

		p.current = produce(p.current, right, token)

		p.pop()
		return parseExprExpr, nil

	}

	if next.tokenType == tokenT_LOPER {

		if right, err := branchLOperatorExpr(p); err == nil {
			p.current = produce(p.current, right, token)

			return parseExprExpr, nil
		} else {
			return nil, newParserError(fmt.Sprintf("unexpected token:%s,line:%d,pos:%d", next.value, next.line, next.pos))
		}
	}

	if next.tokenType == tokenT_LPAR {

		if right, err := branchExpr(p); err == nil {
			p.current = produce(p.current, right, token)
			return parseExprExpr, nil
		} else {
			return nil, newParserError(fmt.Sprintf("unexpected token:%s,line:%d,pos:%d", next.value, next.line, next.pos))
		}
	}

	return nil, newParserError(fmt.Sprintf("unexpected token:%s,line:%d,pos:%d", next.value, next.line, next.pos))
}

func produce(left, right exprNode, token ParserToken) exprNode {

	var current exprNode
	if token.value == string(token_AND) {
		current = &andOperExpr{exprL: left, exprR: right}
	}
	if token.value == string(token_OR) {
		current = &orOperExpr{exprL: left, exprR: right}
	}
	if token.value == string(token_CMP) {
		current = &compareOperExpr{exprL: left, exprR: right}
	}
	if token.value == string(token_NOT) {
		current = &notOperExpr{exprL: left, exprR: right}
	}
	return current
}

func createValueExprNode(val interface{}) exprNode {

	var node exprNode
	switch x := val.(type) {
	case bool:
		{
			node = &boolValueExpr{val: x}
		}
	case int:
		{
			node = &intValueExpr{val: x}
		}
	case string:
		{
			node = &stringValueExpr{val: strings.Trim(x, "'")}
		}
	}

	return node
}

func Eval(input string, variables map[string]interface{}) (bool, error) {

	var err error
	var tstream []ParserToken
	if tstream, err = tokenize(input); err != nil {
		return false, err
	}
	var ex exprNode
	ex, err = parse(tstream, variables)
	if err != nil {
		return false, err
	}

	return ex.evaluate()
}

func Test(input string) error {
	return nil
}
