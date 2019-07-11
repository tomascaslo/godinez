package forms

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	values := url.Values{}
	values.Add("name", "john")
	expectedForm := &Form{values, errors(map[string][]string{})}
	form := New(values)

	if !reflect.DeepEqual(form, expectedForm) {
		t.Errorf("Expected %v got %v", expectedForm, form)
	}
}

func TestRequired(t *testing.T) {
	values := url.Values{}
	values.Add("name", "john")
	values.Add("email", "")
	form := New(values)

	form.Required("email")

	actual := form.Errors.Get("email")
	expected := "This field cannot be blank"
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = form.Errors.Get("name")
	expected = ""
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}

func TestMaxLength(t *testing.T) {
	values := url.Values{}
	values.Add("name", "john")
	values.Add("nickname", "xxxxxxxxxxxxxxx")
	form := New(values)

	form.MaxLength("nickname", 10)

	actual := form.Errors.Get("nickname")
	expected := "This field is too long (maximum is 10 characters)"
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = form.Errors.Get("name")
	expected = ""
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}

func TestPermittedValues(t *testing.T) {
	values := url.Values{}
	values.Add("sex", "female")
	values.Add("occupation", "student")
	sexPermittedValues := []string{"male", "female"}
	occupationPermittedValues := []string{"software engineer", "data analyst"}

	form := New(values)

	form.PermittedValues("sex", sexPermittedValues...)
	form.PermittedValues("occupation", occupationPermittedValues...)

	actual := form.Errors.Get("sex")
	expected := ""
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = form.Errors.Get("occupation")
	expected = fmt.Sprintf("This field is invalid. Permitted %v", occupationPermittedValues)
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}

func TestValid(t *testing.T) {
	values := url.Values{}
	values.Add("name", "john")
	values.Add("email", "")
	form := New(values)

	form.Required("name")
	form.Required("email")

	actual := form.Valid()
	expected := false
	if actual != expected {
		t.Errorf("Expected %t got %t", expected, actual)
	}

	delete(form.Errors, "email")
	actual = form.Valid()
	expected = true
	if actual != expected {
		t.Errorf("Expected %t got %t", expected, actual)
	}
}

func TestMinLength(t *testing.T) {
	values := url.Values{}
	values.Add("name", "john")
	values.Add("nickname", "xxxx")
	form := New(values)

	form.MinLength("nickname", 5)

	actual := form.Errors.Get("nickname")
	expected := "This field is too short (minimum is 5 characters)"
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = form.Errors.Get("name")
	expected = ""
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}

func TestMatchesPattern(t *testing.T) {
	values := url.Values{}
	values.Add("email", "john@test.com")
	values.Add("alternative email", "john@")

	form := New(values)

	form.MatchesPattern("email", EmailRX)
	form.MatchesPattern("alternative email", EmailRX)

	actual := form.Errors.Get("email")
	expected := ""
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = form.Errors.Get("alternative email")
	expected = "This field is invalid"
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}
