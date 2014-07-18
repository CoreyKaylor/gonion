package gonion

import (
	"net/http"
)

//MiddlewareOptions is accessed from calling Use() and
//is how you specify which way to register a middleware handler.
type MiddlewareOptions struct {
	composer    *Composer
	routeFilter func(*RouteModel) bool
}

//ChainLink is called when your middleware handler needs to wrap the rest
//of the handler chain.
func (mo *MiddlewareOptions) ChainLink(ctor func(http.Handler) http.Handler) {
	mo.composer.addMiddleware(ChainLink(ctor), mo.routeFilter)
}

func wrap(handler http.Handler) ChainLink {
	return ChainLink(func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r)
			inner.ServeHTTP(rw, r)
		})
	})
}

//Handler is middleware that conforms to the standard http.Handler interface
func (mo *MiddlewareOptions) Handler(handler http.Handler) {
	mo.composer.addMiddleware(wrap(handler), mo.routeFilter)
}

//Func is a convenience method for a func that matches the signature of the
//standard http.HandlerFunc.
func (mo *MiddlewareOptions) Func(handler func(http.ResponseWriter, *http.Request)) {
	mo.Handler(http.HandlerFunc(handler))
}
