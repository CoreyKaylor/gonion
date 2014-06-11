package gonion

import (
	"net/http"
	"strings"
)

type Composer struct {
	start              string
	routeRegistry      *RouteRegistry
	middlewareRegistry *MiddlewareRegistry
}

func New() *Composer {
	routeRegistry := NewRouteRegistry()
	middlewareRegistry := NewMiddlewareRegistry()
	return &Composer{"", routeRegistry, middlewareRegistry}
}

func (composer *Composer) Sub(pattern string, sub func(*Composer)) {
	subComposer := &Composer{composer.start + pattern, composer.routeRegistry, composer.middlewareRegistry}
	sub(subComposer)
}

type ContextOptions struct {
	factory  func() interface{}
	replaced bool
}

func (co *ContextOptions) CreateContext(factory func() interface{}) {
	co.factory = factory
	co.replaced = true
}

func newContextOptions() *ContextOptions {
	return &ContextOptions{fac, false}
}

func (composer *Composer) addMiddleware(link ChainLink) *ContextOptions {
	contextOptions := newContextOptions()
	composer.middlewareRegistry.Add(func(route *RouteModel) bool {
		return composer.start == "" || strings.HasPrefix(route.Pattern, composer.start)
	}, link, contextOptions)
	return contextOptions
}

func (composer *Composer) Use() *MiddlewareOptions {
	return &MiddlewareOptions{composer}
}

func (composer *Composer) Get(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	composer.Handle("GET", pattern, http.HandlerFunc(handler))
}

type ContextHandler interface {
	ServeHTTP(interface{}, http.ResponseWriter, *http.Request)
}

var fac func() interface{} = func() interface{} {
	return nil
}

func (composer *Composer) GetC(pattern string, handler ContextHandler) *ContextOptions {
	chainHandlerFunc := ChainHandlerFunc(func(context *ChainContext) {
		handler.ServeHTTP(context.i, context.rw, context.req)
	})
	return composer.routeRegistry.AddRoute("GET", composer.start+pattern, chainHandlerFunc)
}

func (composer *Composer) Handle(method string, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	handlerFunc := http.HandlerFunc(handler)
	composer.routeRegistry.AddRoute(method, composer.start+pattern, wrapHandler(handlerFunc))
}

func wrapHandler(handler http.Handler) ChainHandler {
	return ChainHandlerFunc(func(context *ChainContext) {
		handler.ServeHTTP(context.rw, context.req)
	})
}

type Routes []*Route

type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
}

func (composer *Composer) BuildRoutes() Routes {
	routes := make(Routes, 0, 10)
	for _, route := range composer.routeRegistry.Routes {
		middleware := composer.middlewareRegistry.MiddlewareFor(route)
		var factory func() interface{} = fac
		if route.ContextOptions.replaced {
			factory = route.ContextOptions.factory
		} else {
			for _, m := range middleware {
				if m.ContextOptions.replaced {
					//last one wins
					factory = m.ContextOptions.factory
				}
			}
		}

		handler := build(route.Handler, middleware, factory)
		routes = append(routes, &Route{route.Method, route.Pattern, handler})
	}
	return routes
}
