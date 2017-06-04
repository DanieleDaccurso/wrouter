package wrouter

import (
	"net/http"
	"strings"
)

type routeCache struct {
	pairs map[string]*Route
}

func (r *routeCache) get(uri string) *Route {
	route, exists := r.pairs[uri]
	if exists {
		return route
	}
	return nil
}

func (r *routeCache) push(uri string, rt *Route) {
	if len(r.pairs) == 0 {
		r.pairs = make(map[string]*Route)
	}
	r.pairs[uri] = rt
}

type RequestResolver interface {
	Resolve(*http.Request) *Route
}

type rqResolver struct {
	router *Router
	cache  *routeCache
}

func newRqResolver(router *Router) *rqResolver {
	rr := new(rqResolver)
	rr.router = router
	rr.cache = new(routeCache)
	return rr
}

func (r *rqResolver) Resolve(request *http.Request) *Route {
	uri := strings.ToLower(strings.Trim(request.URL.RequestURI(), "?&/"))
	uriParts := strings.Split(uri, "/")
	fullPath := strings.Trim(strings.Split(uri, "?")[0], "/")
	cachePath := request.Method + "__" + fullPath
	cachedRoute := r.cache.get(cachePath)
	if cachedRoute != nil {
		return cachedRoute
	}

	path := uriParts[0]
	if len(uriParts) > 1 {
		// match route
		path = uriParts[0] + "/" + uriParts[1]
	}

	for _, route := range r.router.routes {
		if route.Path == path || route.Path == fullPath {
			// reduce complexity for routes with only one method
			if len(route.Methods) == 1 && request.Method == route.Methods[0] {
				r.cache.push(cachePath, route)
				return route
			}

			// iterate methods for routes with multiple methods
			for _, me := range route.Methods {
				if me == request.Method {
					r.cache.push(cachePath, route)
					return route
				}
			}
		}
	}

	return nil
}
