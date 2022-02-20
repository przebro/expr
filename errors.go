package expr

type LexerError struct {
	msg string
}

func (le LexerError) Error() string {
	return le.msg
}

func newLexerError(msg string) error {
	return LexerError{msg: msg}
}

type ParserError struct {
	msg string
}

func (le ParserError) Error() string {
	return le.msg
}

func newParserError(msg string) error {
	return ParserError{msg: msg}
}

type EvaluateError struct {
	msg string
}

func (ee EvaluateError) Error() string {
	return ee.msg
}

func newEvaluateError(msg string) error {
	return EvaluateError{msg: msg}
}
