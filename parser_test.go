package expr

import (
	"fmt"
	"testing"
)

type testCaseExpect struct {
	testCase      string
	expectedValue bool
	expectedError error
}

func TestEval(t *testing.T) {

	values := map[string]interface{}{
		"label_01": true,
		"label_02": true,
		"label_03": false,
	}

	input := []string{
		//		"()",
		// "",
		"((!(label_01)))",
		"(((!label_01)))",
		"!(!(!(!label_01)))",
		"label_01",
		"(label_01 && label_03) || !label_02",
		"!(label_01)",
		"(!label_01)",
		"!label_01",
		"!label_01 || !label_03",
		"label_01 == true",
		"(label_02 || label_01)",
		"label_01 && label_03 || !label_02",
		"label_02 && !(label_01 && label_03)",
	}

	for i, in := range input {
		r, err := Eval(in, values)
		if err != nil {
			t.Error("unexpected result input:", i, "error:", err)
		}
		fmt.Println(r)
	}

}

func TestGetIntValue(t *testing.T) {

	var actual valueNode = &intValueExpr{val: 345}
	var result int
	if err := actual.value(&result); err != nil {
		t.Error("unexpected result")
	}

}

func TestEvaluateBool_Positive(t *testing.T) {
	input := []testCaseExpect{
		{"true == label_03", false, nil},
		{"label_03 == label_01", false, nil},
		{"label_03 == false", true, nil},
		{"label_03 != false", false, nil},
		{"true != false", true, nil},
		{"true != true", false, nil},
		{"true == true", true, nil},
		{"true || false", true, nil},
		{"true && false", false, nil},
		{"label_02 != true", false, nil},
		{"label_01 == true", true, nil},
		{"label_01 || false", true, nil},
		{"label_01 || label_02", true, nil},
		{"label_01 || label_03", true, nil},
		{"label_01 || (label_03 || label_02)", true, nil},
		{"label_02 && false", false, nil},
		{"label_02 && true", true, nil},
		{"label_02 && label_01", true, nil},
		{"!label_01", false, nil},
		{"!label_03", true, nil},
		{"!label_03 == true", true, nil},
		{"!label_03 != true", false, nil},
	}

	values := map[string]interface{}{
		"label_01": true,
		"label_02": true,
		"label_03": false,
	}

	for i, in := range input {
		r, err := Eval(in.testCase, values)
		if err != nil {
			t.Error("unexpected result input:", i, "error:", err)
		}
		if r != in.expectedValue {
			t.Error("unexpected result:", i, "value:", r, "expected:", in.expectedValue)
		}
	}
}

func TestEvaluateNumbers_Positive(t *testing.T) {
	input := []testCaseExpect{
		{"label_01 != label_04", true, nil},
		{"label_01 == label_04", false, nil},
		{"label_01 == 15", true, nil},
		{"label_01 != 15", false, nil},
		{"13 == 13", true, nil},
		{"13 != 13", false, nil},
		{"15 != 13", true, nil},
		{"13 == 15", false, nil},
		{"13 == 13", true, nil},
		{"!(13 != 13)", false, nil},
	}

	values := map[string]interface{}{
		"label_01": 15,
		"label_02": true,
		"label_03": false,
		"label_04": 7,
	}

	for i, in := range input {
		r, err := Eval(in.testCase, values)
		if err != nil {
			t.Error("unexpected result input:", i, "error:", err)
		}
		if r != in.expectedValue {
			t.Error("unexpected result:", i, "value:", r, "expected:", in.expectedValue)
		}
	}
}

func TestEvaluateNumbers_Negative(t *testing.T) {
	input := []testCaseExpect{
		{"label_01 == label_02", false, EvaluateError{}},
		{"(label_01 == 15) == 15", false, EvaluateError{}},
		{"(label_01 != 15) == 15", false, EvaluateError{}},
		{"13 && 15", false, EvaluateError{}},
		{"13 || 15", false, EvaluateError{}},
	}

	values := map[string]interface{}{
		"label_01": 15,
		"label_02": true,
		"label_03": false,
		"label_04": 7,
	}

	for i, in := range input {
		_, err := Eval(in.testCase, values)
		if err == nil {
			t.Error("unexpected result:", i, "expected error:", in.expectedError)
		} else {
			fmt.Println(err.Error())
		}
	}
}

func TestEvaluateString_Positive(t *testing.T) {
	input := []testCaseExpect{
		{"label_01 == ''", true, nil},
		{"label_02 == 'Test String'", true, nil},
		{"label_02 != 'test string'", true, nil},
		{"label_03 != label_02", true, nil},
		{"label_03 == label_02", false, nil},
	}

	values := map[string]interface{}{
		"label_01": "",
		"label_02": "Test String",
		"label_03": "test string",
	}

	for i, in := range input {
		r, err := Eval(in.testCase, values)
		if err != nil {
			t.Error("unexpected result input:", i, "error:", err)
		}
		if r != in.expectedValue {
			t.Error("unexpected result:", i, "value:", r, "expected:", in.expectedValue)
		}
	}
}

func TestEvaluateString_Negative(t *testing.T) {
	input := []testCaseExpect{
		{"label_01 || label_02", false, EvaluateError{}},
		{"(label_01 && 15) == 15", false, EvaluateError{}},
		{"'test string' && label_03", false, EvaluateError{}},
		{"!label_03", false, EvaluateError{}},
	}

	values := map[string]interface{}{
		"label_01": "",
		"label_02": "Test String",
		"label_03": "test string",
	}

	for i, in := range input {
		_, err := Eval(in.testCase, values)
		if err == nil {
			t.Error("unexpected result:", i, "expected error:", in.expectedError)
		} else {
			fmt.Println(err.Error())
		}
	}
}
