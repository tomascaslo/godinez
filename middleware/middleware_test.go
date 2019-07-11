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

type mockApplicationAuthenticator struct{
	isAuth bool
	redirectTo string
}

func (maa *mockApplicationAuthenticator) isAuthenticated(r *http.Request) bool {
	return maa.isAuth
}

func (maa *mockApplicationAuthenticator) getRedirectTo() string {
	return maa.redirectTo
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

func TestRequireAuthentication(t *testing.T) {
	tests := []struct{
		name string
		rr *httptest.ResponseRecorder
		req *http.Request
		app *mockApplicationAuthenticator
		expectedLocation string
		expectedStatusCode int
	} {
		{
			"Is not authenticated",
			httptest.NewRecorder(),
			httptest.NewRequest("GET", "/", nil),
			&mockApplicationAuthenticator{isAuth: false, redirectTo: "/"},
			"/",
			http.StatusFound,
		},
		{
			"Is authenticated",
			httptest.NewRecorder(),
			httptest.NewRequest("GET", "/secured", nil),
			&mockApplicationAuthenticator{isAuth: true},
			"",
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello"))
			})

			requireAuthenticationMiddleware := RequireAuthentication(tt.app)
			requireAuthenticationHandler := requireAuthenticationMiddleware(next)

			requireAuthenticationHandler.ServeHTTP(tt.rr, tt.req)

			rs := tt.rr.Result()

			if rs.StatusCode != tt.expectedStatusCode {
				t.Errorf("Expected %d got %d", tt.expectedStatusCode, rs.StatusCode)
			}

			actual := rs.Header.Get("Location")
			if actual != tt.expectedLocation {
				t.Errorf("Expected %q got %q", tt.expectedLocation, actual)
			}
		})
	}
}

func TestNoSurf(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	csrfHandler := NoSurf(next)
	csrfHandler.ServeHTTP(rr, req)

	rs := rr.Result()

	cookies := rs.Cookies()

	CSRFTokenCookie := getCookie(cookies, "csrf_token")

	if CSRFTokenCookie == nil {
		t.Error("csrf_token cookie is not set")
	}

	if CSRFTokenCookie.Path != "/" {
		t.Errorf("Expected %q got %q", "/", CSRFTokenCookie.Path)
	}

	if CSRFTokenCookie.HttpOnly != true {
		t.Errorf("Expected %t got %t", true, CSRFTokenCookie.HttpOnly)
	}

	if CSRFTokenCookie.Secure != true {
		t.Errorf("Expected %t got %t", true, CSRFTokenCookie.Secure)
	}
}

func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}


