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
	//"errors"
	//"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	//"strings"
)

type ValidateTypeError struct {
	Value string       // description of value - "bool", "array", "number -5"
	Param string       // the parameer name
	Type  reflect.Type // type of Go value it could not be assigned to
}

func (e *ValidateTypeError) Error() string {
	return "validate: error parsing parameter " + e.Param + " with value " + e.Value + " into Go value of type " + e.Type.String()
}

type field struct {
	name       string
	param      string
	tags       string
	typ        reflect.Type
	optional   bool
	index      int
	validation *Validation
}

type cache struct {
	sync.RWMutex
	m map[reflect.Type][]field
}

var fieldCache cache // for caching field look ups.

func VerifiedAssign(params map[string][]string, v interface{}) error {
	fields, err := getFields(v)
	if err != nil {
		return err
	}
	if err := assign(params, fields, v); err != nil {
		v = nil // we had an error set v to nil
	}
	return err
}

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
		tags := make(map[string]string, 2)
		tags["regex"] = st.Field(i).Tag.Get("regex")
		tags["validate"] = st.Field(i).Tag.Get("validate")
		f := &field{}
		f.typ = st.Field(i).Type
		f.name = st.Field(i).Name
		f.index = i
		f.validation, err = NewValidation(tags, f.typ)
		if err != nil {
			return nil, err
		}
		f.param = f.validation.Param
		f.optional = f.validation.Optional
		fields[i] = *f
	}

	fieldCache.Lock()
	fieldCache.m[cacheKey] = fields
	fieldCache.Unlock()

	return fields, nil
}

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
		} else if size == 0 && f.optional == true {
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
			// only take the verify & assign first value.
			err = verifiedAssign(values[0], &f, settable)
		}
		// we got an error assigning a type or array, error out.
		if err != nil {
			return err
		}
	}
	return nil
}

func verifiedAssign(s string, f *field, settable reflect.Value) error {

	switch settable.Kind() {
	case reflect.String:
		//fmt.Printf("In string case validators len: %d\n", len(f.validation.Validaters))
		for _, validater := range f.validation.Validaters {
			if err := validater.Validate(f.param, s); err != nil {
				return err
			}
		}
		settable.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || settable.OverflowInt(n) {
			return &ValidateTypeError{f.param, s, settable.Type()}
		}

		for _, validater := range f.validation.Validaters {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || settable.OverflowUint(n) {
			return &ValidateTypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validation.Validaters {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, settable.Type().Bits())
		if err != nil || settable.OverflowFloat(n) {
			return &ValidateTypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validation.Validaters {
			if err := validater.Validate(f.param, n); err != nil {
				return err
			}
		}
		settable.SetFloat(n)
	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return &ValidateTypeError{f.param, s, settable.Type()}
		}
		for _, validater := range f.validation.Validaters {
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
