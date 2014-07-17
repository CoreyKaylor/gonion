package gonion

import (
	"net/http"
	"strings"
)

//Composer is the main API in gonion and is responsible for
//composing the middleware and routes of your application
type Composer struct {
	start              string
	routeRegistry      *RouteRegistry
	middlewareRegistry *MiddlewareRegistry
}

//New is a factory method for Composer
func New() *Composer {
	routeRegistry := NewRouteRegistry()
	middlewareRegistry := NewMiddlewareRegistry()
	return &Composer{
		start:              "",
		routeRegistry:      routeRegistry,
		middlewareRegistry: middlewareRegistry,
	}
}

//Sub will allow you to specify middleware and routes that only
//apply for the specified path. This can be done recursively and
//at each level will inherit the previous path's middleware.
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

//Use is the entrypoint to adding middleware
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

//Get adds a route constrained to only 'GET' requests
func (composer *Composer) Get(pattern string, handler http.Handler) {
	composer.Handle("GET", pattern, handler)
}

//Post adds a route constrained to only 'POST' requests
func (composer *Composer) Post(pattern string, handler http.Handler) {
	composer.Handle("POST", pattern, handler)
}

//Put adds a route constrained to only 'PUT' requests
func (composer *Composer) Put(pattern string, handler http.Handler) {
	composer.Handle("PUT", pattern, handler)
}

//Patch adds a route constrained to only 'PATCH' requests
func (composer *Composer) Patch(pattern string, handler http.Handler) {
	composer.Handle("PATCH", pattern, handler)
}

//Delete adds a route constrained to only 'DELETE' requests
func (composer *Composer) Delete(pattern string, handler http.Handler) {
	composer.Handle("DELETE", pattern, handler)
}

//Handle adds a route for the specified method and pattern
func (composer *Composer) Handle(method string, pattern string, handler http.Handler) {
	composer.routeRegistry.AddRoute(method, composer.start+pattern, handler)
}

//RouteConstraint is how middleware is constrained after calling Only()
type RouteConstraint struct {
	composer    *Composer
	routeFilter func(*RouteModel) bool
}

//Only allows you to constrain middleware for only certain types of routes.
//This isn't just a runtime filter, it excludes it from the chain while building the routes
//on startup.
func (composer *Composer) Only() *RouteConstraint {
	return &RouteConstraint{
		composer:    composer,
		routeFilter: nil,
	}
}

//When is a constraint that gives you all the route information to filter upon.
func (rc *RouteConstraint) When(routeFilter func(*RouteModel) bool) *RouteConstraint {
	rc.routeFilter = routeFilter
	return rc
}

//Get constrains the middleware to only apply for 'GET' requests
func (rc *RouteConstraint) Get() *RouteConstraint {
	return rc.methodConstraint("GET")
}

//Post constrains the middleware to only apply for 'POST' requests
func (rc *RouteConstraint) Post() *RouteConstraint {
	return rc.methodConstraint("POST")
}

//Put constrains the middleware to only apply for 'PUT' requests
func (rc *RouteConstraint) Put() *RouteConstraint {
	return rc.methodConstraint("PUT")
}

//Patch constrains the middleware to only apply for 'PATCH' requests
func (rc *RouteConstraint) Patch() *RouteConstraint {
	return rc.methodConstraint("PATCH")
}

//Delete constrains the middleware to only apply for 'DELETE' requests
func (rc *RouteConstraint) Delete() *RouteConstraint {
	return rc.methodConstraint("DELETE")
}

func (rc *RouteConstraint) methodConstraint(method string) *RouteConstraint {
	return rc.When(func(route *RouteModel) bool {
		return route.Method == method
	})
}

//Use is the entrypoint to defining your middleware, but only for the current
//defined route constraint
func (rc *RouteConstraint) Use() *MiddlewareOptions {
	return rc.composer.useWhen(rc.routeFilter)
}

//Routes is the array of routes and built middleware. This will be what's returned
//after calling BuildRoutes
type Routes []*Route

//Route is the handler and route information after calling BuildRoutes. Handler
//is the entire chain of route handler and middleware.
type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
}

//BuildRoutes returns routes with their corresponding handler chain.
//This is typically what you will call before delegating to the router
//you have chosen for your application.
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

//EachRoute is a convenience method for BuildRoutes that you can
//pass a func to be called on each route from BuildRoutes.
func (composer *Composer) EachRoute(router func(*Route)) {
	routes := composer.BuildRoutes()
	for _, route := range routes {
		router(route)
	}
}
