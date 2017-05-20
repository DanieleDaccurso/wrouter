package wrouter

import (
	"net/http"
	"reflect"
)

// PreRequestEventContext contains the event context for events which are fired after the ServeHTTP method is called.
// At this point, nothing about the route and the further life-cycle of the request is known to the router yet.
// If values inside the context are manipulated, the manipulated version will be used in the further lifecycle of this
// request.
type PreRequestEventContext struct {
	// Request contains the current request instance
	Request *http.Request
	// ResponseWriter contains the current ResponseWriter instance
	ResponseWriter http.ResponseWriter
}

// PostRouteResolveEventContext contains the event context for events which are fired after the route is determined.
// If values inside the context are manipulated, the manipulated version will be used in the further lifecycle of this
// request.
type PostRouteResolveEventContext struct {
	// Request contains the current request instance
	Request *http.Request
	// ResponseWriter contains the current ResponseWriter instance
	ResponseWriter http.ResponseWriter
	// Route contains the currently resolved route instance. Please note that manipulating this, will result
	// into using another route for the rest of the lifecycle.
	Route *Route
}

// PostRouteResolveEventContext contains the event context for events which are fired after the controller action has
// been called.
// If values inside the context are manipulated, the manipulated version will be used in the further lifecycle of this
// request.
type PostRequestEventContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Values         []reflect.Value
}

type PreRequestEvent interface {
	Exec(*PreRequestEventContext)
}

type PostRequestEvent interface {
	Exec(*PostRequestEventContext)
}

type PostRouteResolveEvent interface {
	Exec(*PostRouteResolveEventContext)
}

func createPreRequestEventContext(h *http.Request, w http.ResponseWriter) *PreRequestEventContext {
	return &PreRequestEventContext{
		Request:        h,
		ResponseWriter: w,
	}
}

func createPostRequestEventContext(h *http.Request, w http.ResponseWriter, vs []reflect.Value) *PostRequestEventContext {
	return &PostRequestEventContext{
		Request:        h,
		ResponseWriter: w,
		Values:         vs,
	}
}

func createPostRouteResolveEventContext(h *http.Request, w http.ResponseWriter, r *Route) *PostRouteResolveEventContext {
	return &PostRouteResolveEventContext{
		Request:        h,
		ResponseWriter: w,
		Route:          r,
	}
}
