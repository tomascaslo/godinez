// Package middleware provides convenient general and middleware functions
// to simplify API development.
package middleware

import (
	"net/http"
)

// Eme applies middlewares mws to the given function f.
// Middlewares are applied from left to right, in order.
// If no middlewares are passed http.HandlerFunc(f) is returned.
// Use Eme(func(http.ResponseWriter, *http.Request, mw1, mw2, mw3...)
func Eme(f func(http.ResponseWriter, *http.Request), mws ...func(http.Handler) http.Handler) http.Handler {
	hf := http.HandlerFunc(f)
	if len(mws) < 1 {
		return hf
	}
	h := mws[0](hf)
	for _, mw := range mws[1:] {
		h = mw(hf)
	}

	return h
}
