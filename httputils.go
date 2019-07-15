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
	GetErrorLog() *log.Logger
	GetSession() *sessions.Session
	GetTemplateCache(string) (*template.Template, error)
	IsAuthenticated(*http.Request) bool
}

// TODO: Refactor this to more specific behaviors so that
// 	there is no need to implement the whole interface.
type templateData interface {
	EnableCSRFToken() bool
	EnableCurrentYear() bool
	EnableAuthentication() bool
	GetTemplateData() interface{}
	GetCSRFToken() string
	GetCurrentYear() int
	GetIsAuthenticated() bool
	SetCSRFToken(string)
	SetCurrentYear(int)
	SetIsAuthenticated(bool)
}

func ServerError(errorLog *log.Logger, w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func NotFound(w http.ResponseWriter) {
	ClientError(w, http.StatusNotFound)
}

func AddDefaultData(app application, td templateData, r *http.Request) {
	if td.EnableCSRFToken() {
		td.SetCSRFToken(nosurf.Token(r))
	}
	if td.EnableCurrentYear() {
		td.SetCurrentYear(time.Now().Year())
	}
	if td.EnableAuthentication() {
		td.SetIsAuthenticated(app.IsAuthenticated(r))
	}
}

func Render(app application, td templateData, w http.ResponseWriter, r *http.Request, name string) {
	ts, err := app.GetTemplateCache(name)
	if err != nil {
		ServerError(app.GetErrorLog(), w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)

	err = ts.Execute(buf, td.GetTemplateData())
	if err != nil {
		fmt.Println("There was an error")
		ServerError(app.GetErrorLog(), w, err)
	}

	buf.WriteTo(w)
}

// Convenient function to check if request is authenticated with context.
// Method can be wrapped around a method from an struct type that
// implements its custom authentication and be able to use it with
// `application` struct.
func IsAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(ContextKeyIsAuthenticated).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}
