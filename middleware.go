package gonion

import (
	"net/http"
)

type MiddlewareOptions struct {
	composer *Composer
}

type WrappingHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler)
}

type WrappingHandlerFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (wh WrappingHandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	wh(rw, r, next)
}

func (mo *MiddlewareOptions) WrappingFunc(handler func(http.ResponseWriter, *http.Request, http.Handler)) {
	mo.WrappingHandler(WrappingHandlerFunc(handler))
}

func (mo *MiddlewareOptions) WrappingHandler(handler WrappingHandler) {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.rw, context.req, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				context.rw = rw
				context.req = r
				inner.ServeHTTP(context)
			}))
		})
	})
	mo.composer.addMiddleware(chainLink)
}

func (mo *MiddlewareOptions) ConstructorFunc(ctor func(http.Handler) http.Handler) {
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
	mo.composer.addMiddleware(chainLink)
}

func wrap(handler http.Handler) ChainLink {
	return ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.rw, context.req)
			inner.ServeHTTP(context)
		})
	})
}

func (mo *MiddlewareOptions) Handler(handler http.Handler) {
	mo.composer.addMiddleware(wrap(handler))
}

func (mo *MiddlewareOptions) Func(handler func(http.ResponseWriter, *http.Request)) {
	mo.Handler(http.HandlerFunc(handler))
}

func (mo *MiddlewareOptions) HandlerFunc(handler http.HandlerFunc) {
	mo.Handler(handler)
}

func (mo *MiddlewareOptions) ContextHandler(handler ContextHandler) *ContextOptions {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(context *ChainContext) {
			handler.ServeHTTP(context.i, context.rw, context.req)
			inner.ServeHTTP(context)
		})
	})
	return mo.composer.addMiddleware(chainLink)
}
