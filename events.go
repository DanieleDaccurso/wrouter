package wrouter

import (
	"net/http"
	"reflect"
)

type PreRequestEventContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

type PreRequestEvent interface {
	Exec(*PreRequestEventContext)
}

type PostRequestEventContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Values         []reflect.Value
}

type PostRequestEvent interface {
	Exec(*PostRequestEventContext)
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
