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

// ApplyFunc runs middleware with a function and returns the http.HandlerFunc
// by calling do().
func (e *Eme) ApplyFunc(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	hf := http.HandlerFunc(f)
	return do(hf, e.mws...)
}

// Apply runs middleware with a handler function and returns the http.HandlerFunc
// by calling do().
func (e *Eme) Apply(hf http.Handler) http.Handler {
	return do(hf, e.mws...)
}

// Do applies middlewares mws to the given function f.
// Middlewares are applied from left to right, in order.
// If no middlewares are passed http.HandlerFunc(f) is returned.
// Use do(func(http.ResponseWriter, *http.Request), mw1, mw2, mw3...)
func do(hf http.Handler, mws ...Mw) http.Handler {
	if len(mws) < 1 {
		return hf
	}
	for i := range mws {
		// Calling middleware functions in order
		hf = mws[len(mws)-1-i](hf)
	}

	return hf
}
