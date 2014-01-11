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

// This library is for automatically assigning HTTP form values or a map[string][]string
// to a pre-defined structure. It also allows you to validate the data prior to allowing
// assignment to occur. If any field is found to fail validation, an error is immediately
// returned and further processing is stopped. Additionally, you may supply your own
// functions by calling Add. For more information and examples see:
// https://github.com/wirepair/validator/
package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

type TypeError struct {
	Value string       // description of value that caused the error
	Param string       // the parameter name
	Type  reflect.Type // type of Go value it could not be assigned to
}

// Returned when validator is unable to get the proper type from the supplied map of parameters and values.
func (e *TypeError) Error() string {
	return "validate: error parsing parameter " + e.Param + " with value " + e.Value + " into Go value of type " + e.Type.String()
}

type field struct {
	name       string
	param      string
	tags       string
	typ        reflect.Type
	optional   bool
	index      int
	validators []Validater
}

type cache struct {
	sync.RWMutex
	m map[reflect.Type][]field
}

var fieldCache cache // for caching field look ups.

// Assign iterates over input map keys and assigns the value to the passed in structure (v),
// alternatively validating the input.
func Assign(params map[string][]string, v interface{}) error {
	fields, err := getFields(v)
	if err != nil {
		return err
	}

	return assign(params, fields, v)
}

// iterates over each field of the structure and assigns various directives on how to
// parse, validate and process the value to be assigned to that field.
// for performance reasons we also store field lookups in a synchronized cache so
// if we get the same struct many times we only have to analyze the structtags a single
// time.
func getFields(v interface{}) ([]field, error) {
	var err error
	cacheKey := reflect.TypeOf(v)

	fieldCache.RLock()
	if fieldCache.m == nil {
		fieldCache.m = make(map[reflect.Type][]field, 1)
	}
	f := fieldCache.m[cacheKey]
	fieldCache.RUnlock()
	if f != nil {
		//fmt.Printf("We got a cache hit! on %v cache len: %d\n", cacheKey, len(fieldCache.m))
		return f, nil
	}

	st := reflect.TypeOf(v).Elem()
	fields := make([]field, st.NumField())

	for i := 0; i < st.NumField(); i++ {
		f := &field{}
		f.typ = st.Field(i).Type
		f.name = st.Field(i).Name
		f.index = i

		// sets param,optional flags and validators.
		err = setDirectives(st.Field(i).Tag, f)
		if err != nil {
			return nil, err
		}
		fields[i] = *f
	}

	fieldCache.Lock()
	fieldCache.m[cacheKey] = fields
	fieldCache.Unlock()

	return fields, nil
}

// assign validates fields are settable, parameters aren't empty and that fields set
// as optional are validated (unless empty, then disregarded).
func assign(params map[string][]string, fields []field, v interface{}) (err error) {
	st := reflect.ValueOf(v).Elem()

	for _, f := range fields {
		// skip parameters which don't have validate markup
		if f.param == "" {
			continue
		}
		settable := st.Field(f.index)
		if !settable.CanSet() {
			return fmt.Errorf("validate: error struct field %s is not settable\n", f.name)
		}

		values := params[f.param]
		size := len(values)

		// check if the parameter is required or not.
		if size == 0 && f.optional == false {
			return fmt.Errorf("validate: error parameter %s does not exist in input.\n", f.param)
		} else if (size == 0 || size == 1 && values[0] == "") && f.optional == true {
			continue
		}

		if settable.Kind() == reflect.Slice {
			//fmt.Printf("Making slice\n")
			settable.Set(reflect.MakeSlice(reflect.SliceOf(settable.Type().Elem()), size, size))
			for i, v := range values {
				if err := verifiedAssign(v, &f, settable.Index(i)); err != nil {
					return err
				}
			}
		} else {
			// only take the first verify & assign value.
			err = verifiedAssign(values[0], &f, settable)
		}
		// we got an error assigning a type or array, error out.
		if err != nil {
			return err
		}
	}
	return nil
}

// verifiedAssign will take the input string, determine it's type via reflection.
// Then it will run validators against the reflected type to make sure they pass.
// provided they do, the value will be assigned to the structure.
// NOTE: we also check for numerical overflows.
func verifiedAssign(s string, f *field, settable reflect.Value) error {

	switch settable.Kind() {
	case reflect.String:
		//fmt.Printf("In string case validators len: %d\n", len(f.validation.Validaters))
		for _, validater := range f.validators {
			if err := validater.Validate(f.param, s); err != nil {
				return err
			}
		}
		settable.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || settable.OverflowInt(n) {
			return &TypeError{f.param, s, settable.Type()}
		}

		for _, validater := range f.validators {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || settable.OverflowUint(n) {
			return &TypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validators {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, settable.Type().Bits())
		if err != nil || settable.OverflowFloat(n) {
			return &TypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validators {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetFloat(n)
	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return &TypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validators {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetBool(n)
	default:
		return fmt.Errorf("validate: error %v is not a supported type for parameter %s.", settable.Type(), f.param)
	}
	return nil
}
