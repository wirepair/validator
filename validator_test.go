package validator

import (
	//"fmt"
	"net/url"
	"testing"
)

type SomeForm struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 255 chars
	Age  int    `validate:"age,range(1:10)"` // takes "age" http param and validates it's value is between 1 and 10.

	//Birth string `validate:"birth,,optional"` // not required.
}

func TestVerifiedAssign(t *testing.T) {
	val := make(map[string][]string, 2)
	person := &SomeForm{}
	strVals := []string{"John", "Doe"}
	intVal := []string{"1"}
	val["name"] = strVals
	val["age"] = intVal
	err := VerifiedAssign(val, person)
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
}

type CachedForm struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 255 chars
	Age  int    `validate:"age,range(1:10)"` // takes "age" http param and validates it's value is between 1 and 10.

	//Birth string `validate:"birth,,optional"` // not required.
}

type CachedForm2 struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 255 chars
	Age  int    `validate:"age,range(2:10)"` // takes "age" http param and validates it's value is between 1 and 10.

	//Birth string `validate:"birth,,optional"` // not required.
}

func TestCache(t *testing.T) {
	val := makeSimpleMap()
	cache1 := &CachedForm{}
	cache2 := &CachedForm{}
	cache3 := &CachedForm2{}
	cache4 := &CachedForm2{}

	err := VerifiedAssign(val, cache1)
	if err != nil {
		t.Fatalf("error parsing: %v\n", err)
	}

	VerifiedAssign(val, cache2)

	err = VerifiedAssign(val, cache3)
	if err == nil {
		t.Fatalf("error: bad input passed valiation.\n")
	}

	err = VerifiedAssign(val, cache4)
	if err == nil {
		t.Fatalf("error: bad input passed valiation.\n")
	}
}

type NonSettableField struct {
	name  string
	Name  string `validate:"name"`
	Zonks string `validate:"name,optional"`
}

func TestOptionalNonSettable(t *testing.T) {
	val := makeSimpleMap()
	nonSet := &NonSettableField{}
	err := VerifiedAssign(val, nonSet)
	if err != nil {
		t.Fatalf("error: something %v\n", err)
	}
	if nonSet.Name != "John" {
		t.Fatalf("Error Name didn't get set!\n")
	}
}

func TestUrls(t *testing.T) {
	blah, _ := url.ParseQuery("name=John&age=10")
	//fmt.Printf("data: %v\n", blah)
	st := &SomeForm{}
	err := VerifiedAssign(blah, st)
	if err != nil {
		t.Fatalf("Error parsing urls: %v\n", err)
	}
	blah, _ = url.ParseQuery("name=AAAAAAAAAAAAAA&age=10")
	st = &SomeForm{}
	err = VerifiedAssign(blah, st)
	if err == nil {
		t.Fatalf("Error name (over len) passed validation\n")
	}

	blah, _ = url.ParseQuery("name=AAA&age=11")
	st = &SomeForm{}
	err = VerifiedAssign(blah, st)
	if err == nil {
		t.Fatalf("Error age (over range) passed validation\n")
	}
}

type RegexForm struct {
	// be careful with invalid escapes in regexes! \s will fail and regex field will be ignored!
	Name string `validate:"name,len(4:20)" regex:"find,^(john\\sdoe)$"`
}

func TestRegex(t *testing.T) {
	blah, _ := url.ParseQuery("name=john doe")
	st := &RegexForm{}
	err := VerifiedAssign(blah, st)
	if err != nil {
		t.Fatal("Error in validation: %v\n", err)
	}

	blah, _ = url.ParseQuery("name=john Doe")
	st = &RegexForm{}
	err = VerifiedAssign(blah, st)
	if err == nil {
		t.Fatal("Error validation passed regex on bad value: \n")
	}

	// make sure len validation works along side regex
	blah, _ = url.ParseQuery("name=john Doeaaaaaaaaaaaaaaaa")
	st = &RegexForm{}
	err = VerifiedAssign(blah, st)
	if err == nil {
		t.Fatal("Error validation passed regex on bad value: \n")
	}
}

type User struct {
	Name  string `validate:"name" regex: "find,^(\\w*)$"`
	Age   int    `validate:"age,optional"`
	State string `validate:"state,len(2:2)" regex:"find,^([A-Za-z]*)$"`
}

func TestDocExamples(t *testing.T) {
	params, _ := url.ParseQuery("name=someone&state=AZ")
	st := &User{}

	err := VerifiedAssign(params, st)
	if err != nil {
		t.Fatalf("error: user did not parse properly: %v\n", err)
	}
	if st.State != "AZ" {
		t.Fatalf("error: state did not get assigned properly.\n")
	}
	if st.Name != "someone" {
		t.Fatalf("error: name did not get assigned properly.\n")
	}
}

func makeSimpleMap() map[string][]string {
	val := make(map[string][]string, 2)
	strVals := []string{"John", "Doe"}
	intVal := []string{"1"}
	val["name"] = strVals
	val["age"] = intVal
	return val
}
