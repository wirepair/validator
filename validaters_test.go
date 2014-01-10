package validator

import (
	//"reflect"
	"fmt"
	"regexp"
	"testing"
)

func TestGetArguments(t *testing.T) {
	val1, val2, err := getArguments("len(1:2)", "len")
	if err != nil {
		t.Fatalf("error good values failed to parse: %v", err)
	}
	if val1 != "1" && val2 != "2" {
		t.Fatalf("error values are not correct %s %s", val1, val2)
	}

	val1, val2, err = getArguments("len(1:2:3)", "len")
	if err == nil {
		t.Fatalf("error invalid number of arguments passed.")
	}
}

func testOkFn(val string) error {
	return nil
}

func testBadFn(val string) error {
	return fmt.Errorf("bad function failed validation")
}

func TestAdd(t *testing.T) {
	badFn := testBadFn
	okFn := testOkFn

	Add("badFn", badFn)
	Add("okFn", okFn)

	if err := userFns.Funcs["badFn"]("whatever"); err == nil {
		t.Fatalf("error didn't occur.")
	}

	if err := userFns.Funcs["okFn"]("whatever"); err != nil {
		t.Fatalf("error occurred: %v", err)
	}
}

func TestIntFuncArguments(t *testing.T) {
	nmin, nmax, err := intFuncArguments("-1", "2", "range")
	if err != nil {
		t.Fatalf("Error returned from intFuncArguments with good values: %v", err)
	}
	if nmin != -1 || nmax != 2 {
		t.Fatalf("Error invalid min/max returned")
	}

	_, _, err = intFuncArguments("a", "2", "range")
	if err == nil {
		t.Fatalf("intFuncArguments passed with bad values")
	}

	_, _, err = intFuncArguments("1", "a", "range")
	if err == nil {
		t.Fatalf("intFuncArguments passed with bad values")
	}

	_, _, err = intFuncArguments("4", "1", "range")
	if err == nil {
		t.Fatalf("intFuncArguments passed with bad values")
	}
}

func TestUintFuncArguments(t *testing.T) {
	nmin, nmax, err := uintFuncArguments("1", "2", "range")
	if err != nil {
		t.Fatalf("Error returned from uintFuncArguments with good values: %v", err)
	}
	if nmin != 1 || nmax != 2 {
		t.Fatalf("Error invalid min/max returned")
	}

	nmin, nmax, err = uintFuncArguments("-1", "2", "range")
	if err == nil {
		t.Fatalf("uintFuncArguments passed with bad values: %d %d", nmin, nmax)
	}

	_, _, err = uintFuncArguments("1.0", "2", "range")
	if err == nil {
		t.Fatalf("uintFuncArguments passed with bad values")
	}

	_, _, err = uintFuncArguments("4", "1", "range")
	if err == nil {
		t.Fatalf("uintFuncArguments passed with bad values")
	}
}

func TestFloatFuncArguments(t *testing.T) {
	nmin, nmax, err := floatFuncArguments("1.003", "2.004", "range")
	if err != nil {
		t.Fatalf("Error returned from floatFuncArguments with good values: %v", err)
	}
	if nmin != 1.003 || nmax != 2.004 {
		t.Fatalf("Error invalid min/max returned")
	}

	_, _, err = floatFuncArguments("a", "2.0001", "range")
	if err == nil {
		t.Fatalf("floatFuncArguments passed with bad values")
	}

	_, _, err = floatFuncArguments("1.0", "a", "range")
	if err == nil {
		t.Fatalf("floatFuncArguments passed with bad values")
	}

	_, _, err = floatFuncArguments("0.1", "0.0001", "range")
	if err == nil {
		t.Fatalf("floatFuncArguments passed with bad values")
	}
}

func TestRegexMatch(t *testing.T) {
	r := regexFromString("^(O|R)$")

	if err := r.Validate("testParam", "O"); err != nil {
		t.Fatalf("Error regex didn't pass validation: %v", err)
	}

	if err := r.Validate("testParam", "z"); err == nil {
		t.Fatalf("Error invalid regex passed validation")
	}
}

func regexFromString(regex string) *regexValidate {
	p := regexp.MustCompile(regex)
	return &regexValidate{Pattern: p, MatchType: regexMatch}
}
