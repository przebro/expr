package expr

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"
)

type TokenValue string

const (
	token_AND       TokenValue = "&&"
	token_OR        TokenValue = "||"
	token_NOT       TokenValue = "!="
	token_NEG       TokenValue = "!"
	token_CMP       TokenValue = "=="
	token_BRACKET_R TokenValue = ")"
	token_BRACKET_L TokenValue = "("
	token_TRUE      TokenValue = "true"
	token_FALSE     TokenValue = "false"
	token_EMPTY     TokenValue = ""
)

type TokenType int

const (
	tokenT_END    TokenType = 0
	tokenT_OPER   TokenType = 1
	tokenT_CONS   TokenType = 3
	tokenT_NUMBER TokenType = 4
	tokenT_STRVAL TokenType = 5
	tokenT_IDENT  TokenType = 6
	tokenT_LPAR   TokenType = 7
	tokenT_RPAR   TokenType = 8
	tokenT_LOPER  TokenType = 9
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

func Extract(expr string) ([]string, error) {
	tokens, err := tokenize(expr)
	variables := []string{}

	if err != nil {
		return []string{}, err
	}
	for _, t := range tokens {
		if t.tokenType == tokenT_IDENT {
			variables = append(variables, t.value)
		}
	}

	return variables, nil
}

func Translate(expr string) (string, []string, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return "", []string{}, err
	}

	result := ""
	variables := []string{}

	for n, t := range tokens {
		spc := " "
		if t.tokenType == tokenT_IDENT {
			if t.value == "label_04.DATE" {
				fmt.Println("")
			}
			variables = append(variables, t.value)

			if n < len(tokens) {
				if n-1 >= 0 && tokens[n-1].tokenType == tokenT_LOPER {
					spc = " "
				} else if n+1 < len(tokens) {
					if tokens[n+1].value != "==" {

						result = result + t.value + " == "
						t = ParserToken{tokenT_CONS, "true", 4, t.line, t.pos}
						spc = " "
					}
				} else {
					result = result + t.value + " == "
					t = ParserToken{tokenT_CONS, "true", 4, t.line, t.pos}
					spc = " "
				}

			}
		}
		if tokens[n].tokenType == tokenT_LOPER {
			spc = ""
		}

		result = result + t.value + spc
	}

	return result, variables, nil
}

func tokenize(expr string) ([]ParserToken, error) {

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

	rLiteral := regexp.MustCompile(`^[A-Za-z][\w\d_\.\-]*$`)
	nLiteral := regexp.MustCompile(`^\-?\d+$`)
	strLiteral := regexp.MustCompile(`^\'[^\t\n\'\r]*'$`)

	switch true {
	case lex.buffer == string(token_BRACKET_L):
		{
			return tokenT_LPAR, nil
		}
	case lex.buffer == string(token_BRACKET_R):
		{
			return tokenT_RPAR, nil
		}
	case lex.buffer == string(token_CMP):
		fallthrough
	case lex.buffer == string(token_NOT):
		fallthrough
	case lex.buffer == string(token_OR):
		fallthrough
	case lex.buffer == string(token_AND):
		{
			return tokenT_OPER, nil
		}
	case lex.buffer == string(token_NEG):
		{
			return tokenT_LOPER, nil
		}
	case lex.buffer == "true" || lex.buffer == "false":
		{
			return tokenT_CONS, nil
		}
	case rLiteral.Match([]byte(lex.buffer)):
		{
			return tokenT_IDENT, nil
		}
	case nLiteral.Match([]byte(lex.buffer)):
		{
			return tokenT_NUMBER, nil
		}
	case strLiteral.Match([]byte(lex.buffer)):
		{
			return tokenT_STRVAL, nil
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

	if (unicode.IsDigit(state.next) && len(state.buffer) != 0) || unicode.IsLetter(state.next) || state.next == '_' || state.next == '-' || state.next == '.' {
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
			if t == tokenT_LOPER && unicode.IsSpace(state.next) {
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
