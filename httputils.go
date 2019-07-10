// Package godinez provides some useful functions for
// managing http requests and responses.
package godinez

import (
	"bytes"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type contextKey string

var ContextKeyIsAuthenticated = contextKey("isAuthenticated")

type application interface {
	getErrorLog() *log.Logger
	getSession() *sessions.Session
	getTemplateCache(string) (*template.Template, error)
	isAuthenticated(*http.Request) bool
}

type templateData interface {
	enableCSRFToken() bool
	enableCurrentYear() bool
	enableAuthentication() bool
	getTemplateData() interface{}
	getCSRFToken() string
	getCurrentYear() int
	getIsAuthenticated() bool
	setCSRFToken(string)
	setCurrentYear(int)
	setIsAuthenticated(bool)
}

func serverError(errorLog *log.Logger, w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func notFound(w http.ResponseWriter) {
	clientError(w, http.StatusNotFound)
}

func addDefaultData(app application, td templateData, r *http.Request) {
	if td.enableCSRFToken() {
		td.setCSRFToken(nosurf.Token(r))
	}
	if td.enableCurrentYear() {
		td.setCurrentYear(time.Now().Year())
	}
	if td.enableAuthentication() {
		td.setIsAuthenticated(app.isAuthenticated(r))
	}
}

func render(app application, td templateData, w http.ResponseWriter, r *http.Request, name string) {
	ts, err := app.getTemplateCache(name)
	if err != nil {
		serverError(app.getErrorLog(), w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)

	err = ts.Execute(buf, td.getTemplateData())
	if err != nil {
		fmt.Println("There was an error")
		serverError(app.getErrorLog(), w, err)
	}

	buf.WriteTo(w)
}

// Convenient function to check if request is authenticated with context.
// Method can be wrapped around a method from an struct type that
// implements its custom authentication and be able to use it with
// `application` struct.
func isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(ContextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}
