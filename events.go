package wrouter

import (
	"net/http"
	"reflect"
)

type PreRequestEvent interface {
	Exec(*http.Request, http.ResponseWriter)
}

type PostRequestEvent interface {
	Exec(*http.Request, http.ResponseWriter, []reflect.Value)
}