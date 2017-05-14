package wrouter

import (
	"net/http"
	"reflect"
)

// Event represents one callable event
type EventContainer struct {
	Controller interface{}
	RMethod    reflect.Method
}

type Context struct {
	w http.ResponseWriter
	h *http.Request
	//Storage is arbitrary space to store things that should last the lifetime of this call and be accessible in events and routes
	storage map[string]interface{}
	//previously results would be stored in their own var, now we add to the context.
	//This allows pre hooks to also add result values if necessary
	//Such as adding some generic information which may be rendered on every call
	results []reflect.Value
}

func NewContext(w http.ResponseWriter, h *http.Request) *Context {
	storage := new(map[string]interface{})
	return &Context{w: w, h: h, storage: *storage}
}
