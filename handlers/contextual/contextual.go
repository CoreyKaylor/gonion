package contextual

import (
	"net/http"

	"code.google.com/p/go.net/context"
)

//Handler is similar to http.Handler but with the added context.Context parameter
//that can be used to share data between layers of an API
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

//HandlerFunc is a simple func to represent a contextual handler
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

//ServeHTTP is the HandlerFunc implementation of the contextual.Handler interface
func (c HandlerFunc) ServeHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	c(ctx, rw, r)
}

//ChainedHandler represents a contextual chain handler that requires each handler
//to wrap the rest of the chain
type ChainedHandler interface {
	ChainLink(Handler) Handler
}

//ChainLinkFunc is a simple func that represents a ChainedHandler
type ChainLinkFunc func(Handler) Handler

//ChainLink is the implementation of ChainedHandler for the ChainLinkFunc
func (chain ChainLinkFunc) ChainLink(inner Handler) Handler {
	return chain(inner)
}

//Handlers is a variadic function that will describe the order of contextual handlers
//and returns a standard http.Handler to be used for the request.
func Handlers(handlers ...ChainedHandler) http.Handler {
	contextChain := buildChain(handlers...)
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		contextChain.ServeHTTP(context.Background(), rw, r)
	})
}

//Last is intended for the last handler in the chain that will not require wrapping the next handler.
func Last(handler Handler) ChainedHandler {
	return ChainLinkFunc(func(inner Handler) Handler {
		return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(ctx, rw, r)
			//ignore inner
		})
	})
}

func buildChain(handlers ...ChainedHandler) Handler {
	length := len(handlers)
	var chain Handler
	for i := length - 1; i >= 0; i-- {
		chain = handlers[i].ChainLink(chain)
	}
	return chain
}
