package wrouter

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Route represents one callable route
type Route struct {
	Methods    []string
	Controller interface{}
	RMethod    reflect.Method
	Path       string
}

// Add an HTTP Method to the routes
// Allowed methods: "get", "post", "put", "patch", "head", "trace", "connect", "options"
func (r *Route) AddMethod(method string) {
	r.Methods = append(r.Methods, method)
}

// Router represents one implementation of the http.Handler interface
type Router struct {
	routes      []*Route
	injectors   []Injector
	preRequest  []PreRequestEvent
	postRequest []PostRequestEvent
}

// Create a new Router
func NewRouter() *Router {
	return &Router{}
}

// ServeHTTP satisfies the http.Handler interface, so that this Router can be used as the
// second parameter of http.ListenAndServe
func (r *Router) ServeHTTP(w http.ResponseWriter, h *http.Request) {

	// Execute PreRequestEvents if any
	if len(r.preRequest) != 0 {
		ctx := createPreRequestEventContext(h, w)
		for _, event := range r.preRequest {
			event.Exec(ctx)
		}
	}

	route := r.findRequestRoute(h)
	if route == nil {
		// @TODO: Implement check if an ErrorController exists
		// Define Specifications for ErrorController (ex: StatusNotFoundAction??)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
		return
	}

	r.callRoute(route, w, h)
}

// PrintRoutes will print out all routes to io.Writer
// @TODO: Print alias for *Index* Routes
func (r *Router) PrintRoutes(writer io.Writer) {
	for i, route := range r.routes {
		ms := ""
		for _, me := range route.Methods {
			ms += me + " "
		}
		fmt.Fprintln(writer, "ID: "+strconv.Itoa(i)+"\t"+ms+"\t\t"+route.Path)
	}
}

// AddController will add a new controller to the router
func (r *Router) AddController(controller interface{}) {
	rc := reflect.TypeOf(controller)
	verifyController(rc)

	for i := 0; i < rc.NumMethod(); i++ {
		route := createRouteByMethod(controller, rc, rc.Method(i))
		r.AddRoute(route)
	}
}

func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) AddInjector(in Injector) {
	r.injectors = append(r.injectors, in)
}

func (r *Router) AddPreRequestEvent(ev PreRequestEvent) {
	r.preRequest = append(r.preRequest, ev)
}

func (r *Router) AddPostRequestEvent(ev PostRequestEvent) {
	r.postRequest = append(r.postRequest, ev)
}

func (r *Router) findRequestRoute(h *http.Request) *Route {
	uri := strings.ToLower(strings.Trim(h.URL.RequestURI(), "?&/"))
	uriParts := strings.Split(uri, "/")

	// use "IndexController" if no route is specified
	if uriParts[0] == "" {
		uriParts[0] = "index"
	}

	// append index action to general controller call
	if len(uriParts) < 2 {
		uriParts = append(uriParts, "index")
	}

	// match route
	path := uriParts[0] + "/" + uriParts[1]
	for _, route := range r.routes {
		if route.Path == path {
			// reduce complexity for routes with only one method
			if len(route.Methods) == 1 && h.Method == route.Methods[0] {
				return route
			}

			// iterate methods for routes with multiple methods
			for _, me := range route.Methods {
				if me == h.Method {
					return route
				}
			}
		}
	}

	return nil
}

func (r *Router) callRoute(route *Route, w http.ResponseWriter, h *http.Request) {
	values := make([]reflect.Value, 0)

	ctx := createInjectorContext(h, route, r, w)

	// Call controller method and inject arguents by reflection
	for i := 0; i < route.RMethod.Type.NumIn(); i++ {
		arg := route.RMethod.Type.In(i)
		switch arg.String() {
		case reflect.TypeOf(route.Controller).String():
			values = append(values, reflect.ValueOf(route.Controller))
		case "http.ResponseWriter":
			values = append(values, reflect.ValueOf(w))
		case "*http.Request":
			values = append(values, reflect.ValueOf(h))
		default:
			// Execute dynamic injection by
			values = append(values, reflect.ValueOf(r.inject(arg.String(), ctx)))
		}
	}

	ret := route.RMethod.Func.Call(values)

	// Execute PostRequest events
	if len(r.postRequest) != 0 {
		ctx := createPostRequestEventContext(h, w, ret)
		for _, event := range r.postRequest {
			event.Exec(ctx)
		}
	}
}

func (r *Router) inject(t string, ctx *InjectorContext) interface{} {
	if len(r.injectors) != 0 {
		for _, injector := range r.injectors {
			if injector.Supports(t) {
				return injector.Get(ctx)
			}
		}
	}

	return nil
}

/********
 *
 *	HELPER FUNCTIONS
 *
 ********/
func createRouteByMethod(controller interface{}, rc reflect.Type, method reflect.Method) *Route {
	r := new(Route)
	methodName := strings.ToLower(method.Name)
	r.Controller = controller
	r.RMethod = method

	// Add HTTP methods
	if strings.Contains(methodName, "_") {
		var allowed []string = []string{"get", "post", "put", "patch", "head", "trace",
			"connect", "options"}
		methodParts := strings.Split(methodName, "_")
		methodString := methodParts[0]
		methodName = methodParts[1]
		for _, method := range allowed {
			if strings.Contains(methodString, method) {
				r.AddMethod(strings.ToUpper(method))
			}
		}
	}
	if len(r.Methods) == 0 {
		r.AddMethod("GET")
	}

	// Calculate path
	if strings.Contains(methodName, "action") {
		methodName = strings.Replace(methodName, "action", "", -1)
	}
	r.Path = strings.Replace(strings.ToLower(rc.Elem().Name()), "controller", "", -1) + "/" + methodName
	return r
}

func verifyController(rc reflect.Type) {
	if rc.Kind() != reflect.Ptr || rc.Elem().Kind() != reflect.Struct {
		panic(rc.String() + " is a " + rc.Kind().String() + ", pointer to Struct expected")
	}

	if !strings.Contains(rc.String(), "Controller") {
		panic(rc.String() + " does not end in Controller")
	}
}
