package gonion

import (
	"net/http"
	"reflect"
)

type Arg interface{}

type Invoker func(...Arg)

type InvokerLocator func(Handler) Invoker

type InvocationRegistry struct {
	Locators []InvokerLocator
}

func NewInvocationRegistry() *InvocationRegistry {
	registry := InvocationRegistry{make([]InvokerLocator, 0, 10)}
	registry.AddLocator(InvokeAnythingWithReflectionFallback)
	registry.AddLocator(InvokeStringArgumentNoReturn)
	registry.AddLocator(InvokeHttpHandlerFunc)
	return &registry
}

func (r *InvocationRegistry) AddLocator(l InvokerLocator) {
	r.Locators = append(r.Locators, l)
}

func (r *InvocationRegistry) GetInvoker(h Handler) Invoker {
	for i := len(r.Locators) - 1; i >= 0; i-- {
		if invoker := r.Locators[i](h); invoker != nil {
			return invoker
		}
	}

	panic("gonion: no invoker found for Handler")
}

func InvokeAnythingWithReflectionFallback(h Handler) Invoker {
	value := reflect.ValueOf(h)
	handlerType := value.Type()
	numArgs := handlerType.NumIn()

	return func(params ...Arg) {
		if params != nil && len(params) != numArgs {
			panic("gonion: incorrect number of arguments for Invoke")
		}
		args := make([]reflect.Value, numArgs)
		for i := 0; i < numArgs; i++ {
			args[i] = reflect.ValueOf(params[i])
		}
		value.Call(args)
	}
}

func InvokeStringArgumentNoReturn(h Handler) Invoker {
	if fun, ok := h.(func(arg string)); ok {
		return func(params ...Arg) {
			fun(params[0].(string))
		}
	}
	return nil
}

func InvokeHttpHandlerFunc(h Handler) Invoker {
	if fun, ok := h.(func(http.ResponseWriter, *http.Request)); ok {
		return func(params ...Arg) {
			fun(params[0].(http.ResponseWriter),
				params[1].(*http.Request))
		}
	}
	return nil
}
