package gonion

import (
	"net/http"
)

type ChainLink func(http.Handler) http.Handler

func build(handler http.Handler, middleware []*Middleware) http.Handler {
	chain := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		chain = middleware[i].Handler(chain)
	}
	return chain
}

//Storage for route information
type RouteRegistry struct {
	Routes []*RouteModel
}

type RouteModel struct {
	Method  string
	Pattern string
	Handler http.Handler
}

func (r *RouteRegistry) AddRoute(method string, pattern string, handler http.Handler) {
	route := &RouteModel{method, pattern, handler}
	r.Routes = append(r.Routes, route)
}

//Creates a new RouteRegistry for storing route information
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{make([]*RouteModel, 0, 10)}
}

type MiddlewareRegistry struct {
	Middleware []*Middleware
}

type Middleware struct {
	Filter  RouteFilter
	Handler ChainLink
}

type RouteFilter func(*RouteModel) bool

func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{make([]*Middleware, 0, 10)}
}

func (m *MiddlewareRegistry) AppliesToAllRoutes(handler ChainLink) {
	m.Add(func(route *RouteModel) bool {
		return true
	}, handler)
}

func (m *MiddlewareRegistry) Add(filter RouteFilter, handler ChainLink) {
	m.Middleware = append(m.Middleware, &Middleware{filter, handler})
}

func (m *MiddlewareRegistry) MiddlewareFor(route *RouteModel) []*Middleware {
	ret := make([]*Middleware, 0, 10)
	for _, middle := range m.Middleware {
		if middle.Filter(route) {
			ret = append(ret, middle)
		}
	}
	return ret
}
