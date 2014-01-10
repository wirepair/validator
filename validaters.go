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
	"sync"
)

type ValidatorTypeError struct {
	Func  string // description of function
	Param string // the parameter name
	Type  string // the value type
}

// Called when a function is defined for the wrong type of parameter value.
func (e *ValidatorTypeError) Error() string {
	return "validate: error " + e.Func + " function for " + e.Param + " is invalid for " + e.Type
}

type ValidatorFuncError struct {
	Value string // the value being validated
	Type  string // the type of the field
	Name  string // function name of the validator
}

func (e *ValidatorFuncError) Error() string {
	return "validate: error " + e.Value + " for " + e.Type + " invalid value for " + e.Name
}

type ValidationError struct {
	Value string // the value being validated
	Param string // the Parameter name
}

// Called when the input fails validation for the Validater.
func (e *ValidationError) Error() string {
	return "validate: error param " + e.Param + " failed validation with value " + e.Value
}

type ValidateTagError struct {
	Tag   string // the tag key that failed (regex/validate)
	Field string // the Field name that caused the tag validation error
}

// Returned when a tag for a field did not parse properly.
func (e *ValidateTagError) Error() string {
	return "validate: error " + e.Tag + " for " + e.Field + " was not set correctly."
}

// An interface which describes a Validater. The string is the parameter name from the input map, the interface{} is the value to validate.
type Validater interface {
	Validate(string, interface{}) error // Returns error if validation fails.
}

// contains our function -> Validater mappings.
type validatorFunctions struct {
	sync.RWMutex
	Funcs map[string]func(string) error
}

var userFns *validatorFunctions

// Adds a new validater function type to allow custom validators to be used.
// Simply pass in the function name as a key to use in the struct tag which
// will be used to look up and execute the function when validation is set
// to occur. Note you must call this prior to running validation on a struct
// which uses the function.
func Add(fn string, validateFn func(string) error) error {
	if fn == "optional" || fn == "range" || fn == "len" {
		return fmt.Errorf("validate: error supplied function %s matches built in name", fn)
	}

	if userFns == nil {
		userFns = &validatorFunctions{}
	}

	if validateFn != nil {
		userFns.Lock()
		if userFns.Funcs == nil {
			userFns.Funcs = map[string]func(string) error{}
		}
		userFns.Funcs[fn] = validateFn
		userFns.Unlock()
	}
	return nil
}

// setDirectives validates each field individually. Returns ValidateTagError if
// we see the key in the tag as a string but fail to get the value with Get
func setDirectives(t reflect.StructTag, f *field) error {
	f.validators = make([]Validater, 0)

	tag := string(t)

	validate := t.Get("validate")
	if validate == "" && strings.Contains(tag, "validate") {
		return &ValidateTagError{Tag: "validate", Field: f.name}
	} else {
		if err := parseValidate(validate, f); err != nil {
			return err
		}
	}

	regex := t.Get("regex")
	if regex == "" && strings.Contains(tag, "regex") {
		return &ValidateTagError{Tag: "regex", Field: f.name}
	} else {
		if err := parseRegex(regex, f); err != nil {
			return err
		}
	}

	return nil
}

const (
	regexMatch = iota
	regexFind
)

// parseRegex extracts the pattern from the regex key.
func parseRegex(reg string, f *field) error {
	// allow either matching or finding. default is MatchString, probably don't need find.
	regexType := regexMatch
	if strings.HasPrefix(reg, "find,") {
		regexType = regexFind
		reg = reg[5:]
	} else if strings.HasPrefix(reg, "match,") {
		reg = reg[6:]
	}

	pattern, err := regexp.Compile(reg)
	if err != nil {
		return err
	}

	f.validators = append(f.validators, &regexValidate{MatchType: regexType, Pattern: pattern})
	return nil
}

// parses the validate struct tag and sets the field parameter name, whether it is optional
// and any Validator functions (including user supplied).
func parseValidate(values string, f *field) error {
	directives := strings.Split(values, ",")
	if len(directives) <= 0 {
		return nil
	}

	kind := f.typ.Kind()
	if kind == reflect.Slice {
		kind = f.typ.Elem().Kind()
	}

	f.param = directives[0] // first field is always the map key.
	for i := 1; i < len(directives); i++ {
		if directives[i] == "optional" {
			f.optional = true
		} else if strings.HasPrefix(directives[i], "range") {
			rangeValidator, err := newRangeValidator(directives[i], "range", f, kind)
			if err != nil {
				return err
			}

			f.validators = append(f.validators, rangeValidator)

		} else if strings.HasPrefix(directives[i], "len") {
			lenValidator, err := newLenValidator(directives[i], "len", f, kind)
			if err != nil {
				return err
			}
			f.validators = append(f.validators, lenValidator)
		} else {
			// check custom user functions
			if userFns != nil {
				userFns.RLock()
				if userFns.Funcs[directives[i]] != nil {
					userValidator := &userValidate{validateFn: userFns.Funcs[directives[i]]}
					f.validators = append(f.validators, userValidator)
				}
				userFns.RUnlock()
			} else {
				return fmt.Errorf("validate: error unknown validation function %s\n", directives[i])
			}
		}
	}
	return nil
}

// newLenValidator validates the length of a string.
func newLenValidator(input, fname string, f *field, kind reflect.Kind) (Validater, error) {
	// len only works on strings.
	if kind != reflect.String {
		return nil, &ValidatorTypeError{Func: "len", Param: f.param, Type: kind.String()}
	}

	min, max, err := getArguments(input, fname)
	if err != nil {
		return nil, err
	}

	nmin, err := strconv.Atoi(min)
	if err != nil {
		return nil, &ValidatorFuncError{Value: min, Type: kind.String(), Name: fname}
	}
	nmax, err := strconv.Atoi(max)
	if err != nil {
		return nil, &ValidatorFuncError{Value: max, Type: kind.String(), Name: fname}
	}
	if nmax < nmin {
		return nil, &ValidatorFuncError{Value: "max " + max + " < " + min + " min", Type: kind.String(), Name: fname}
	}

	return &lenValidate{Min: nmin, Max: nmax}, nil
}

// newRangeValidator validates that a numerical value falls with in the specified range.
func newRangeValidator(input, fname string, f *field, kind reflect.Kind) (Validater, error) {
	// can't do ranges on strings.
	if kind == reflect.String {
		return nil, &ValidatorTypeError{Func: "range", Param: f.param, Type: kind.String()}
	}
	min, max, err := getArguments(input, fname)
	if err != nil {
		return nil, err
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		nmin, nmax, errInt := intFuncArguments(min, max, fname)
		if errInt != nil {
			return nil, errInt
		}
		return &rangeIntValidate{Min: nmin, Max: nmax}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		nmin, nmax, errUint := uintFuncArguments(min, max, fname)
		if errUint != nil {
			return nil, errUint
		}
		return &rangeUintValidate{Min: nmin, Max: nmax}, nil
	case reflect.Float32, reflect.Float64:
		nmin, nmax, errFloat := floatFuncArguments(min, max, fname)
		if errFloat != nil {
			return nil, errFloat
		}
		return &rangeFloatValidate{Min: nmin, Max: nmax}, nil
	default:
		return nil, fmt.Errorf("validate: error %s is not a supported type for function %s.", kind.String(), fname)
	}
}

func getArguments(data, fname string) (string, string, error) {
	r := data[strings.Index(data, "(")+1 : strings.Index(data, ")")]
	vals := strings.Split(r, ":")
	if len(vals) != 2 {
		return "", "", fmt.Errorf("validate: invalid number of arguments to %s validator function", fname)
	}
	return vals[0], vals[1], nil
}

func intFuncArguments(min, max, fname string) (int64, int64, error) {
	nmin, err := strconv.ParseInt(min, 10, 64)
	if err != nil {
		return -1, -1, &ValidatorFuncError{Value: min, Type: "Int", Name: fname}
	}
	nmax, err := strconv.ParseInt(max, 10, 64)
	if err != nil {
		return -1, -1, &ValidatorFuncError{Value: max, Type: "Int", Name: fname}
	}
	if nmax < nmin {
		return -1, -1, &ValidatorFuncError{Value: "max " + max + " < " + min + " min", Type: "Int", Name: fname}
	}
	return nmin, nmax, nil
}

func uintFuncArguments(min, max, fname string) (uint64, uint64, error) {
	nmin, err := strconv.ParseUint(min, 10, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: "Uint", Name: fname}
	}

	nmax, err := strconv.ParseUint(max, 10, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: "Uint", Name: fname}
	}

	if nmax < nmin {
		return 0, 0, &ValidatorFuncError{Value: "max " + max + " < " + min + " min", Type: "Uint", Name: fname}
	}
	return nmin, nmax, nil
}

func floatFuncArguments(min, max, fname string) (float64, float64, error) {
	nmin, err := strconv.ParseFloat(min, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: min, Type: "Float", Name: fname}
	}
	nmax, err := strconv.ParseFloat(max, 64)
	if err != nil {
		return 0, 0, &ValidatorFuncError{Value: max, Type: "Float", Name: fname}
	}
	if nmax < nmin {
		return 0, 0, &ValidatorFuncError{Value: "max " + max + " < " + min + " min", Type: "Float", Name: fname}
	}
	return nmin, nmax, nil
}

type rangeIntValidate struct {
	Min int64
	Max int64
}

func (r *rangeIntValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Int()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatInt(val, 10)}
	}
	return nil
}

type rangeUintValidate struct {
	Min uint64
	Max uint64
}

func (r *rangeUintValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Uint()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatUint(val, 10)}
	}
	return nil
}

type rangeFloatValidate struct {
	Min float64
	Max float64
}

func (r *rangeFloatValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.Float()
	if val < r.Min || val > r.Max {
		return &ValidationError{Param: param, Value: strconv.FormatFloat(val, 'e', 10, 64)}
	}
	return nil
}

type lenValidate struct {
	Min int
	Max int
}

func (r *lenValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	val := v.String()
	l := len(val)

	if l < r.Min || l > r.Max {
		return &ValidationError{Param: param, Value: val}
	}
	return nil
}

type regexValidate struct {
	Pattern   *regexp.Regexp
	MatchType int
}

func (r *regexValidate) Validate(param string, value interface{}) error {
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

// for wrapping custom user functions.
type userValidate struct {
	validateFn func(string) error
}

// validates the input against a custom user function.
func (u *userValidate) Validate(param string, value interface{}) error {
	v := reflect.ValueOf(value)
	return u.validateFn(v.String())
}
