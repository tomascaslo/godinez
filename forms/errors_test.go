package forms

import (
	"reflect"
	"testing"
)

func TestAdd(t *testing.T) {
	errs := errors{}
	expectedErrors := errors{"name": []string{"invalid", "length error"}}

	errs.Add("name", "invalid")
	errs.Add("name", "length error")
	if !reflect.DeepEqual(errs, expectedErrors) {
		t.Errorf("Expected %v got %v", expectedErrors, errs)
	}
}

func TestGet(t *testing.T) {
	errs := errors{}

	actual := errs.Get("name")
	expected := ""
	if actual != expected {
		t.Errorf("Expected %v got %v", expected, actual)
	}

	errs.Add("name", "invalid")
	errs.Add("name", "length error")

	actual = errs.Get("name")
	expected = "invalid"
	if actual != expected {
		t.Errorf("Expected %v got %v", expected, actual)
	}
}

