package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	SecureHeaders(next).ServeHTTP(rr, req)

	rs := rr.Result()
	defer rs.Body.Close()

	frameOptions := rs.Header.Get("X-Frame-Options")
	if frameOptions != "deny" {
		t.Errorf("Expected %q got %q", "deny", frameOptions)
	}

	xssProtection := rs.Header.Get("X-XSS-Protection")
	if xssProtection != "1;mode=block" {
		t.Errorf("Expected %q got %q", "1;mode=block", xssProtection)
	}

	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello" {
		t.Errorf("Expected %q got %q", "Hello", string(body))
	}
}

type mockLogHolder struct{
	infoLog *log.Logger
	errorLog *log.Logger
}

func (mlh *mockLogHolder) getInfoLogger() *log.Logger {
	return mlh.infoLog
}

func (mlh *mockLogHolder) getErrorLogger() *log.Logger {
	return mlh.errorLog
}

type mockApplicationAuthenticator struct{}

func (maa *mockApplicationAuthenticator) isAuthenticated(r *http.Request) bool {
	return false
}

func TestLogRequest(t *testing.T) {
	logBuf := new(bytes.Buffer)
	infoLog := log.New(logBuf, "", 0)
	logHolder := &mockLogHolder{infoLog, nil}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	logRequestMiddleware := LogRequest(logHolder)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	logRequestHandler := logRequestMiddleware(next)
	logRequestHandler.ServeHTTP(rr, req)

	actual := strings.Trim(logBuf.String(), "\n")
	expected := fmt.Sprintf("%s - %s %s %s", req.RemoteAddr, req.Proto, req.Method, req.URL.RequestURI())
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}

func TestRecoverPanic(t *testing.T) {
	logBuf := new(bytes.Buffer)
	errorLog := log.New(logBuf, "", 0)
	logHolder := &mockLogHolder{nil, errorLog}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	recoverPanicMiddleware := RecoverPanic(logHolder)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("panic"))
	})

	recoverPanicHandler := recoverPanicMiddleware(next)
	recoverPanicHandler.ServeHTTP(rr, req)

	rs := rr.Result()

	actual := rs.Header.Get("Connection")
	expected := "close"
	if actual != expected {
		t.Errorf("Expected %q got %q", expected, actual)
	}

	actual = strings.Trim(logBuf.String(), "\n")
	if !strings.Contains(actual, "panic") {
		t.Errorf("Expected %q got %q", expected, actual)
	}
}
