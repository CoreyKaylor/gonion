package gonion

import (
	"net/http"
)

//ChainLink is used when your http.Handler needs to wrap the rest
//of the handler chain.
type ChainLink func(http.Handler) http.Handler

func build(handler http.Handler, middleware []*middleware) http.Handler {
	chain := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		chain = middleware[i].handler(chain)
	}
	return chain
}

type routeRegistry struct {
	routes []*RouteModel
}

//RouteModel is the pre-build model representing a single handler
//without middleware.
type RouteModel struct {
	Method  string
	Pattern string
	Handler http.Handler
}

func (r *routeRegistry) addRoute(method string, pattern string, handler http.Handler) {
	route := &RouteModel{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}
	r.routes = append(r.routes, route)
}

func newRouteRegistry() *routeRegistry {
	return &routeRegistry{
		routes: make([]*RouteModel, 0, 10),
	}
}

type middlewareRegistry struct {
	middleware []*middleware
}

type middleware struct {
	filter  routeFilter
	handler ChainLink
}

type routeFilter func(*RouteModel) bool

func newMiddlewareRegistry() *middlewareRegistry {
	return &middlewareRegistry{
		middleware: make([]*middleware, 0, 10),
	}
}

func (m *middlewareRegistry) appliesToAllRoutes(handler ChainLink) {
	m.add(func(route *RouteModel) bool {
		return true
	}, handler)
}

func (m *middlewareRegistry) add(filter routeFilter, handler ChainLink) {
	middleware := &middleware{
		filter:  filter,
		handler: handler,
	}
	m.middleware = append(m.middleware, middleware)
}

func (m *middlewareRegistry) middlewareFor(route *RouteModel) []*middleware {
	ret := make([]*middleware, 0, 10)
	for _, middle := range m.middleware {
		if middle.filter(route) {
			ret = append(ret, middle)
		}
	}
	return ret
}
