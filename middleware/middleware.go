package middleware

import (
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/tomascaslo/godinez"
	"log"
	"net/http"
)

type logHolder interface {
	GetInfoLogger() *log.Logger
	GetErrorLogger() *log.Logger
}

type applicationAuthenticator interface {
	IsAuthenticated(*http.Request) bool
	GetRedirectTo() string
}

func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func LogRequest(lh logHolder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lh.GetInfoLogger().Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

			next.ServeHTTP(w, r)
		})
		return fn
	}
}

func RecoverPanic(lh logHolder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					godinez.ServerError(lh.GetErrorLogger(), w, fmt.Errorf("%s", err))
				}
			}()
			next.ServeHTTP(w, r)
		})

		return fn
	}
}

func RequireAuthentication(app applicationAuthenticator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !app.IsAuthenticated(r) {
				fmt.Println("Testing redirect")
				http.Redirect(w, r, app.GetRedirectTo(), http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
		return fn
	}
}

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}
