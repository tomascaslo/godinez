package godinez

import (
	"errors"
	"github.com/golangcollege/sessions"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestServerError(t *testing.T) {
	errorLog := log.New(ioutil.Discard, "", 0)
	rr := httptest.NewRecorder()
	err := errors.New("Error")

	ServerError(errorLog, rr, err)

	expectedStatus := http.StatusInternalServerError
	if rr.Code != expectedStatus {
		t.Errorf("Expected %q got %q", rr.Body, expectedStatus)
	}

	actualStatusText := strings.TrimSuffix(rr.Body.String(), "\n")
	expectedStatusText := http.StatusText(http.StatusInternalServerError)
	if actualStatusText != expectedStatusText {
		t.Errorf("Expected %q got %q", expectedStatusText, actualStatusText)
	}
}

func TestClientError(t *testing.T) {
	rr := httptest.NewRecorder()
	expectedStatus := http.StatusBadRequest

	ClientError(rr, expectedStatus)

	if rr.Code != expectedStatus {
		t.Errorf("Expected %q got %q", rr.Body, expectedStatus)
	}

	actualStatusText := strings.TrimSuffix(rr.Body.String(), "\n")
	expectedStatusText := http.StatusText(expectedStatus)
	if actualStatusText != expectedStatusText {
		t.Errorf("Expected %q got %q", expectedStatusText, actualStatusText)
	}
}

func TestNotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	expectedStatus := http.StatusNotFound

	ClientError(rr, expectedStatus)

	if rr.Code != expectedStatus {
		t.Errorf("Expected %q got %q", rr.Body, expectedStatus)
	}

	actualStatusText := strings.TrimSuffix(rr.Body.String(), "\n")
	expectedStatusText := http.StatusText(expectedStatus)
	if actualStatusText != expectedStatusText {
		t.Errorf("Expected %q got %q", expectedStatusText, actualStatusText)
	}
}

type addDefaultDataSpy struct {
	calls []string
}

func (spy *addDefaultDataSpy) addCall(funcName string) {
	spy.calls = append(spy.calls, funcName)
}

type mockApplication struct {
	spy              *addDefaultDataSpy
	funcReturnValues map[string]interface{}
}

func (ma *mockApplication) checkAndAddCall(funcName string) {
	ma.spy.addCall(funcName)
}

func (ma *mockApplication) GetErrorLogger() *log.Logger {
	ma.checkAndAddCall("getErrorLog")

	if errorLog, ok := ma.funcReturnValues["getErrorLog"].(*log.Logger); ok {
		return errorLog
	}

	return nil
}

func (ma *mockApplication) GetSession() *sessions.Session {
	ma.checkAndAddCall("getSession")
	return nil
}

func (ma *mockApplication) GetTemplateCache(string) (*template.Template, error) {
	ma.checkAndAddCall("getTemplateCache")

	if templ, ok := ma.funcReturnValues["getTemplateCache"].(*template.Template); ok {
		return templ, nil
	}

	// In case nil is set
	_, ok := ma.funcReturnValues["getTemplateCache"]
	if ok {
		return nil, errors.New("no template found")
	}

	return nil, nil
}

func (ma *mockApplication) IsAuthenticated(*http.Request) bool {
	ma.checkAndAddCall("isAuthenticated")
	return false
}

type mockTemplateData struct {
	spy              *addDefaultDataSpy
	funcReturnValues map[string]interface{}
}

func (mtd *mockTemplateData) checkAndAddCall(funcName string) {
	mtd.spy.addCall(funcName)
}

func (mtd *mockTemplateData) EnableCSRFToken() bool {
	mtd.checkAndAddCall("enableCSRFToken")

	if returnValue, ok := mtd.funcReturnValues["enableCSRFToken"].(bool); ok {
		return returnValue
	}

	return false
}

func (mtd *mockTemplateData) EnableCurrentYear() bool {
	mtd.checkAndAddCall("enableCurrentYear")

	if returnValue, ok := mtd.funcReturnValues["enableCurrentYear"].(bool); ok {
		return returnValue
	}

	return false
}

func (mtd *mockTemplateData) EnableAuthentication() bool {
	mtd.checkAndAddCall("enableAuthentication")

	if returnValue, ok := mtd.funcReturnValues["enableAuthentication"].(bool); ok {
		return returnValue
	}

	return false
}

func (mtd *mockTemplateData) GetTemplateData() interface{} {
	mtd.checkAndAddCall("getTemplateData")

	if templData, ok := mtd.funcReturnValues["getTemplateData"].(struct{ Data string }); ok {
		return templData
	}

	return nil
}

func (mtd *mockTemplateData) GetCSRFToken() string {
	mtd.checkAndAddCall("getCSRFToken")
	return ""
}

func (mtd *mockTemplateData) GetCurrentYear() int {
	mtd.checkAndAddCall("getCurrentYear")
	return 0
}

func (mtd *mockTemplateData) GetIsAuthenticated() bool {
	mtd.checkAndAddCall("getIsAuthenticated")
	return false
}

func (mtd *mockTemplateData) SetCSRFToken(string) {
	mtd.checkAndAddCall("setCSRFToken")
}

func (mtd *mockTemplateData) SetCurrentYear(int) {
	mtd.checkAndAddCall("setCurrentYear")
}

func (mtd *mockTemplateData) SetIsAuthenticated(bool) {
	mtd.checkAndAddCall("setIsAuthenticated")
}

func TestAddDefaultData(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name          string
		app           *mockApplication
		td            *mockTemplateData
		req           *http.Request
		expectedCalls []string
	}{
		{
			"CSRFToken enabled",
			&mockApplication{},
			&mockTemplateData{funcReturnValues: map[string]interface{}{"enableCSRFToken": true}},
			r,
			[]string{"enableCSRFToken", "setCSRFToken", "enableCurrentYear", "enableAuthentication"},
		},
		{
			"CurrentYear enabled",
			&mockApplication{},
			&mockTemplateData{funcReturnValues: map[string]interface{}{"enableCurrentYear": true}},
			r,
			[]string{"enableCSRFToken", "enableCurrentYear", "setCurrentYear", "enableAuthentication"},
		},
		{
			"Authentication enabled",
			&mockApplication{},
			&mockTemplateData{funcReturnValues: map[string]interface{}{"enableAuthentication": true}},
			r,
			[]string{"enableCSRFToken", "enableCurrentYear", "enableAuthentication", "isAuthenticated", "setIsAuthenticated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup spy
			spy := &addDefaultDataSpy{[]string{}}
			tt.app.spy = spy
			tt.td.spy = spy

			AddDefaultData(tt.app, tt.td, tt.req)

			if !reflect.DeepEqual(spy.calls, tt.expectedCalls) {
				t.Errorf("Expected calls %v got %v", tt.expectedCalls, spy.calls)
			}
		})
	}
}

func TestRender(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	templ := template.New("index")
	templ = template.Must(templ.Parse(`<html><head></head><body>{{.Data}}</body>`))
	tests := []struct {
		name      string
		app       *mockApplication
		td        *mockTemplateData
		rr        *httptest.ResponseRecorder
		req       *http.Request
		templName string
		expected  string
	}{
		{
			"Render sucess",
			&mockApplication{funcReturnValues: map[string]interface{}{"getTemplateCache": templ}},
			&mockTemplateData{funcReturnValues: map[string]interface{}{"getTemplateData": struct{ Data string }{"Hello World!"}}},
			httptest.NewRecorder(),
			req,
			"",
			`<html><head></head><body>Hello World!</body>`,
		},
		{
			"Render with template data nil",
			&mockApplication{funcReturnValues: map[string]interface{}{"getTemplateCache": templ}},
			&mockTemplateData{funcReturnValues: map[string]interface{}{"getTemplateData": nil}},
			httptest.NewRecorder(),
			req,
			"",
			`<html><head></head><body></body>`,
		},
		{
			"Template error",
			&mockApplication{funcReturnValues: map[string]interface{}{"getTemplateCache": nil, "getErrorLog": log.New(ioutil.Discard, "", 0)}},
			&mockTemplateData{},
			httptest.NewRecorder(),
			req,
			"index",
			http.StatusText(http.StatusInternalServerError),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup spy
			spy := &addDefaultDataSpy{[]string{}}
			tt.app.spy = spy
			tt.td.spy = spy

			Render(tt.app, tt.td, tt.rr, tt.req, tt.templName)

			actualStatusText := strings.TrimSuffix(tt.rr.Body.String(), "\n")
			if actualStatusText != tt.expected {
				t.Errorf("Expected %q got %q", tt.expected, actualStatusText)
			}
		})
	}
}
