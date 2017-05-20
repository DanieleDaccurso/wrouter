package wrouter

import (
	"errors"
	"fmt"
	"github.com/owtorg/events"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Route represents one callable route
type Route struct {
	// Methods contains a slice of strings, identifying the allowed HTTP methods for the current route
	Methods []string
	// Controller contains the current Controller
	Controller Controller
	// RMethod contains the reflect.Method of the Controller's action to be called on the current route
	RMethod reflect.Method
	// Path contains the path for the current route
	Path string
}

// EventPriorityError is returned when an Event's priority is already taken by another event.
// Every event priority may be only given once.
var EventPriorityError = errors.New("Event with given priority already exists")

var AllowedMethods = []string{"get", "post", "put", "patch", "head", "trace", "connect", "options", "delete"}

// Add an HTTP Method to the routes
// Allowed methods: "get", "post", "put", "patch", "head", "trace", "connect", "options", "delete"
func (r *Route) AddMethod(method string) {
	r.Methods = append(r.Methods, method)
}

// Router represents one implementation of the http.Handler interface
type Router struct {
	routes    []*Route
	injectors []Injector

	// The event-dispatcher requires a map[int]interface{} in order to act generically.
	// Please do not modify this to the exact types. Also do not attach whatever interface{}
	// to those events. Use the Adder methods for both request maps for type validation
	preRequest  map[int]events.Event
	postRequest map[int]events.Event

	// Configuration contains the router configuration
	Configuration *Configuration
	Resolver      RouteResolver
}

// Create a new Router
func NewRouter() *Router {
	r := new(Router)
	r.preRequest = make(map[int]events.Event)
	r.postRequest = make(map[int]events.Event)
	r.Configuration = createDefaultConfiguration()
	r.Resolver = newRouteResolver(r.Configuration)
	return r
}

// ServeHTTP satisfies the http.Handler interface, so that this Router can be used as the
// second parameter of http.ListenAndServe
func (r *Router) ServeHTTP(w http.ResponseWriter, h *http.Request) {
	// Execute PreRequestEvents if any
	if len(r.preRequest) != 0 {
		ctx := createPreRequestEventContext(h, w)
		events.DispatchEvents(r.preRequest, ctx)
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
func (r *Router) PrintRoutes(writer io.Writer) {
	for i, route := range r.routes {
		ms := ""
		for _, me := range route.Methods {
			ms += me + " "
		}
		fmt.Fprintln(writer, "ID: "+strconv.Itoa(i)+"\t"+ms+"\t\t"+route.Path)
	}
}

// AddController will add a new controller to the router.
func (r *Router) AddController(controller interface{}) {
	routes, err := r.Resolver.Resolve(controller)
	if err != nil {
		panic(err)
	}
	if len(routes) > 0 {
		for i := 0; i < len(routes); i++ {
			r.AddRoute(routes[i])
		}
	}
}

// AddRoute will add a new Route to a controller.
func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

// AddInjector will append an Injector to the list of Injectors in this router.
// It will automatically be the last executed injector as the first injector to support a certain
// type, will be the one who serves the value.
func (r *Router) AddInjector(in Injector) {
	r.injectors = append(r.injectors, in)
}

// AddPreRequestEvent will add a PreRequestEvent with a given priority. It will return an error
// if the selected priority is already taken.
func (r *Router) AddPreRequestEvent(ev PreRequestEvent, priority int) error {
	_, exists := r.preRequest[priority]
	if exists {
		return EventPriorityError
	}
	r.preRequest[priority] = ev
	return nil
}

// AddPostRequestEvent will add a PostRequestEvent with a given priority. It will return an error
// if the selected priority is already taken.
func (r *Router) AddPostRequestEvent(ev PostRequestEvent, priority int) error {
	_, exists := r.postRequest[priority]
	if exists {
		return EventPriorityError
	}
	r.postRequest[priority] = ev
	return nil
}

// AppendPreRequestEvent will append a PreRequestEvent at the end of the current PreRequestEvent
// queue. If you want to set a priority, see AddPreRequestEvent
func (r *Router) AppendPreRequestEvent(ev PreRequestEvent) {
	key := getHighestKey(r.preRequest) + 1
	r.preRequest[key] = ev
}

// AppendPostRequestEvent will append a PreRequestEvent at the end of the current AddPostRequestEvent
// queue. If you want to set a priority, see AddPostRequestEvent
func (r *Router) AppendPostRequestEvent(ev PostRequestEvent) {
	key := getHighestKey(r.postRequest) + 1
	r.postRequest[key] = ev
}

func (r *Router) findRequestRoute(h *http.Request) *Route {
	uri := strings.ToLower(strings.Trim(h.URL.RequestURI(), "?&/"))
	uriParts := strings.Split(uri, "/")

	// required full-path to match subroutes, without strict slashes
	fullPath := strings.Trim(strings.Split(uri, "?")[0], "/")

	path := uriParts[0]
	if len(uriParts) > 1 {
		// match route
		path = uriParts[0] + "/" + uriParts[1]
	}

	for _, route := range r.routes {
		if route.Path == path || route.Path == fullPath {
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

	// Currently, every controller action needs to be part of a struct, therefore
	// the first argument of the method, is the struct itself. This happens implicit
	// when a method is defined as func (s *struct) doit()
	values = append(values, reflect.ValueOf(route.Controller))

	// argument resolving switch is only called, if a method has more than one argument
	if route.RMethod.Type.NumIn() > 1 {
		// Call controller method and inject arguents by reflection
		for i := 1; i < route.RMethod.Type.NumIn(); i++ {
			arg := route.RMethod.Type.In(i)
			switch arg.String() {
			case "http.ResponseWriter":
				values = append(values, reflect.ValueOf(w))
			case "*http.Request":
				values = append(values, reflect.ValueOf(h))
			default:
				values = append(values, reflect.ValueOf(r.inject(arg.String(), ctx)))
			}
		}
	}

	ret := route.RMethod.Func.Call(values)

	// Execute PostRequest events
	if len(r.postRequest) != 0 {
		ctx := createPostRequestEventContext(h, w, ret)
		events.DispatchEvents(r.postRequest, ctx)
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
