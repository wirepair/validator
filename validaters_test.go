package validator

import (
	"reflect"
	"testing"
)

func TestParseRegexFind(t *testing.T) {
	validation := &Validation{}
	tag := "^(O|R)$"
	if err := parseRegex(tag, validation, reflect.TypeOf("")); err != nil {
		t.Fatalf("Error parsing regex: %v", err)
	}

	if err := validation.Validaters[0].Validate("test", "O"); err != nil {
		t.Fatalf("Error OK input didn't match regex: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", "Z"); err == nil {
		t.Fatalf("Error bad input matched regex")
	}

	validation = &Validation{}
	tag = "find,^(O|R)$"
	if err := parseRegex(tag, validation, reflect.TypeOf("")); err != nil {
		t.Fatalf("Error parsing regex: %v", err)
	}

	if err := validation.Validaters[0].Validate("test", "O"); err != nil {
		t.Fatalf("Error OK input didn't match regex: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", "Z"); err == nil {
		t.Fatalf("Error bad input matched regex")
	}
}

func TestParseRegexMatch(t *testing.T) {
	validation := &Validation{}
	tag := "match,^(O|R)$"
	if err := parseRegex(tag, validation, reflect.TypeOf("")); err != nil {
		t.Fatalf("Error parsing regex: %v", err)
	}

	if err := validation.Validaters[0].Validate("test", "O"); err != nil {
		t.Fatalf("Error OK input didn't match regex: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", "r"); err == nil {
		t.Fatalf("Error bad input matched regex\n")
	}
}

func TestParseValidateLen(t *testing.T) {
	validation := &Validation{}
	tag := "name,len(0:3)"

	if err := parseValidate(tag, validation, reflect.TypeOf("")); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	validation = &Validation{}
	tag = "name,len(0:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parse validate allowed len function on int\n")
	}

	validation = &Validation{}
	tag = "name,len(0:3:4)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,len((0:3)" // technically this is 'ok' due to how we parse...
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,len(:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (empty)\n")
	}

	validation = &Validation{}
	tag = "name,len(1,3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (invalid seperator)\n")
	}

	// test actual validation is OK
	validation = &Validation{}
	tag = "name,len(1:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err != nil {
		t.Fatalf("Error parsing valid args: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", "asd"); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", "asdfasdf"); err == nil {
		t.Fatalf("Error we didn't get an error for a > len input\n")
	}

	if err := validation.Validaters[0].Validate("test", ""); err == nil {
		t.Fatalf("Error we didn't get an error for a < len input\n")
	}
}

func TestParseValidateLenSlice(t *testing.T) {
	validation := &Validation{}
	tag := "name,len(0:3)"
	slicey := make([]string, 2)
	slicey[0] = "bop"
	slicey[1] = "zoop"
	if err := parseValidate(tag, validation, reflect.TypeOf(slicey)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", slicey[0]); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", slicey[1]); err == nil {
		t.Fatalf("Error > range passed validation.\n", err)
	}
}

func TestParseValidateRangeInt(t *testing.T) {
	validation := &Validation{}
	tag := "name,range(0:3)"

	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	validation = &Validation{}
	tag = "name,range(0:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parse validate allowed len function on string\n")
	}

	validation = &Validation{}
	tag = "name,range(0:3:4)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range((0:3)" // technically this is 'ok' due to how we parse...
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range(:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (empty)\n")
	}

	validation = &Validation{}
	tag = "name,range(1,3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (invalid seperator)\n")
	}

	// test actual validation is OK
	validation = &Validation{}
	tag = "name,range(1:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err != nil {
		t.Fatalf("Error parsing valid args: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 2); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 4); err == nil {
		t.Fatalf("Error we didn't get an error for a > range input\n")
	}

	if err := validation.Validaters[0].Validate("test", 0); err == nil {
		t.Fatalf("Error we didn't get an error for a < range input\n")
	}

	// test negative values.
	validation = &Validation{}
	tag = "name,range(-10:10)"
	if err := parseValidate(tag, validation, reflect.TypeOf(-1)); err != nil {
		t.Fatalf("Error valid values returned error: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", -5); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", -10); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 20); err == nil {
		t.Fatalf("Error we didn't get an error for a > range input\n")
	}

	if err := validation.Validaters[0].Validate("test", -20); err == nil {
		t.Fatalf("Error we didn't get an error for a < range input\n")
	}

}

func TestParseValidateRangeUint(t *testing.T) {
	validation := &Validation{}
	tag := "name,range(0:3)"

	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	validation = &Validation{}
	tag = "name,range(0:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parse validate allowed len function on string\n")
	}

	validation = &Validation{}
	tag = "name,range(0:3:4)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range((0:3)" // technically this is 'ok' due to how we parse...
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range(:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (empty)\n")
	}

	validation = &Validation{}
	tag = "name,range(1,3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (invalid seperator)\n")
	}

	// test actual validation is OK
	validation = &Validation{}
	tag = "name,range(1:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err != nil {
		t.Fatalf("Error parsing valid args: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 2); err != nil {
		t.Fatalf("Error %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 4); err == nil {
		t.Fatalf("Error we didn't get an error for a > range input\n")
	}

	if err := validation.Validaters[0].Validate("test", 0); err == nil {
		t.Fatalf("Error we didn't get an error for a < range input\n")
	}

	// test actual validation is OK
	validation = &Validation{}
	tag = "name,range(1:3)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1)); err != nil {
		t.Fatalf("Error parsing valid args: %v\n", err)
	}
	if err := validation.Validaters[0].Validate("test", -2); err == nil {
		t.Fatalf("Error negative value passed validation\n")
	}
}

func TestParseValidateRangeFloat(t *testing.T) {
	validation := &Validation{}
	tag := "name,range(0.0003:3.0)"

	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	validation = &Validation{}
	tag = "name,range(0.4:6.3)"
	if err := parseValidate(tag, validation, reflect.TypeOf("")); err == nil {
		t.Fatalf("Error parse validate allowed len function on string\n")
	}

	validation = &Validation{}
	tag = "name,range(1.2:3.3:4.4)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range((1:3.03)" // technically this is 'ok' due to how we parse...
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err == nil {
		t.Fatalf("Error parse validate allowed incorrect # of args\n")
	}

	validation = &Validation{}
	tag = "name,range(:3.0)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (empty)\n")
	}

	validation = &Validation{}
	tag = "name,range(1.0,3.0)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err == nil {
		t.Fatalf("Error parsing allowed for incorrect # of args (invalid seperator)\n")
	}

	// test actual validation is OK
	validation = &Validation{}
	tag = "name,range(0.001:3.0)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err != nil {
		t.Fatalf("Error parsing valid args: %v\n", err)
	}

	if err := validation.Validaters[0].Validate("test", 4.2); err == nil {
		t.Fatalf("Error we didn't get an error for a > range input\n")
	}

	if err := validation.Validaters[0].Validate("test", 0.0001); err == nil {
		t.Fatalf("Error we didn't get an error for a < range input\n")
	}

}

func TestParseValidateOptional(t *testing.T) {
	validation := &Validation{}
	tag := "name,range(0.0003:3.0),optional"

	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}

	if validation.Optional == false {
		t.Fatalf("Optional did not get set!")
	}

	if validation.Param != "name" {
		t.Fatalf("param did not get set to name!")
	}

	validation = &Validation{}
	tag = "name,optional,range(0.0003:3.0)"
	if err := parseValidate(tag, validation, reflect.TypeOf(1.0)); err != nil {
		t.Fatalf("Error parse validate on good values failed %v\n", err)
	}
	if validation.Optional == false {
		t.Fatalf("First directive of optional did not get set!")
	}

}
