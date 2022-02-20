package expr

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"
)

type TokenValue string

const (
	Token_AND       TokenValue = "&&"
	Token_OR        TokenValue = "||"
	Token_NOT       TokenValue = "!="
	Token_NEG       TokenValue = "!"
	Token_CMP       TokenValue = "=="
	Token_BRACKET_R TokenValue = ")"
	Token_BRACKET_L TokenValue = "("
	Token_TRUE      TokenValue = "true"
	Token_FALSE     TokenValue = "false"
	Token_EMPTY     TokenValue = ""
)

type TokenType int

const (
	TokenT_END    TokenType = 0
	TokenT_OPER   TokenType = 1
	TokenT_CONS   TokenType = 3
	TokenT_NUMBER TokenType = 4
	TokenT_STRVAL TokenType = 5
	TokenT_IDENT  TokenType = 6
	TokenT_LPAR   TokenType = 7
	TokenT_RPAR   TokenType = 8
	TokenT_LOPER  TokenType = 9
)

type ParserToken struct {
	tokenType TokenType
	value     string
	length    int
	line      int
	pos       int
}

type lexerState struct {
	stream string
	next   rune
	buffer string
	output []ParserToken
	err    error
	line   int32
	pos    int32
}

type lexerFunc func(lexer *lexerState) lexerFunc

func Tokenize(expr string) ([]ParserToken, error) {

	if len(expr) == 0 {
		return nil, errors.New("empty stream")
	}

	var state lexerFunc = lexEmpty
	lexer := &lexerState{
		stream: expr,
		next:   rune(expr[0]),
		buffer: "",
		output: []ParserToken{},
		line:   1,
		pos:    -1,
		err:    nil,
	}

	lexer.move()

	for state != nil {
		state = state(lexer)
	}

	return lexer.output, lexer.err
}

func lexEmpty(state *lexerState) lexerFunc {

	if state.next == ' ' || state.next == '\t' || state.next == '\n' || state.next == '\r' {
		return lexWS(state)
	}
	if (unicode.IsDigit(state.next) && len(state.buffer) != 0) || unicode.IsLetter(state.next) || state.next == '_' {
		return lexIdent(state)
	}

	if unicode.IsDigit(state.next) && len(state.buffer) == 0 {
		return lexNumber(state)
	}

	if state.next == '\'' {
		return lexString(state)
	}

	if state.next == '(' || state.next == ')' {
		state.buffer = state.buffer + string(state.next)
		if t, err := state.classify(); err == nil {
			state.produce(t)
			state.buffer = ""
			state.move()
		} else {
			state.err = newLexerError(fmt.Sprintf("unexpected char:%c,line:%d,position:%d", state.next, state.line, state.pos))
			return nil
		}

		return lexEmpty
	}

	if state.next == '!' || state.next == '|' || state.next == '&' || state.next == '=' {
		return lexOper(state)
	}

	if state.next == 0x00 {
		return nil
	}
	state.err = newLexerError(fmt.Sprintf("unexpected result line:%d,position:%d", state.line, state.pos))
	return nil
}

func (lex *lexerState) produce(tp TokenType) {

	lex.output = append(lex.output, ParserToken{
		tokenType: tp,
		value:     lex.buffer,
		length:    len(lex.buffer),
		pos:       int(lex.pos) - len(lex.buffer),
		line:      int(lex.line),
	})

}
func (lex *lexerState) move() {
	if lex.stream == "" {
		lex.next = 0x00
		return
	}
	lex.next, lex.stream = rune(lex.stream[0]), lex.stream[1:]
	lex.pos++
}
func (lex *lexerState) classify() (TokenType, error) {

	rLiteral := regexp.MustCompile(`^[A-Za-z][\w\d_]*$`)
	nLiteral := regexp.MustCompile(`^\-?\d+$`)
	strLiteral := regexp.MustCompile(`^\'[^\t\n\'\r]*'$`)

	switch true {
	case lex.buffer == string(Token_BRACKET_L):
		{
			return TokenT_LPAR, nil
		}
	case lex.buffer == string(Token_BRACKET_R):
		{
			return TokenT_RPAR, nil
		}
	case lex.buffer == string(Token_CMP):
		fallthrough
	case lex.buffer == string(Token_NOT):
		fallthrough
	case lex.buffer == string(Token_OR):
		fallthrough
	case lex.buffer == string(Token_AND):
		{
			return TokenT_OPER, nil
		}
	case lex.buffer == string(Token_NEG):
		{
			return TokenT_LOPER, nil
		}
	case lex.buffer == "true" || lex.buffer == "false":
		{
			return TokenT_CONS, nil
		}
	case rLiteral.Match([]byte(lex.buffer)):
		{
			return TokenT_IDENT, nil
		}
	case nLiteral.Match([]byte(lex.buffer)):
		{
			return TokenT_NUMBER, nil
		}
	case strLiteral.Match([]byte(lex.buffer)):
		{
			return TokenT_STRVAL, nil
		}
	}

	return 0, newLexerError(fmt.Sprintf("unrecognized token:%s,line:%d,position:%d", lex.buffer, lex.line, lex.pos))
}

func lexWS(state *lexerState) lexerFunc {

	state.buffer = ""

	if state.next == ' ' || state.next == '\t' || state.next == '\r' {
		state.move()
		return lexWS
	} else if state.next == '\n' {
		state.line++
		state.move()
		state.pos = 1
		return lexWS

	} else {
		return lexEmpty
	}

}

func lexIdent(state *lexerState) lexerFunc {

	state.buffer = state.buffer + string(state.next)
	state.move()

	if (unicode.IsDigit(state.next) && len(state.buffer) != 0) || unicode.IsLetter(state.next) || state.next == '_' {
		return lexIdent
	} else {
		if t, err := state.classify(); err == nil {
			state.produce(t)
			state.buffer = ""
			return lexEmpty
		}
		return nil
	}
}

func lexNumber(state *lexerState) lexerFunc {

	state.buffer = state.buffer + string(state.next)
	state.move()

	if unicode.IsDigit(state.next) {
		return lexNumber
	} else {
		if t, err := state.classify(); err == nil {
			state.produce(t)
			state.buffer = ""
			return lexEmpty
		}
		return nil
	}
}

func lexString(state *lexerState) lexerFunc {

	state.buffer = state.buffer + string(state.next)
	state.move()

	if state.next == '\'' {

		state.buffer = state.buffer + string(state.next)
		state.move()

		if state.next == ' ' || state.next == '\t' || state.next == '\r' || state.next == 0x00 {

			if t, err := state.classify(); err == nil {
				state.produce(t)
				state.buffer = ""
				return lexEmpty
			}
			return nil
		} else {
			state.err = newLexerError(fmt.Sprintf("unexpected char:%c,line:%d,position:%d", state.next, state.line, state.pos))
		}
	} else {
		return lexString
	}

	return nil
}

func lexOper(state *lexerState) lexerFunc {

	state.buffer = state.buffer + string(state.next)
	state.move()

	if state.next == '!' || state.next == '|' || state.next == '&' || state.next == '=' {
		return lexOper
	} else {
		if t, err := state.classify(); err == nil {
			state.produce(t)
			state.buffer = ""
			if t == TokenT_LOPER && unicode.IsSpace(state.next) {
				state.err = newLexerError(fmt.Sprintf("unexpected char:%c,line:%d,position:%d", state.next, state.line, state.pos))
				return nil
			}
			return lexEmpty
		} else {
			state.err = newLexerError(fmt.Sprintf("unexpected char:%c,line:%d,position:%d", state.next, state.line, state.pos))
			return nil
		}
	}

}
