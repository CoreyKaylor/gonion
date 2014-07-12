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
		return ChainHandlerFunc(func(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r, http.HandlerFunc(func(rwAfter http.ResponseWriter, rAfter *http.Request) {
				inner.ServeHTTP(requestContext, rwAfter, rAfter)
			}))
		})
	})
	mo.composer.addMiddleware(chainLink)
}

func (mo *MiddlewareOptions) ConstructorFunc(ctor func(http.Handler) http.Handler) {
	chainLink := ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request) {
			current := ctor(http.HandlerFunc(func(rwAfter http.ResponseWriter, rAfter *http.Request) {
				inner.ServeHTTP(requestContext, rwAfter, rAfter)
			}))
			current.ServeHTTP(rw, r)
		})
	})
	mo.composer.addMiddleware(chainLink)
}

func wrap(handler http.Handler) ChainLink {
	return ChainLink(func(inner ChainHandler) ChainHandler {
		return ChainHandlerFunc(func(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(rw, r)
			inner.ServeHTTP(requestContext, rw, r)
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
		return ChainHandlerFunc(func(requestContext map[string]interface{}, rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(requestContext["user-context"], rw, r)
			inner.ServeHTTP(requestContext, rw, r)
		})
	})
	return mo.composer.addMiddleware(chainLink)
}
