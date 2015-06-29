package sapi

import (
	"github.com/gorilla/mux"
	"net/http"
	"testing"
)

func foo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("foo"))
}

var newRequestHost = func(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func TestMakeEndpoints(t *testing.T) {

	Regist(MakeApiEndpoints("/foo", "GET", http.HandlerFunc(foo)))
	req := newRequestHost("GET", "http://example.com/foo")

	var match mux.RouteMatch
	ok := gRouter.Path("/foo").Match(req, &match)
	if !ok {
		t.Errorf("Expected url match http://example.com/foo, but false")
	}

	Regist(MakeApiEndpoints("/bar", "POST", http.HandlerFunc(foo)))
	req = newRequestHost("GET", "http://example.com/bar0")
	ok = gRouter.Path("/bar").Match(req, &match)
	if ok {
		t.Errorf("Unxpected url match http://example.com/bar, but true")
	}
}

type Foo struct {
	s []byte
}

func (f *Foo) Get(w http.ResponseWriter, r *http.Request) {
	w.Write(f.s)
}

func (f *Foo) Mapper() ApiEndpoints {
	e := make(ApiEndpoints)

	e["GET"] = map[string]http.Handler{"/foo": http.HandlerFunc(f.Get)}

	return e
}

func TestApiInterface(t *testing.T) {
	Regist(new(Foo))
	req := newRequestHost("GET", "http://example.com/foo")

	var match mux.RouteMatch
	ok := gRouter.Path("/foo").Match(req, &match)
	if !ok {
		t.Errorf("Expected url match http://example.com/foo, but false")
	}
}
