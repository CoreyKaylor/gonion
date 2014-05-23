package gonion

import (
	"net/http"
)

type Arg interface{}

type HandlerMap func(Handler, ContextFunc) (ContextFunc, bool)

type ChainBuilder struct {
	HandlerMaps []HandlerMap
}

func NewChainBuilder() *ChainBuilder {
	registry := ChainBuilder{make([]HandlerMap, 0, 10)}
	registry.AddMap(WrapStandardHandlerWithContext)
	registry.AddMap(WrapWrappingHandlerWithContext)
	return &registry
}

func (r *ChainBuilder) AddMap(l HandlerMap) {
	r.HandlerMaps = append(r.HandlerMaps, l)
}

func (r *ChainBuilder) GetInvoker(h Handler, next ContextFunc) (ContextFunc, bool) {
	for i := len(r.HandlerMaps) - 1; i >= 0; i-- {
		if invoker, needsNext := r.HandlerMaps[i](h, next); invoker != nil {
			return invoker, needsNext
		}
	}

	panic("gonion: no invoker found for Handler")
}

func (r *ChainBuilder) build(handlers ...Handler) HandlerFunc {
	firstFunc, _ := r.GetInvoker(handlers[len(handlers)-1], ContextFunc(func(context Context) {
	}))
	chain := func(context Context) {
		firstFunc(context)
	}
	for i := len(handlers) - 2; i >= 0; i-- {
		currentChain := chain
		current, needsNext := r.GetInvoker(handlers[i], currentChain)
		if !needsNext {
			chain = func(context Context) {
				current(context)
				currentChain(context)
			}
		} else {
			chain = current
		}
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		context := Context{rw, req}
		chain(context)
	}
}

type Context struct {
	rw  http.ResponseWriter
	req *http.Request
}

type ContextFunc func(Context)

func WrapStandardHandlerWithContext(h Handler, next ContextFunc) (ContextFunc, bool) {
	if fun, ok := h.(func(http.ResponseWriter, *http.Request)); ok {
		return func(context Context) {
			fun(context.rw, context.req)
		}, false
	}
	return nil, false
}

func WrapWrappingHandlerWithContext(h Handler, next ContextFunc) (ContextFunc, bool) {
	if fun, ok := h.(func(http.ResponseWriter, *http.Request, NextHandler)); ok {
		return func(context Context) {
			fun(context.rw, context.req, func(rw http.ResponseWriter, req *http.Request) {
				context.rw, context.req = rw, req
				next(context)
			})
		}, true
	}
	return nil, false
}

type HandlerFunc func(http.ResponseWriter, *http.Request)
type NextHandler func(http.ResponseWriter, *http.Request)
type WrappingHandler func(http.ResponseWriter, *http.Request, NextHandler)

type MiddlewareRegistry struct {
	Middleware []*Middleware
}

type Middleware struct {
	Name    string
	Filter  RouteFilter
	Handler Handler
}

type RouteFilter func(*RouteModel) bool

func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{make([]*Middleware, 0, 10)}
}

func (m *MiddlewareRegistry) AppliesToAllRoutes(name string, handler Handler) {
	m.Add(name, func(route *RouteModel) bool {
		return true
	}, handler)
}

func (m *MiddlewareRegistry) Add(name string, filter RouteFilter, handler Handler) {
	m.Middleware = append(m.Middleware, &Middleware{name, filter, handler})
}

func (m *MiddlewareRegistry) MiddlewareFor(route *RouteModel) []Handler {
	ret := make([]Handler, 0, 10)
	for _, middle := range m.Middleware {
		if middle.Filter(route) {
			ret = append(ret, middle.Handler)
		}
	}
	return ret
}
