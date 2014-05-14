package gonion

import (
	"net/http"
)

type App struct {
	routerRegistry     *RouteRegistry
	invocationRegistry *InvocationRegistry
	middlewareRegistry *MiddlewareRegistry
}

func New() *App {
	routeRegistry := NewRouteRegistry()
	invocationRegistry := NewInvocationRegistry()
	middlewareRegistry := NewMiddlewareRegistry()
	return &App{routeRegistry, invocationRegistry, middlewareRegistry}
}

func (app *App) Use(name string, handler Handler) {
	app.middlewareRegistry.AppliesToAllRoutes(name, handler)
}

func (app *App) Handle(handler Handler) {
	app.routerRegistry.AddFunc(handler)
}

type Runtime struct {
	Routes []*RuntimeRoute
}

type RouteHandler func(rw http.ResponseWriter, r *http.Request)

type RuntimeRoute struct {
	Method  string
	Pattern string
	Handler RouteHandler
}

func (app *App) buildRuntime() *Runtime {
	runtime := &Runtime{}
	runtime.Routes = make([]*RuntimeRoute, 0, 10)
	for _, route := range app.routerRegistry.Routes {
		invoker := app.invocationRegistry.GetInvoker(route.Handler)
		handler := func(rw http.ResponseWriter, r *http.Request) {
			//much more to do here, but this is where it starts
			invoker(rw, r)
		}
		runtime.Routes = append(runtime.Routes, &RuntimeRoute{route.Method, route.Pattern, handler})
	}
	return runtime
}
