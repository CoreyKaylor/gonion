package gonion

import (
	"net/http"
)

type ChainHandler interface {
	ServeHTTP(*ChainContext)
}

type ChainContext struct {
	rw  http.ResponseWriter
	req *http.Request
	i   interface{} //user specific context
}

type ChainHandlerFunc func(*ChainContext)

func (c ChainHandlerFunc) ServeHTTP(context *ChainContext) {
	c(context)
}

type ChainLink func(ChainHandler) ChainHandler

func build(handler ChainHandler, chainLinks []ChainLink) http.Handler {
	chain := handler
	for i := len(chainLinks) - 1; i >= 0; i-- {
		chain = chainLinks[i](chain)
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		context := &ChainContext{rw, req, fac()}
		chain.ServeHTTP(context)
	})
}

//Storage for route information
type RouteRegistry struct {
	Routes []*RouteModel
}

type RouteModel struct {
	Method  string
	Pattern string
	Handler ChainHandler
}

func (r *RouteRegistry) AddRoute(method string, pattern string, handler ChainHandler) {
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

func (m *MiddlewareRegistry) MiddlewareFor(route *RouteModel) []ChainLink {
	ret := make([]ChainLink, 0, 10)
	for _, middle := range m.Middleware {
		if middle.Filter(route) {
			ret = append(ret, middle.Handler)
		}
	}
	return ret
}
