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

func (composer *Composer) Use() *MiddlewareOptions {
	return &MiddlewareOptions{composer}
}

func (composer *Composer) Get(pattern string, handler http.Handler) {
	composer.Handle("GET", pattern, handler)
}

func (composer *Composer) Post(pattern string, handler http.Handler) {
	composer.Handle("POST", pattern, handler)
}

func (composer *Composer) Put(pattern string, handler http.Handler) {
	composer.Handle("PUT", pattern, handler)
}

func (composer *Composer) Patch(pattern string, handler http.Handler) {
	composer.Handle("PATCH", pattern, handler)
}

func (composer *Composer) Delete(pattern string, handler http.Handler) {
	composer.Handle("DELETE", pattern, handler)
}

func (composer *Composer) Handle(method string, pattern string, handler http.Handler) {
	composer.routeRegistry.AddRoute(method, composer.start+pattern, handler)
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
