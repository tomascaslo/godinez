// Package middleware provides convenient general and middleware functions
// to simplify API development.
package middleware

import (
	"net/http"
)

type Mw func(http.Handler) http.Handler

type Eme struct {
	mws []Mw
}

// NewEme initializes a *Eme struct with the provided mws.
// Middleware is applied from left to right, so order is important.
func NewEme(mws ...Mw) *Eme {
	return &Eme{append([]Mw{}, mws...)}
}

// Apply runs middleware and returns the http.HandlerFunc
// by calling do().
func (e *Eme) Apply(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return do(f, e.mws...)
}

// Do applies middlewares mws to the given function f.
// Middlewares are applied from left to right, in order.
// If no middlewares are passed http.HandlerFunc(f) is returned.
// Use do(func(http.ResponseWriter, *http.Request), mw1, mw2, mw3...)
func do(f func(http.ResponseWriter, *http.Request), mws ...Mw) http.Handler {
	hf := http.HandlerFunc(f)
	if len(mws) < 1 {
		return hf
	}
	var h http.Handler
	for _, m := range mws {
		h = m(hf)
	}

	return h
}
