package wrouter

import (
	"net/http"
	"reflect"
)


// Injector example. Wrapper to directly write strings to the http.ResponseWriter
// instead of using byte-arrays. This is not injected by default.
// if you want to inject the StringResponseWriter, use the AddInjector method of your
// router instance
type StringResponseWriterInjector struct {}

func (s *StringResponseWriterInjector) Supports(t string) bool {
	v := reflect.TypeOf(s)
	println(v.String())
	return t == "*wrouter.StringResponseWriter"
}

func (s *StringResponseWriterInjector) Get(ctx *InjectorContext) interface{} {
	return &StringResponseWriter{
		writer: ctx.ResponseWriter,
	}
}


// StringResponseWriter is a wrapper for http.ResponseWriter, which writes strings
// instead of []byte
type StringResponseWriter struct {
	writer http.ResponseWriter
}

// NewStringResponseWriter creates a new StringResponseWriter and passes the underlying
// http.ResponseWriter as a constructor injection
func NewStringResponseWriter(w http.ResponseWriter) *StringResponseWriter {
	return &StringResponseWriter{
		writer: w,
	}
}

// Header just calls the http.ResponseWriter's Header method
func (w *StringResponseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write converts the input from string to []byte and passes it to the http.ResponseWriter
func (w *StringResponseWriter) Write(str string) (int, error) {
	return w.writer.Write([]byte(str))
}

// WriteHeader just calls the http.ResponseWriter's WriteHeader method
func (w *StringResponseWriter) WriteHeader(h int) {
	w.writer.WriteHeader(h)
}

