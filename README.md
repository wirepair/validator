# validator 
This library is for automatically assigning HTTP form values or a map[string][]string to a pre-defined structure. It also allows you to validate the data prior to allowing assignment to occur. It fails hard, if any field is found to fail validation, an error is immediately returned. 

### installation
go get github.com/wirepair/validator

### example  
Usage is pretty simple, simply define your structure with the "validate" struct tag along with the parameter name and then a validation function if you want it.
```Go
package main

import (
	"fmt"
	"github.com/wirepair/validator"
	"html/template"
	"net/http"
)

const (
	userForm = `<html>
	<head></head>
	<body><form action="/form" method="POST">
		Name: <input type="text" name="name" value="bob"></input><br>
		Age: <input type="text" name="age" value=""></input><br>
		State: <input type="text" name="state" value="AZ"></input><br>
		<input type="submit" value="Submit"></input>
	</form>
	</body>`

	userPage = `<html>
	<head></head>
	<body>
		Name: {{.Name}}<br>
		Age: {{.Age}}<br>
		State: {{.State}}<br>
		Internal: {{.Internal}}<br>
		Error: {{.Error}}<br>
	</body>`
)

type User struct {
	// be careful with invalid escapes in regexes! \w will fail (\\w is correct)
	Name     string `validate:"name" regex:"^[a-z]*$"`
	Age      int    `validate:"age,optional"`
	State    string `validate:"state,len(2:2)" regex:"^[A-Za-z]*$"`
	Internal string
	Error    string
}

// call parse form first, then validate and use
func HttpFormHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form values!", http.StatusInternalServerError)
		return
	}

	user := &User{}
	if err := validator.VerifiedAssign(r.Form, user); err != nil {
		user.Error = err.Error()
	}
	// Note you really shouldn't do this. If you get validation errors, throw it away and ask the user again.
	user.Internal = "this is was ignored by VerifiedAssign..."
	tmpl, err := template.New("user").Parse(userPage)

	err = tmpl.Execute(w, user)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func FormPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, userForm)
}

func main() {
	http.HandleFunc("/", FormPage)
	http.HandleFunc("/form", HttpFormHandler)
	http.ListenAndServe(":8080", nil)
}
```

## gotchas
Struct tags are very unforgiving, if you get any part of your struct tag definition incorrect, an error will be returned stating which field was incorrectly configured.
```Go
type BadUser struct {
	// BAD there is a space between regex: and the value
	Name     string `validate:"name" regex: "^[a-z]*$"`
}

type GoodUser struct {
	// GOOD tags are aligned properly with double quotes, no spacing and the correct specifiers used.
	Name     string `validate:"name" regex:"^[a-z]*$"`
}

type BadRegexUser struct {
	// BAD there the regex contains and incorrect escape sequence
	Name     string `validate:"name" regex:"^\w*$"`	
}

type GoodRegexUser struct {
	// GOOD the \w specifier is properly escaped.
	Name     string `validate:"name" regex:"^\\w*$"`	
}
```

Other things of note, unexported structure fields will not work.
```Go
type BadUnexportedUser struct {
	// BAD, name is not exported so it is not possible with reflection to set the unexported name field.
	name     string `validate:"name" regex:"^[a-z]*$"`
}
```

If you don't want a value set, just don't use any struct tags on the field, this is perfectly ok:
```Go
type UnVerifiedFieldsUser struct {
	IwantThis         string `validate:"this" regex:"^[a-z]*$"`
	IdontWantThis     string 
	AndNotThis        string
}
```

#### validate tag functions
Currently only two validation functions exist:
- len(min,max)  This will validate strings (or each individual slice of type string) is > minimum length and < maximum length. 
- range(min,max) This will validate that Int, Uint and Floats fall with in a specified range. 

```Go
// Example structure which takes the "name" parameter and validates it is > 4 characters and < 20 characters
// Age is 'optional' as in, if it doesn't exist in the original map as a key, we can safely disregard it. 
// If it does exist, it will still be validated.
// Balance will be assigned as a float provided it parses correctly into a float value. if not it will fail and
// no values will be returned.
type PersonForm struct {
	Name string `validate:"name,len(4:20)"`
	Age int `validate:"age,range(0:120),optional"`
	Balance float `validate:"balance,range(0,4000000.0)"`
}
```

#### custom functions
You may define your own validators to be used by calling validator.Add(key, function). Note that this must occur prior to calling VerifiedAssign on the structure otherwise it won't exist and an error will be returned stating an unknown function is defined. Note that the value will be passed as a string, so it is up to you to reflect it to the correct type. The validator must follow the format of: func userValidator(input string) error.

Example:
```Go
package main

import (
	"encoding/hex"
	"github.com/wirepair/validator"
	"log"
	"net/url"
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

func main() {
	hashForm := &HashForm{}
	validator.Add("hash", hashCheck)
	formValues, _ := url.ParseQuery("h=d83582b40325d7a3d723f05307b7534a")
	err := validator.VerifiedAssign(formValues, hashForm)
	if err != nil {
		log.Fatalf("Error occurred parsing hash: %v", err)
	}
}
```

#### regex tag functions
Currently match (calls MatchString) are supported for strings (or each slice of a slice of strings).
```Go
// Example structure which takes the "name" parameter and validates it is > 4 characters and < 20 characters and
// matches "john doe"
type RegexForm struct {
	FirstName string `validate:"name,len(4:20)" regex:"match,^(john)$"`
	// same as above, but don't need the match, part.
	LastName string `validate:"name,len(4:20)" regex:"^(doe)$"`
}
```


#### more examples?
See the unit tests!
