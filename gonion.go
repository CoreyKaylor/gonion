package gonion

type App struct {
	routeRegistry      *RouteRegistry
	chainBuilder       *ChainBuilder
	middlewareRegistry *MiddlewareRegistry
}

func New() *App {
	routeRegistry := NewRouteRegistry()
	chainBuilder := NewChainBuilder()
	middlewareRegistry := NewMiddlewareRegistry()
	return &App{routeRegistry, chainBuilder, middlewareRegistry}
}

var noName string = ""

func (app *App) UseNamed(name string, handler Handler) {
	//todo: add ability to specify order globally
	app.middlewareRegistry.AppliesToAllRoutes(name, handler)
}

func (app *App) Use(handler Handler) {
	app.UseNamed(noName, handler)
}

func (app *App) Handle(handler Handler) {
	//todo: make variadic?
	app.routeRegistry.AddFunc(handler)
}

type Runtime struct {
	Routes []*Route
}

type Route struct {
	Method  string
	Pattern string
	Handler HandlerFunc
}

func (app *App) buildRuntime() *Runtime {
	runtime := &Runtime{}
	runtime.Routes = make([]*Route, 0, 10)
	for _, route := range app.routeRegistry.Routes {
		middleware := app.middlewareRegistry.MiddlewareFor(route)
		chain := app.chainBuilder.build(append(middleware, route.Handler)...)
		runtime.Routes = append(runtime.Routes, &Route{route.Method, route.Pattern, chain})
	}
	return runtime
}
