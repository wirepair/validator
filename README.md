= validator =
This library is for automatically assigning HTTP form values or a map[string][]string to a pre-defined structure. It also allows you to validate the data prior to allowing assignment to occur. It fails hard, if any field is found to fail validation, an error is immediately returned. 

== Example == 
Usage is pretty simple, simply define your structure with the "validate" struct tag along with the parameter name and then a validation function.
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
	// be careful with invalid escapes in regexes! \w will fail (\\w is correct) and regex field will be ignored!
	Name     string `validate:"name" regex: "find,^(\\w*)$"`
	Age      int    `validate:"age,optional"`
	State    string `validate:"state,len(2:2)" regex:"find,^([A-Za-z]*)$"`
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
	user.Internal = "this is was ignored by VerifiedAssign..."
	tmpl, err := template.New("user").Parse(userPage)
	// Error checking elided
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

=== validate tag functions ===
Currently only two validation functions exist:
- len(min,max) : Will validate strings (or each individual slice of type string) is > minimum length and < maximum length.
- range(min,max) : Will validate Int, Uint, Float's fall with in a specified range. 

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

=== regex tag functions ===
Currently only find (Internally calls FindString on the input) and match (calls MatchString) are supported for strings (or each slice of a slice of strings).
```Go
// Example structure which takes the "name" parameter and validates it is > 4 characters and < 20 characters and
// matches "john doe"
type RegexForm struct {
	// be careful with invalid escapes in regexes! \s will fail and regex field will be ignored!
	Name string `validate:"name,len(4:20)" regex:"find,^(john\\sdoe)$"`
}
```