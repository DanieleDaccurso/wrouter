package wrouter

import "net/http"

type InjectorContext struct {
	Request        *http.Request
	Route          *Route
	Router         *Router
	ResponseWriter http.ResponseWriter
}

func createInjectorContext(request *http.Request, route *Route, router *Router, rwriter http.ResponseWriter) *InjectorContext {
	return &InjectorContext{
		Request:        request,
		Route:          route,
		Router:         router,
		ResponseWriter: rwriter,
	}
}

type Injector interface {
	Supports(string) bool
	Get(*InjectorContext) interface{}
}
