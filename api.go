package sapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	gRouter = mux.NewRouter()
	Methods = []string{GET, DELETE, POST, UPDATE}
)

type ApiEndpoints map[string]map[string]http.Handler

type Apier interface {
	Mapper() ApiEndpoints
}

type ApiMapperFunc func() ApiEndpoints

func (f ApiMapperFunc) Mapper() ApiEndpoints {
	return f()
}

func Regist(apiers ...Apier) {
	for _, apier := range apiers {
		mapping := apier.Mapper()
		for method, routeMap := range mapping {
			for route, handler := range routeMap {
				gRouter.Handle(route, handler).Methods(method)
			}
		}
	}
}

func MakeApiEndpoints(method, route string, f http.Handler) ApiMapperFunc {
	return func() ApiEndpoints {
		endpoints := make(ApiEndpoints)
		endpoints[method] = map[string]http.Handler{route: f}
		return endpoints
	}
}

func Router() *mux.Router {
	return gRouter
}

func HttpError(rw http.ResponseWriter, err string, e error, code int) {
	http.Error(rw, err, code)
	if e != nil {
		Log().Error(fmt.Sprintf("%s", e))
	}
}
