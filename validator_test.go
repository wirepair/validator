package validator

import (
	"encoding/hex"
	"net/url"
	"testing"
)

func hashCheck(hash string) error {
	// dumb example that simply sees if decodes properly.
	hashBytes := make([]byte, hex.DecodedLen(len(hash)))
	if _, err := hex.Decode(hashBytes, []byte(hash)); err != nil {
		return err
	}
	return nil
}

type HashForm struct {
	Hash string `validate:"h,hash"` // takes "h" http param and validates it decodes as valid hex
}

type HashFormTwo struct {
	Hash string `validate:"h,hash,len(32:32)"`
}

func TestCustomFunctions(t *testing.T) {
	// must call add first!
	Add("hash", hashCheck)
	hf := &HashForm{}
	val, _ := url.ParseQuery("h=d83582b40325d7a3d723f05307b7534a")
	err := Assign(val, hf)
	if err != nil {
		t.Fatalf("Error occurred parsing valid hash: %v", err)
	}

	hf = &HashForm{}
	val, _ = url.ParseQuery("h=d83582b40325d7a3d723f05307b7534ZZZ")
	err = Assign(val, hf)
	if err == nil {
		t.Fatalf("Error invalid hash passed validation\n")
	}

	// test that our built in functions work along side.
	hf2 := &HashFormTwo{}
	val, _ = url.ParseQuery("h=d83582b40325d7a3d723f05307b7534a")
	err = Assign(val, hf2)
	if err != nil {
		t.Fatalf("Error occurred parsing valid hash: %v", err)
	}

	hf2 = &HashFormTwo{}
	val, _ = url.ParseQuery("h=47503b623d7ffca7cc40fb0fc4ce53269b86f6b3")
	err = Assign(val, hf2)
	if err == nil {
		t.Fatalf("Error hash length > 32 but passed validation anyways!\n")
	}
}

type SomeForm struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 5 chars
	Age  int    `validate:"age,range(1:10)"` // takes "age" http param and validates it's value is between 1 and 10.
}

func TestAssignMap(t *testing.T) {
	val := make(map[string][]string, 2)
	person := &SomeForm{}
	strVals := []string{"John", "Doe"}
	intVal := []string{"1"}
	val["name"] = strVals
	val["age"] = intVal
	err := Assign(val, person)
	if err != nil {
		t.Fatalf("Error: %v\n", err)
	}
	if person.Name != "John" && person.Age != 1 {
		t.Fatalf("error invalid values returned: %v", person)
	}
}

type BadRegexField struct {
	Name string `validate:"breg" regex: "^(john)"`
}

func TestBadField(t *testing.T) {
	val, _ := url.ParseQuery("breg=john&breg2=john&breg3=john&bfns=john&bfcs=john&age=3")

	breg := &BadRegexField{}
	err := Assign(val, breg)
	switch err := err.(type) {
	case *TagError:
		// OK
	case nil:
		t.Fatalf("Error: space in regex definition passed.\n")
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}
}

type BadRegexFieldTwo struct {
	Name string `validate:"breg2" regex,"^(john)"`
}

func TestBadRegexFieldTwo(t *testing.T) {
	val, _ := url.ParseQuery("breg=john&breg2=john&breg3=john&bfns=john&bfcs=john&age=3")
	breg2 := &BadRegexFieldTwo{}
	err := Assign(val, breg2)
	switch err := err.(type) {
	case *TagError:
		// OK
	case nil:
		t.Fatalf("Error: comma seperator in regex passed.\n")
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}
}

type BadRegexFieldThree struct {
	Name string `validate:"breg3" regex:'^(john)'`
}

func TestBadRegexFieldThree(t *testing.T) {
	val, _ := url.ParseQuery("breg=john&breg2=john&breg3=john&bfns=john&bfcs=john&age=3")
	breg3 := &BadRegexFieldThree{}
	err := Assign(val, breg3)
	switch err := err.(type) {
	case *TagError:
		// OK
	case nil:
		t.Fatalf("Error: single quotes for regex passed.\n")
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}

}

type BadFieldCommaSeperated struct {
	Name string `validate:"bfcs",regex:"^(john)"`
}

func TestBadFieldCommaSeperated(t *testing.T) {
	val, _ := url.ParseQuery("breg=john&breg2=john&breg3=john&bfns=john&bfcs=john&age=3")
	bfcs := &BadFieldCommaSeperated{}
	err := Assign(val, bfcs)
	switch err := err.(type) {
	case *TagError:
		// OK
	case nil:
		t.Fatalf("Error: comma seperated tag keys passed.\n")
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}
}

type BadFieldInvalidRegexEscape struct {
	Name string `validate:"bfire",regex:"^(\sjohn)"`
}

func TestBadFieldInvalidRegexEscape(t *testing.T) {
	val, _ := url.ParseQuery("bfire=john&bfcs=john&age=3")
	bfcs := &BadFieldCommaSeperated{}
	err := Assign(val, bfcs)
	switch err := err.(type) {
	case *TagError:
		// OK
	case nil:
		t.Fatalf("Error: comma seperated tag keys passed.\n")
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}
}

// apparently this is OK.
type FieldNoSpace struct {
	Name string `validate:"bfns"regex:"^(john)"`
}

func TestFieldNoSpace(t *testing.T) {
	val, _ := url.ParseQuery("breg=john&breg2=john&breg3=john&bfns=john&bfcs=john&age=3")
	fns := &FieldNoSpace{}
	err := Assign(val, fns)
	switch err := err.(type) {
	case *TagError:
		t.Fatalf("Error: single no space between keys did not pass %v.\n", err)
	case nil:
		// OK
	default:
		t.Fatalf("error occurred but not validatetagerror: %v\n", err)
	}
}

type CachedForm struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 5 chars
	Age  int    `validate:"age,range(1:10)"` // takes "age" http param and validates it's value is between 1 and 10.

	//Birth string `validate:"birth,,optional"` // not required.
}

type CachedForm2 struct {
	// by default fields are required, to make optional use optional directive.
	Name string `validate:"name,len(0:5)"`   // takes "name" http param and validates it's < 5 chars
	Age  int    `validate:"age,range(2:10)"` // takes "age" http param and validates it's value is between 1 and 10.

	//Birth string `validate:"birth,,optional"` // not required.
}

func TestCache(t *testing.T) {
	val, _ := url.ParseQuery("name=john&age=3")
	cache1 := &CachedForm{}
	cache2 := &CachedForm{}
	cache3 := &CachedForm2{}
	cache4 := &CachedForm2{}

	err := Assign(val, cache1)
	if err != nil {
		t.Fatalf("error parsing on good input cache1: %v\n", err)
	}

	err = Assign(val, cache2)
	if err != nil {
		t.Fatalf("error parsing on good input cache2: %v\n", err)
	}

	val, _ = url.ParseQuery("name=john&age=23")
	err = Assign(val, cache3)
	if err == nil {
		t.Fatalf("error: bad input passed validation on cache3.\n")
	}

	err = Assign(val, cache4)
	if err == nil {
		t.Fatalf("error: bad input passed validation on cache4.\n")
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
	err := Assign(val, nonSet)
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
	err := Assign(blah, st)
	if err != nil {
		t.Fatalf("Error parsing urls: %v\n", err)
	}
	blah, _ = url.ParseQuery("name=AAAAAAAAAAAAAA&age=10")
	st = &SomeForm{}
	err = Assign(blah, st)
	if err == nil {
		t.Fatalf("Error name (> len) passed validation\n")
	}

	blah, _ = url.ParseQuery("name=AAA&age=11")
	st = &SomeForm{}
	err = Assign(blah, st)
	if err == nil {
		t.Fatalf("Error age (> range) passed validation\n")
	}
}

type RegexForm struct {
	// be careful with invalid escapes in regexes! \s will fail
	Name string `validate:"name,len(4:20)" regex:"^(john\\sdoe)$"`
}

func TestRegex(t *testing.T) {
	blah, _ := url.ParseQuery("name=john doe")
	st := &RegexForm{}
	err := Assign(blah, st)
	if err != nil {
		t.Fatal("Error in validation: %v\n", err)
	}

	blah, _ = url.ParseQuery("name=john Doe")
	st = &RegexForm{}
	err = Assign(blah, st)
	if err == nil {
		t.Fatal("Error validation passed regex on bad value.\n")
	}

	// make sure len validation works along side regex
	blah, _ = url.ParseQuery("name=john Doeaaaaaaaaaaaaaaaa")
	st = &RegexForm{}
	err = Assign(blah, st)
	if err == nil {
		t.Fatal("Error validation passed regex on bad value: \n")
	}
}

type User struct {
	Name  string `validate:"name" regex:"^[a-z]*$"`
	Age   int    `validate:"age,optional"`
	State string `validate:"state,len(2:2)" regex:"^[A-Za-z]*$"`
}

func TestDocExamples(t *testing.T) {
	params, _ := url.ParseQuery("name=someone&state=AZ")
	st := &User{}

	err := Assign(params, st)
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

type SliceyUser struct {
	Name []string `validate:"name" regex:"^[a-z]*$"`
	Age  []int    `validate:"age,range(1:10)"`
}

func TestSliceValidation(t *testing.T) {
	params, _ := url.ParseQuery("name=someone&name=else&name=zonks&age=1&age=2")
	st := &SliceyUser{}

	err := Assign(params, st)
	if err != nil {
		t.Fatalf("error: user did not parse properly: %v\n", err)
	}

	if st.Name[0] != "someone" && st.Name[1] != "zonks" {
		t.Fatalf("names set incorrectly.")
	}

	params, _ = url.ParseQuery("name=aaaa&name=1234&name=zonks&age=1&age=2")
	st = &SliceyUser{}
	err = Assign(params, st)

	if err == nil {
		t.Fatalf("error: name[1] should have failed validation\n")
	}

	params, _ = url.ParseQuery("name=aaaa&name=bbb&name=zonks&age=1&age=24")
	st = &SliceyUser{}
	err = Assign(params, st)

	if err == nil {
		t.Fatalf("error: age[1] should have failed validation\n")
	}
}

type RequiredUser struct {
	Name  string `validate:"name" regex:"^[a-z]*$"`
	Age   int    `validate:"age"`
	State string `validate:"state,len(2:2)" regex:"^[A-Za-z]*$"`
}

func TestRequired(t *testing.T) {
	params, _ := url.ParseQuery("name=someone&state=AZ")
	st := &RequiredUser{}

	err := Assign(params, st)
	if err == nil {
		t.Fatalf("error: age didn't exist in input and is required!\n")
	}
}

type Floatsies struct {
	Balance float32 `validate:"bal,range(0.00000003:0.000006)"`
}

func TestFloatsies(t *testing.T) {
	params, _ := url.ParseQuery("bal=0.00000004")
	st := &Floatsies{}
	err := Assign(params, st)
	if err != nil {
		t.Fatalf("error: float failed to parse even though it is valid: %v\n", err)
	}
	if st.Balance != 0.00000004 {
		t.Fatalf("error: float value not properly reflected, got: %f\n", st.Balance)
	}
}

func TestAssignSingle(t *testing.T) {
	params := map[string]string{}
	params["name"] = "john"
	params["age"] = "31"
	params["state"] = "MA"
	st := &RequiredUser{}
	err := AssignSingle(params, st)
	if err != nil {
		t.Fatalf("error: failed to parse even though it is valid: %v\n", err)
	}

	if st.Name != "john" {
		t.Fatalf("error: , got: %f\n", st.Name)
	}

	// test required age missing
	params = map[string]string{}
	params["name"] = "john"
	params["state"] = "MA"
	st = &RequiredUser{}
	err = AssignSingle(params, st)
	if err == nil {
		t.Fatalf("error: missing a required statement passed validation.\n")
	}

	switch err := err.(type) {
	case *RequiredParamError:
		//  OK!
	default:
		t.Fatalf("Error: got a different error back! %v\n", err)
	}
}

//HELPERS
func makeSimpleMap() map[string][]string {
	val := make(map[string][]string, 2)
	strVals := []string{"John", "Doe"}
	intVal := []string{"1"}
	val["name"] = strVals
	val["age"] = intVal
	return val
}
