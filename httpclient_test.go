package sapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	var (
		case1 = "/"
		case2 = "/withHeader"
	)

	tserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != GET {
			t.Errorf("Excpected method %q, got %q", GET, r.Method)
		}

		if r.Header == nil {
			t.Errorf("Excpected non-nil Header")
		}

		switch r.URL.Path {
		case case1:
			t.Logf("Case %v", case1)
		case case2:
			t.Logf("Case %v", case2)
			if r.Header.Get("key1") != "value1" {
				t.Errorf("Excpected Header %s with value %s, got %s", "key1", "value1", r.Header.Get("key1"))
			}
		default:
			t.Logf("No test case for %s", r.URL.Path)
		}
	}))
	defer tserver.Close()

	NewHttpAgent().Get(tserver.URL + case1).Issue()
	NewHttpAgent().Get(tserver.URL+case2).
		SetHeader("key1", "value1").
		Issue()
}
