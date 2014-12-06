package recovery

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bar"))
	})
	myPanicHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("this did not work")
	})
)

func TestNoConfigGood(t *testing.T) {
	r := New()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/should/not/panic/", nil)
	r.Handler(myHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Body.String(), "bar")
}

func TestDefaultPanic(t *testing.T) {
	r := New(Options{
		Out: ioutil.Discard,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/should/panic/", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)
	expect(t, strings.TrimSpace(res.Body.String()), strings.TrimSpace(http.StatusText(http.StatusInternalServerError)))
}

func TestCustomPanicHandler(t *testing.T) {
	r := New(Options{
		Out: ioutil.Discard,
	})
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("You got 400 yo!"))
	})
	r.SetPanicHandler(customHandler)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/should/400/", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusBadRequest)
	expect(t, res.Body.String(), "You got 400 yo!")
}

func TestDefaultConfig(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		Out: buf,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	// Expect some debug info.
	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")

	// LstdFlags output.
	curDate := time.Now().Format("2006/01/02 15:04")
	expectContainsTrue(t, buf.String(), curDate)
}

func TestStackSize(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		Out:       buf,
		StackSize: 50,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	// Expect some debug info.
	expectContainsTrue(t, buf.String(), "Recovering from Panic:")

	// This will be too far down the stack results.
	expectContainsFalse(t, buf.String(), "src/net/http/server.go")

	// LstdFlags output.
	curDate := time.Now().Format("2006/01/02 15:04")
	expectContainsTrue(t, buf.String(), curDate)
}

func TestIncludeFullStack(t *testing.T) {
	bufNormal := bytes.NewBufferString("")
	bufFull := bytes.NewBufferString("")

	r := New(Options{
		Out:              bufNormal,
		IncludeFullStack: false,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	r = New(Options{
		Out:              bufFull,
		IncludeFullStack: true,
	})

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	// Full stack should be larger than non.
	if bufNormal.Len() >= bufFull.Len() {
		t.Errorf("Full stack should be larger than non. %d | %d", bufNormal.Len(), bufFull.Len())
	}
}

func TestCustomPrefix(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		Prefix: "testapp_-_yo",
		Out:    buf,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")
	expectContainsTrue(t, buf.String(), "[testapp_-_yo] ")
}

func TestCustomPrefixWithNoBrackets(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		Prefix:              "testapp_-_yo2()",
		DisableAutoBrackets: true,
		Out:                 buf,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")
	expectContainsTrue(t, buf.String(), "testapp_-_yo2()")
	expectContainsFalse(t, buf.String(), "[testapp_-_yo2()] ")
}

func TestCustomFlags(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		OutputFlags: log.Lshortfile,
		Out:         buf,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	// Log should start with...
	expect(t, buf.String()[0:12], "recovery.go:")
	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")

	// Should not include a date now.
	curDate := time.Now().Format("2006/01/02")
	expectContainsFalse(t, buf.String(), curDate)
}

func TestCustomFlagsZero(t *testing.T) {
	buf := bytes.NewBufferString("")

	r := New(Options{
		OutputFlags: -1,
		Out:         buf,
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")

	// Should not include a date now.
	curDate := time.Now().Format("2006/01/02")
	expectContainsFalse(t, buf.String(), curDate)
}

func TestIgnoreMultipleConfigs(t *testing.T) {
	buf := bytes.NewBufferString("")

	opt1 := Options{Out: buf}
	opt2 := Options{Out: os.Stderr, OutputFlags: -1}

	r := New(opt1, opt2)

	res := httptest.NewRecorder()
	url := "/should/output/to/buf/only/"
	req, _ := http.NewRequest("GET", url, nil)
	req.RequestURI = url
	r.Handler(myPanicHandler).ServeHTTP(res, req)

	expect(t, res.Code, http.StatusInternalServerError)

	expectContainsTrue(t, buf.String(), "Recovering from Panic:")
	expectContainsTrue(t, buf.String(), "src/net/http/server.go")

	// LstdFlags output.
	curDate := time.Now().Format("2006/01/02 15:04")
	expectContainsTrue(t, buf.String(), curDate)
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func expectContainsTrue(t *testing.T, a, b string) {
	if !strings.Contains(a, b) {
		t.Errorf("Expected [%s] to contain [%s]", a, b)
	}
}

func expectContainsFalse(t *testing.T, a, b string) {
	if strings.Contains(a, b) {
		t.Errorf("Expected [%s] to contain [%s]", a, b)
	}
}
