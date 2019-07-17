package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

func TestEmeApply(t *testing.T) {
	b := new(bytes.Buffer)
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	customMiddleware := func(l string) func(http.Handler) http.Handler {
		return func(f http.Handler) http.Handler {
			fmt.Fprint(b, l)
			return f
		}
	}
	m := customMiddleware("m")
	n := customMiddleware("n")
	tests := []struct {
		name           string
		mws            []Mw
		expectedResult string
	}{
		{
			"Calls all middlewares",
			[]Mw{m, n, m, n},
			"nmnm",
		},
		{
			"Returns HandlerFunc in no middleware",
			[]Mw{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := NewEme(tt.mws...).Apply(hf)

			if b.String() != tt.expectedResult {
				t.Errorf("Expected %q got %q", tt.expectedResult, b.String())
			}

			// Testing that the handler function is getting returned
			actual := fmt.Sprintf("%v", h)
			expected := fmt.Sprintf("%v", http.HandlerFunc(hf))
			if actual != expected {
				t.Errorf("Expected %s to be %s", actual, expected)
			}
			b.Reset()
		})
	}
}

func TestEmeApplyFunc(t *testing.T) {
	b := new(bytes.Buffer)
	f := func(w http.ResponseWriter, r *http.Request) {}
	m := func(f http.Handler) http.Handler {
		fmt.Fprint(b, "m")
		return f
	}
	tests := []struct {
		name           string
		mws            []Mw
		expectedResult string
	}{
		{
			"Calls all middlewares",
			[]Mw{m, m, m},
			"mmm",
		},
		{
			"Returns HandlerFunc in no middleware",
			[]Mw{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := NewEme(tt.mws...).ApplyFunc(f)

			if b.String() != tt.expectedResult {
				t.Errorf("Expected %q got %q", tt.expectedResult, b.String())
			}

			// Testing that the handler function is getting returned
			actual := fmt.Sprintf("%v", h)
			expected := fmt.Sprintf("%v", http.HandlerFunc(f))
			if actual != expected {
				t.Errorf("Expected %s to be %s", actual, expected)
			}
			b.Reset()
		})
	}
}
