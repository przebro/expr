package expr

import "testing"

func TestErrors(t *testing.T) {

	error_msg := "test_error_msg"
	err := newParserError(error_msg)
	if err.Error() != error_msg {
		t.Error("unexpected result:", err.Error(), "expected:", error_msg)
	}
	err = newLexerError(error_msg)
	if err.Error() != error_msg {
		t.Error("unexpected result:", err.Error(), "expected:", error_msg)
	}

	err = newEvaluateError(error_msg)
	if err.Error() != error_msg {
		t.Error("unexpected result:", err.Error(), "expected:", error_msg)
	}
}
