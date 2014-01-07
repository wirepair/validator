/*
The MIT License (MIT)

Copyright (c) 2014 isaac dawson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidatorTypeError struct {
	Func  string // description of function
	Param string // the Parameter name
	Type  string
}

func (e *ValidatorTypeError) Error() string {
	return "validate: error " + e.Func + " function for " + e.Param + " is invalid for " + e.Type
}

type ValidatorFuncError struct {
	Value string // description of value - "bool", "array", "number -5"
	Type  string // the parameter name
	Name  string // function name
}

func (e *ValidatorFuncError) Error() string {
	return "validate: error " + e.Value + " for " + e.Type + " is invalid value for " + e.Name
}

type ValidationError struct {
	Value string // description of Value - "bool", "array", "number -5"
	Param string // the Parameter name
}

func (e *ValidationError) Error() string {
	return "validate: error param " + e.Param + " failed validation with value " + e.Value
}

type Validater interface {
	Validate(string, interface{}) error
}

type Validation struct {
	Optional   bool
	Param      string
	Validaters []Validater
}

func NewValidation(tag map[string]string, typ reflect.Type) (*Validation, error) {
	validation := &Validation{}
	validation.Validaters = make([]Validater, 0)

	for k, v := range tag {
		if v == "" {
			continue
		}
		switch k {
		case "validate":
			if err := parseValidate(v, validation, typ); err != nil {
				return nil, err
			}
		case "regex":
			if err := parseRegex(v, validation, typ); err != nil {
				return nil, err
			}
		}
	}
	return validation, nil
}

const (
	regexMatch = iota
	regexFind
)

func parseRegex(reg string, validation *Validation, typ reflect.Type) error {
	// allow either matching or finding. default is MatchString, probably don't need find.
	find := regexMatch
	if strings.HasPrefix(reg, "find,") {
		find = regexFind
		reg = reg[5:]
	} else if strings.HasPrefix(reg, "match,") {
		reg = reg[6:]
	}

	pattern, err := regexp.Compile(reg)
	if err != nil {
		return err
	}

	validation.Validaters = append(validation.Validaters, &RegexValidate{MatchType: find, Pattern: pattern})
	return nil
}

func parseValidate(values string, validation *Validation, typ reflect.Type) error {
	directives := strings.Split(values, ",")
	if len(directives) <= 0 {
		return nil
	}
	k := typ.Kind()
	if k == reflect.Slice {
		k = typ.Elem().Kind()
	}
	validation.Param = directives[0]
	for i := 1; i < len(directives); i++ {
		if directives[i] == "optional" {
			validation.Optional = true
		} else if strings.HasPrefix(directives[i], "range(") && strings.HasSuffix(directives[i], ")") {
			if k == reflect.String {
				return &ValidatorTypeError{Func: "range", Param: validation.Param, Type: typ.Kind().String()}
			}

			if validator, err := newRangeValidator(directives[i], "range", typ); err != nil {
				return err
			} else {
				validation.Validaters = append(validation.Validaters, validator)
			}
		} else if strings.HasPrefix(directives[i], "len(") && strings.HasSuffix(directives[i], ")") {
			if k != reflect.String {
				return &ValidatorTypeError{Func: "len", Param: validation.Param, Type: typ.Kind().String()}
			}

			if validator, err := newLenValidator(directives[i], "len", typ); err != nil {
				return err
			} else {
				validation.Validaters = append(validation.Validaters, validator)
			}
		} else {
			return fmt.Errorf("validate: error unknown validation function %s\n", directives[i])
		}
	}
	return nil
}

func newLenValidator(input, fname string, typ reflect.Type) (Validater, error) {
	min, max, err := getArguments(input, fname)
	if err != nil {
		return nil, err
	}
	nmin, err := strconv.Atoi(min)
	if err != nil {
		return nil, &ValidatorFuncError{Value: min, Type: typ.Kind().String(), Name: fname}
	}
	nmax, err := strconv.Atoi(max)
	if err != nil {
		return nil, &ValidatorFuncError{Value: max, Type: typ.Kind().String(), Name: fname}
	}
	if nmax < nmin {
		return nil, &ValidatorFuncError{Value: max + "<" + min, Type: typ.Kind().String(), Name: fname}
	}

	return &LenValidate{Min: nmin, Max: nmax}, nil
}

func newRangeValidator(input, fname string, typ reflect.Type) (Validater, error) {
	min, max, err := getArguments(input, fname)
	if err != nil {
		return nil, err
	}
	// if we are a slice, get the underlying type.
	t := typ.Kind()
	if t == reflect.Slice {
		t = typ.Elem().Kind()
	}
	switch t {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		nmin, nmax, errInt := intFuncArguments(min, max, fname)
		if errInt != nil {
			return nil, errInt
		}
		return &RangeIntValidate{Min: nmin, Max: nmax}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		nmin, nmax, errUint := uintFuncArguments(min, max, fname)
		if errUint != nil {
			return nil, errUint
		}
		return &RangeUintValidate{Min: nmin, Max: nmax}, nil
	case reflect.Float32, reflect.Float64:
		nmin, nmax, errFloat := floatFuncArguments(min, max, fname)
		if errFloat != nil {
			return nil, errFloat
		}
		return &RangeFloatValidate{Min: nmin, Max: nmax}, nil
	default:
		return nil, fmt.Errorf("validate: error %v is not a supported type for function %s.", t, fname)
	}
}

func intFuncArguments(min, max, fname string) (int64, int64, error) {
	nmin, err := strconv.ParseInt(min, 10, 64)
	if err != nil {
		return -1, -1, &ValidatorFuncError{Value: min, Type: reflect.Int.String(), Name: fname}
	}
	nmax, err := strconv.ParseInt(max, 10, 64)
	if err != nil {
		return -1, -1, &ValidatorFuncError{Value: max, Type: reflect.Int.String(), Name: fname}
	}
	if nmax < nmin {
		return -1, -1, &ValidatorFuncError{Value: max + "<" + min, Type: reflect.Int.String(), Name: fname}
	}
	return nmin, nmax, nil
}

func uintFuncArguments(min, max, fname string) (uint64, uint64, error) {
	nmin, err := strconv.ParseUint(min, 10, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: reflect.Uint.String(), Name: fname}
	}

	nmax, err := strconv.ParseUint(max, 10, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: reflect.Uint.String(), Name: fname}
	}

	if nmax < nmin {
		return 0, 0, &ValidatorFuncError{Value: max + "<" + min, Type: reflect.Uint.String(), Name: fname}
	}
	return nmin, nmax, nil
}

func floatFuncArguments(min, max, fname string) (float64, float64, error) {
	nmin, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: reflect.Float64.String(), Name: fname}
	}
	nmax, err := strconv.ParseFloat(max, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: max, Type: reflect.Float64.String(), Name: fname}
	}
	if nmax < nmin {
		return 0, 0, &ValidatorFuncError{Value: max + "<" + min, Type: reflect.Float64.String(), Name: fname}
	}
	return nmin, nmax, nil
}

func getArguments(data, fname string) (string, string, error) {
	r := data[strings.Index(data, "(")+1 : strings.Index(data, ")")]
	vals := strings.Split(r, ":")
	if len(vals) != 2 {
		return "", "", fmt.Errorf("validate: invalid number of arguments to %s validater function", fname)
	}
	return vals[0], vals[1], nil
}

type RangeIntValidate struct {
	Min int64
	Max int64
}

func (r *RangeIntValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Int()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatInt(val, 10)}
	}
	return nil
}

type RangeUintValidate struct {
	Min uint64
	Max uint64
}

func (r *RangeUintValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Uint()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatUint(val, 10)}
	}
	return nil
}

type RangeFloatValidate struct {
	Min float64
	Max float64
}

func (r *RangeFloatValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Float()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatFloat(val, 'e', 5, 64)}
	}
	return nil
}

type LenValidate struct {
	Min int
	Max int
}

func (r *LenValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.String()
	l := len(val)

	if l < r.Min || l > r.Max {
		return &ValidationError{Param: param, Value: val}
	}
	return nil
}

type RegexValidate struct {
	Pattern   *regexp.Regexp
	MatchType int
}

func (r *RegexValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.String()

	if r.MatchType == regexMatch {
		if matched := r.Pattern.MatchString(val); !matched {
			return &ValidationError{Param: param, Value: val}
		}
		// probably don't need regexFind
	} else if r.MatchType == regexFind {
		if found := r.Pattern.FindString(val); found == "" {
			return &ValidationError{Param: param, Value: val}
		}
	}

	return nil
}
