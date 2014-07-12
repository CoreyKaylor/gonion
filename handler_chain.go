package gonion

import (
	"net/http"
)

type ChainHandler interface {
	ServeHTTP(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request)
}

type ChainHandlerFunc func(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request)

func (c ChainHandlerFunc) ServeHTTP(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request) {
	c(requestContext, rw, r)
}

type ChainLink func(ChainHandler) ChainHandler

func build(handler ChainHandler, middleware []*Middleware, contextFactory func() interface{}) http.Handler {
	chain := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		chain = middleware[i].Handler(chain)
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestContext := make(map[string]interface{})
		userContext := contextFactory()
		if userContext != nil {
			requestContext["user-context"] = userContext
		}
		chain.ServeHTTP(requestContext, rw, req)
	})
}

//Storage for route information
type RouteRegistry struct {
	Routes []*RouteModel
}

type RouteModel struct {
	Method         string
	Pattern        string
	Handler        ChainHandler
	ContextOptions *ContextOptions
}

func (r *RouteRegistry) AddRoute(method string, pattern string, handler ChainHandler) *ContextOptions {
	contextOptions := newContextOptions()
	route := &RouteModel{method, pattern, handler, contextOptions}
	r.Routes = append(r.Routes, route)
	return contextOptions
}

//Creates a new RouteRegistry for storing route information
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{make([]*RouteModel, 0, 10)}
}

type MiddlewareRegistry struct {
	Middleware []*Middleware
}

type Middleware struct {
	Filter         RouteFilter
	Handler        ChainLink
	ContextOptions *ContextOptions
}

type RouteFilter func(*RouteModel) bool

func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{make([]*Middleware, 0, 10)}
}

func (m *MiddlewareRegistry) AppliesToAllRoutes(handler ChainLink) *ContextOptions {
	contextOptions := newContextOptions()
	m.Add(func(route *RouteModel) bool {
		return true
	}, handler, contextOptions)
	return contextOptions
}

func (m *MiddlewareRegistry) Add(filter RouteFilter, handler ChainLink, contextOptions *ContextOptions) {
	m.Middleware = append(m.Middleware, &Middleware{filter, handler, contextOptions})
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
