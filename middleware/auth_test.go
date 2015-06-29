package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegangsta/negroni"
)

func TestAuth(t *testing.T) {
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:admin"))

	h := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("auth"))
	})

	m := negroni.New()
	m.Use(NewBasicAuth())
	m.UseHandler(h)

	r, _ := http.NewRequest("GET", "", nil)
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	if recorder.Code != 401 {
		t.Error("Response not 401")
	}
	respBody := recorder.Body.String()
	if respBody == "auth" {
		t.Error("Auth block failed")
	}
	recorder = httptest.NewRecorder()
	r.Header.Set("Authorization", auth)
	m.ServeHTTP(recorder, r)
	if recorder.Code == 401 {
		t.Error("Response is 401")
	}
	if recorder.Body.String() != "auth" {
		t.Error("Auth failed, got: ", recorder.Body.String())
	}
}
