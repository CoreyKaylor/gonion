package gonion

import (
	"net/http"
)

type MiddlewareOptions struct {
	composer *Composer
}

func (mo *MiddlewareOptions) ChainLink(ctor func(http.Handler) http.Handler) {
	mo.composer.addMiddleware(ChainLink(ctor))
}

func wrap(handler http.Handler) ChainLink {
	return ChainLink(func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r)
			inner.ServeHTTP(rw, r)
		})
	})
}

func (mo *MiddlewareOptions) Handler(handler http.Handler) {
	mo.composer.addMiddleware(wrap(handler))
}

func (mo *MiddlewareOptions) Func(handler func(http.ResponseWriter, *http.Request)) {
	mo.Handler(http.HandlerFunc(handler))
}
