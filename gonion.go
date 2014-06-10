package gonion

import (
	"net/http"
	"strings"
)

type App struct {
	start              string
	routeRegistry      *RouteRegistry
	middlewareRegistry *MiddlewareRegistry
}

func New() *App {
	routeRegistry := NewRouteRegistry()
	middlewareRegistry := NewMiddlewareRegistry()
	return &App{"", routeRegistry, middlewareRegistry}
}

func (app *App) Sub(pattern string, sub func(*App)) {
	subApp := &App{app.start + pattern, app.routeRegistry, app.middlewareRegistry}
	sub(subApp)
}

func (app *App) addMiddleware(link ChainLink) {
	app.middlewareRegistry.Add(func(route *RouteModel) bool {
		return app.start == "" || strings.HasPrefix(route.Pattern, app.start)
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

func (app *App) UseFunc(handler func(http.ResponseWriter, *http.Request)) {
	app.Use(http.HandlerFunc(handler))
}

func (app *App) Get(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	app.Handle("GET", pattern, http.HandlerFunc(handler))
}

func (app *App) Use(handler http.Handler) {
	app.addMiddleware(wrap(handler))
}

type WrappingHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler)
}

type WrappingHandlerFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (wh WrappingHandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	wh(rw, r, next)
}

func (app *App) UseWrappingHandlerFunc(handler func(http.ResponseWriter, *http.Request, http.Handler)) {
	app.UseWrappingHandler(WrappingHandlerFunc(handler))
}

func (app *App) UseWrappingHandler(handler WrappingHandler) {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.rw, context.req, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				context.rw = rw
				context.req = r
				inner.ServeHTTP(context)
			}))
		})
	})
	app.addMiddleware(chainLink)
}

func (app *App) UseMiddlewareConstructor(ctor func(http.Handler) http.Handler) {
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
	app.addMiddleware(chainLink)
}

type ContextWrapper interface {
	Wrap() func(interface{}, http.ResponseWriter, *http.Request)
}

var fac func() interface{} = func() interface{} {
	return nil
}

func (app *App) CreateContext(factory func() interface{}) {
	fac = factory
}

func (app *App) UseContextualHandler(wrapper ContextWrapper) {
	chainLink := app.buildChainLink(wrapper)
	app.addMiddleware(chainLink)
}

func (app *App) buildChainLink(wrapper ContextWrapper) ChainLink {
	contextFunc := wrapper.Wrap()
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			contextFunc(context.i, context.rw, context.req)
			inner.ServeHTTP(context)
		})
	})
	return chainLink
}

func (app *App) GetC(pattern string, wrapper ContextWrapper) {
	contextFunc := wrapper.Wrap()
	chainHandlerFunc := ChainHandlerFunc(func(context *ChainContext) {
		contextFunc(context.i, context.rw, context.req)
	})
	app.routeRegistry.AddRoute("GET", app.start+pattern, chainHandlerFunc)
}

func (app *App) Handle(method string, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	handlerFunc := http.HandlerFunc(handler)
	app.routeRegistry.AddRoute(method, app.start+pattern, wrapHandler(handlerFunc))
}

func wrapHandler(handler http.Handler) ChainHandler {
	return ChainHandlerFunc(func(context *ChainContext) {
		handler.ServeHTTP(context.rw, context.req)
	})
}

type Runtime struct {
	Routes []*Route
}

type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
}

func (app *App) BuildRoutes() *Runtime {
	runtime := &Runtime{}
	runtime.Routes = make([]*Route, 0, 10)
	for _, route := range app.routeRegistry.Routes {
		middleware := app.middlewareRegistry.MiddlewareFor(route)
		handler := build(route.Handler, middleware)
		runtime.Routes = append(runtime.Routes, &Route{route.Method, route.Pattern, handler})
	}
	return runtime
}
