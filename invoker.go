package gonion

import (
	"net/http"
	"reflect"
)

type Arg interface{}

type Invoker func(...Arg)

type InvokerLocator func(Handler) Invoker

type InvocationRegistry struct {
	Binder   *Binder
	Locators []InvokerLocator
}

func NewInvocationRegistry() *InvocationRegistry {
	registry := InvocationRegistry{NewBinder(), make([]InvokerLocator, 0, 10)}
	registry.AddLocator(InvokeAnythingWithReflectionFallback)
	registry.AddLocator(InvokeStringArgumentNoReturn)
	registry.AddLocator(InvokeHttpHandlerFunc)
	return &registry
}

func (r *InvocationRegistry) AddLocator(l InvokerLocator) {
	r.Locators = append(r.Locators, l)
}

type ParameterInfo struct {
	Name string
	Type reflect.Type
}

func getParameters(h Handler) []*ParameterInfo {
	value := reflect.ValueOf(h)
	handlerType := value.Type()
	numArgs := handlerType.NumIn()
	args := make([]*ParameterInfo, numArgs)
	for i := 0; i < numArgs; i++ {
		arg := handlerType.In(i)
		parameter := &ParameterInfo{arg.Name(), arg}
		args[i] = parameter
	}
	return args
}

type ChainFunc func(http.ResponseWriter, *http.Request, map[string]string)

func (r *InvocationRegistry) GetInvoker(h Handler) ChainFunc {
	invoker := r.findInvoker(h)
	parameters := getParameters(h)
	x := func(rw http.ResponseWriter, req *http.Request, additional map[string]string) {
		args := make([]Arg, 0, 10)
		for i := 0; i < len(parameters); i++ {
			ctx := &BindingContext{rw, req, additional, parameters[i].Name, parameters[i].Type}
			if arg, ok := r.Binder.Bind(ctx); ok {
				args = append(args, arg)
			} else {
				panic("gonion: Unable to bind parameter")
			}
		}
		invoker(args...)
	}
	return x
}

func (r *InvocationRegistry) findInvoker(h Handler) Invoker {
	for i := len(r.Locators) - 1; i >= 0; i-- {
		if invoker := r.Locators[i](h); invoker != nil {
			return invoker
		}
	}

	panic("gonion: no invoker found for Handler")
}

func (r *InvocationRegistry) buildInvocationChain(handlers ...Handler) ChainFunc {
	firstFunc := r.GetInvoker(handlers[len(handlers)-1])
	chain := func(rw http.ResponseWriter, req *http.Request, additional map[string]string) {
		firstFunc(rw, req, additional)
	}
	for i := len(handlers) - 2; i >= 0; i-- {
		current := r.GetInvoker(handlers[i])
		currentChain := chain
		chain = func(rw http.ResponseWriter, req *http.Request, additional map[string]string) {
			current(rw, req, additional)
			currentChain(rw, req, additional)
		}
	}
	return chain
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

func InvokeNoArgumentsNoReturn(h Handler) Invoker {
	if fun, ok := h.(func()); ok {
		return func(params ...Arg) {
			fun()
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
