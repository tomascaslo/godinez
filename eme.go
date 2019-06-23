// Package eme helps handle middleware in a more human-readable
// and friendly manner.
package eme

import (
	"net/http"
)

// Eme applies middlewares mws to the given function f.
// Middlewares are applied from left to right, in order.
// If no middlewares are passed http.HandlerFunc(f) is returned.
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
