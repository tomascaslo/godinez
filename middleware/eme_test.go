package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

func TestEmeCallsAllMiddlewares(t *testing.T) {
	b := new(bytes.Buffer)
	f := func(w http.ResponseWriter, r *http.Request) {}
	m := func(f http.Handler) http.Handler {
		fmt.Fprint(b, "m")
		return f
	}

	Eme(f, m, m, m)

	expected := "mmm"
	actual := b.String()
	if actual != expected {
		t.Errorf("Expected %q to be %q", actual, expected)
	}
}

func TestEmeWithNoMiddlewaresReturnsHandler(t *testing.T) {
	f := func(w http.ResponseWriter, r *http.Request) {}

	actual := fmt.Sprintf("%v", Eme(f))
	expected := fmt.Sprintf("%v", http.HandlerFunc(f))

	if actual != expected {
		t.Errorf("Expected %s to be %s", actual, expected)
	}
}
