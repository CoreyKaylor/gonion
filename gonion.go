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

func (composer *Composer) addMiddleware(link ChainLink) {
	composer.middlewareRegistry.Add(func(route *RouteModel) bool {
		return composer.start == "" || strings.HasPrefix(route.Pattern, composer.start)
	}, link)
}

func wrap(handler http.Handler) ChainLink {
	return ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.rw, context.req)
			inner.ServeHTTP(context)
		})
	})
}

func (composer *Composer) UseFunc(handler func(http.ResponseWriter, *http.Request)) {
	composer.Use(http.HandlerFunc(handler))
}

func (composer *Composer) Get(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	composer.Handle("GET", pattern, http.HandlerFunc(handler))
}

func (composer *Composer) Use(handler http.Handler) {
	composer.addMiddleware(wrap(handler))
}

type WrappingHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler)
}

type WrappingHandlerFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (wh WrappingHandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	wh(rw, r, next)
}

func (composer *Composer) UseWrappingHandlerFunc(handler func(http.ResponseWriter, *http.Request, http.Handler)) {
	composer.UseWrappingHandler(WrappingHandlerFunc(handler))
}

func (composer *Composer) UseWrappingHandler(handler WrappingHandler) {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.rw, context.req, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				context.rw = rw
				context.req = r
				inner.ServeHTTP(context)
			}))
		})
	})
	composer.addMiddleware(chainLink)
}

func (composer *Composer) UseMiddlewareConstructor(ctor func(http.Handler) http.Handler) {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			current := ctor(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				context.rw = rw
				context.req = r
				inner.ServeHTTP(context)
			}))
			current.ServeHTTP(context.rw, context.req)
		})
	})
	composer.addMiddleware(chainLink)
}

type ContextWrapper interface {
	Wrap() func(interface{}, http.ResponseWriter, *http.Request)
}

var fac func() interface{} = func() interface{} {
	return nil
}

func (composer *Composer) CreateContext(factory func() interface{}) {
	fac = factory
}

func (composer *Composer) UseContextualHandler(wrapper ContextWrapper) {
	chainLink := composer.buildChainLink(wrapper)
	composer.addMiddleware(chainLink)
}

func (composer *Composer) buildChainLink(wrapper ContextWrapper) ChainLink {
	contextFunc := wrapper.Wrap()
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			contextFunc(context.i, context.rw, context.req)
			inner.ServeHTTP(context)
		})
	})
	return chainLink
}

func (composer *Composer) GetC(pattern string, wrapper ContextWrapper) {
	contextFunc := wrapper.Wrap()
	chainHandlerFunc := ChainHandlerFunc(func(context *ChainContext) {
		contextFunc(context.i, context.rw, context.req)
	})
	composer.routeRegistry.AddRoute("GET", composer.start+pattern, chainHandlerFunc)
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
		handler := build(route.Handler, middleware)
		routes = append(routes, &Route{route.Method, route.Pattern, handler})
	}
	return routes
}
