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
	return &Composer{
		start:              "",
		routeRegistry:      routeRegistry,
		middlewareRegistry: middlewareRegistry,
	}
}

func (composer *Composer) Sub(pattern string, sub func(*Composer)) {
	subComposer := &Composer{
		start:              composer.start + pattern,
		routeRegistry:      composer.routeRegistry,
		middlewareRegistry: composer.middlewareRegistry,
	}
	sub(subComposer)
}

func (composer *Composer) addMiddleware(link ChainLink, routeFilter func(*RouteModel) bool) {
	composer.middlewareRegistry.Add(func(route *RouteModel) bool {
		return (composer.start == "" || strings.HasPrefix(route.Pattern, composer.start)) && routeFilter(route)
	}, link)
}

func (composer *Composer) Use() *MiddlewareOptions {
	return composer.useWhen(func(route *RouteModel) bool {
		return true
	})
}

func (composer *Composer) useWhen(routeFilter func(*RouteModel) bool) *MiddlewareOptions {
	return &MiddlewareOptions{
		composer:    composer,
		routeFilter: routeFilter,
	}
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

type RouteConstraint struct {
	composer    *Composer
	routeFilter func(*RouteModel) bool
}

func (composer *Composer) Only() *RouteConstraint {
	return &RouteConstraint{
		composer:    composer,
		routeFilter: nil,
	}
}

func (rc *RouteConstraint) When(routeFilter func(*RouteModel) bool) *RouteConstraint {
	rc.routeFilter = routeFilter
	return rc
}

func (rc *RouteConstraint) Get() *RouteConstraint {
	return rc.methodConstraint("GET")
}

func (rc *RouteConstraint) Post() *RouteConstraint {
	return rc.methodConstraint("POST")
}

func (rc *RouteConstraint) Put() *RouteConstraint {
	return rc.methodConstraint("PUT")
}

func (rc *RouteConstraint) Patch() *RouteConstraint {
	return rc.methodConstraint("PATCH")
}

func (rc *RouteConstraint) Delete() *RouteConstraint {
	return rc.methodConstraint("DELETE")
}

func (rc *RouteConstraint) methodConstraint(method string) *RouteConstraint {
	return rc.When(func(route *RouteModel) bool {
		return route.Method == method
	})
}

func (rc *RouteConstraint) Use() *MiddlewareOptions {
	return rc.composer.useWhen(rc.routeFilter)
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
		route := &Route{
			Method:  route.Method,
			Pattern: route.Pattern,
			Handler: handler,
		}
		routes = append(routes, route)
	}
	return routes
}

func (composer *Composer) EachRoute(router func(*Route)) {
	routes := composer.BuildRoutes()
	for _, route := range routes {
		router(route)
	}
}
