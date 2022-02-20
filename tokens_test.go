package expr

import (
	"fmt"
	"testing"
)

func TestLexer_Tokenize(t *testing.T) {
	in := []string{
		"''",
		"'string value'",
		"label_01",
		"1234",
		"label_01 || label_02",
		"()",
		"!",
		"!=",
		"||",
		"&&",
		"==",
		"!()",
		"!label",
		"label_01 || label_02 && (label_03 != false && !label_04)",
	}

	for n, input := range in {

		_, err := Tokenize(input)
		if err != nil {
			t.Error("unexepected result, in:", n)
		}
	}

}
func TestLexer_Tokenize_Data(t *testing.T) {

	in := []string{
		"label_05  == 'string value' && some_int == 4321",
		"label_01 || label_02 && (label_03 != false && !label_04) && label_05 == 'string value'",
		"label_01 || label_02 &&\n (label_03 != false &&\n !label_04\n)",
	}

	for n, input := range in {

		result, err := Tokenize(input)
		if err != nil {
			t.Error("unexepected result, in:", n)
		}

		for _, token := range result {
			fmt.Println("type:", token.tokenType, "value", token.value, "pos:", token.pos, "len:", token.length, "line:", token.line)
		}
	}

}
func TestLexer_Tokenize_errors(t *testing.T) {

	in := []string{
		"'''",
		"&=",
		"",
		"$label",
		"label$",
		"lab#el",
		"! label",
		"%label",
		"label % label_01",
	}

	for n, input := range in {

		_, err := Tokenize(input)
		if err == nil {
			t.Error("unexepected result, in:", n)
		}
	}
}
