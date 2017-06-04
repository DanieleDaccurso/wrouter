package wrouter

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

// Controller is an empty interface, to make the use of interface{} in regard of Controller handling unnecessary
type Controller interface{}

// ReflectController is an alias for reflect.Type, used in context of the Controller's reflection instance
type ReflectController reflect.Type

// WrongTypeError indicates that the controller cannot be resolved, because it's not a pointer to a struct
var WrongTypeError = errors.New("A constroller must be a pointer to a struct")

// WrongNameError indicates that the controller cannot be resolved, because it doesn't end in Controller
var WrongNameError = errors.New("A constroller struct name must end in \"Controller\"")

// cleanSlahes is a regex used to replace multiple slashes with one slash in a path. It is pre-compiled
// for performance reasons.
var cleanSlashes, _ = regexp.Compile(`\/+`)

// RouteResolver is a tiny interface for route resolving. It can be used to switch to a custom route resolver
// See: Route.Resolver
type RouteResolver interface {
	Resolve(Controller) ([]*Route, error)
}

type ctrRouteResolver struct {
	configuration *Configuration
}

func newRouteResolver(cfg *Configuration) *ctrRouteResolver {
	rs := new(ctrRouteResolver)
	rs.configuration = cfg
	return rs
}

// Resolve implements the RouteResolver interface
func (rs *ctrRouteResolver) Resolve(controller Controller) ([]*Route, error) {
	rct := rs.createControllerReflection(controller)
	routes := rs.getRoutes(controller, rct, "")
	return routes, nil
}

// createReflection will create a reflect.Type by an interface{}
func (rs *ctrRouteResolver) createControllerReflection(controller Controller) ReflectController {
	return reflect.TypeOf(controller)
}

// verifyController will take a reflection type and verify if it is eligible to be used as a
// controller.
func (rs *ctrRouteResolver) verifyController(rct ReflectController) error {
	// Assert that the given controller is actually a pointer to a struct
	if rct.Kind() != reflect.Ptr || rct.Elem().Kind() != reflect.Struct {
		return WrongNameError
	}

	// Assert that the given controller name ends in Controller
	if !strings.Contains(rct.String(), "Controller") {
		return WrongTypeError
	}

	return nil
}

func (rs *ctrRouteResolver) getRoutes(controller Controller, rct ReflectController, prefix string) []*Route {
	// verification of the controller is needed. If the verification doesn't pass, the execution
	// will terminate and the application will panic. The reason for such strict handling is,
	// that such error usually occurs when building the application and not later on runtime.
	verErr := rs.verifyController(rct)
	if verErr != nil {
		panic(verErr)
	}

	var routes = []*Route{}

	// Iterate over all methods of the controller and create routes for it
	for i := 0; i < rct.NumMethod(); i++ {
		// Create routes for the given method. This will typically be one route by the
		// naming convention, and possibly an index route.
		rt := rs.createRoutesByMethod(controller, rct.Method(i), rct, prefix)
		for i := 0; i < len(rt); i++ {
			routes = append(routes, rt[i])
		}
	}

	// Add sub-controllers if allowed by configuration
	if rs.configuration.AllowSubController {
		subControllers := rs.getSubControllers(rct)
		if len(subControllers) != 0 {

			// Create a new prefix for the sub-controller. Because of the recursive nature of the procedure
			// this is typically the current prefix, and the parent controller path.
			newPrefix := prefix + controllerPath(controller) + "/"
			for _, subController := range subControllers {
				// Get subroutes
				sRoutes := rs.getRoutes(subController, rs.createControllerReflection(subController),
					newPrefix)

				// append subroutes to route collection
				if len(sRoutes) != 0 {
					for j := 0; j < len(sRoutes); j++ {
						routes = append(routes, sRoutes[j])
					}
				}
			}
		}
	}

	return routes
}

func (rs *ctrRouteResolver) createRoutesByMethod(controller Controller, rfm reflect.Method,
	rct ReflectController, prefix string) []*Route {
	// create main route
	route := new(Route)
	route.Controller = controller
	route.RMethod = rfm

	methodName := strings.ToLower(rfm.Name)

	// extract HTTP methods
	if strings.Contains(methodName, "_") {
		methodParts := strings.Split(methodName, "_")
		methodsString := methodParts[0]
		methodName = methodParts[1]
		for _, method := range AllowedMethods {
			if strings.Contains(methodsString, method) {
				route.AddMethod(strings.ToUpper(method))
			}
		}
	}

	if len(route.Methods) == 0 {
		route.AddMethod("GET")
	}

	if strings.Contains(methodName, "action") {
		methodName = strings.Replace(methodName, "action", "", -1)
	}

	route.Path = prefix + strings.Replace(strings.ToLower(rct.Elem().Name()), "controller", "", -1) +
		"/" + methodName
	routes := []*Route{route}
	if rs.configuration.CreateAliasRoutes && strings.Contains(route.Path, "index") {
		aliasRoute := new(Route)
		aliasRoute.Methods = route.Methods
		aliasRoute.Controller = route.Controller
		aliasRoute.RMethod = route.RMethod

		newPath := strings.Replace(route.Path, "index", "", -1)
		aliasRoute.Path = strings.Trim(cleanSlashes.ReplaceAllString(newPath, "/"), "/")
		routes = append(routes, aliasRoute)
	}

	return routes
}

func (rs *ctrRouteResolver) getSubControllers(rct ReflectController) []Controller {
	var controllers []Controller
	nFn := rct.Elem().NumField()
	if nFn == 0 {
		return controllers
	}

	for i := 0; i < nFn; i++ {
		subrt := rct.Elem().Field(i).Type
		err := rs.verifyController(subrt)
		if err == nil {
			r := reflect.New(subrt.Elem())
			controllers = append(controllers, r.Interface())
		}
	}
	return controllers
}
