package gonion

type MiddlewareRegistry struct {
	Middleware []*Middleware
}

type Middleware struct {
	Name    string
	Filter  RouteFilter
	Handler Handler
}

type RouteFilter func(*Route) bool

func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{make([]*Middleware, 0, 10)}
}

func (m *MiddlewareRegistry) AppliesToAllRoutes(name string, handler Handler) {
	m.Add(name, func(route *Route) bool {
		return true
	}, handler)
}

func (m *MiddlewareRegistry) Add(name string, filter RouteFilter, handler Handler) {
	m.Middleware = append(m.Middleware, &Middleware{name, filter, handler})
}

func (m *MiddlewareRegistry) MiddlewareFor(route *Route) []*Middleware {
	ret := make([]*Middleware, 0, 10)
	for _, middle := range m.Middleware {
		if middle.Filter(route) {
			ret = append(ret, middle)
		}
	}
	return ret
}
