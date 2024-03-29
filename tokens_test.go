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

		_, err := tokenize(input)
		if err != nil {
			t.Error("unexepected result, in:", n)
		}
	}

}
func TestLexer_TT(t *testing.T) {
	input := "label_01.PREV == true || !label_02 && !(label_03.NEXT || label_04.DATE)"

	result, vars, err := Translate(input)
	if err != nil {
		t.Error("unexepected result", err)
	}
	if result == "" {
		t.Error("unexepected result, in", result)
	}
	if len(vars) != 4 {
		t.Error("unexepected result, in", vars)
	}

}
func TestLexer_TA(t *testing.T) {
	input := "label-01.PREV||  LABEL-02"

	result, vars, err := Translate(input)
	if err != nil {
		t.Error("unexepected result", err)
	}
	if result == "" {
		t.Error("unexepected result, in", result)
	}
	if len(vars) != 2 {
		t.Error("unexepected result, in", vars)
	}

}

func TestLexer_TB(t *testing.T) {
	input := "label-01.PREV ||  !(LABEL-02 && LABEL-03.NEXT)"
	z := false
	x := true
	y := true
	fmt.Println(z || !(x && y))
	result, err := Eval(input, map[string]interface{}{"label-01.PREV": false, "LABEL-02": true, "LABEL-03.NEXT": true})

	fmt.Println("result:", result, "err:", err)

}

func TestLexer_Tokenize_Data(t *testing.T) {

	in := []string{
		"label_01.PREV ||  LABEL_02",
		"label_01.PREV == true || !label_02 && !(label_03.NEXT || label_04.DATE)",
		"label_01.PREV || !label_02 && (label_03.NEXT || label_04)",
		"label_05  == 'string value' && some_int == 4321",
		"label_01 || label_02 && (label_03 != false && !label_04) && label_05 == 'string value'",
		"label_01 || label_02 &&\n (label_03 != false &&\n !label_04\n)",
	}

	for n, input := range in {

		result, err := tokenize(input)
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

		_, err := tokenize(input)
		if err == nil {
			t.Error("unexepected result, in:", n)
		}
	}
}
