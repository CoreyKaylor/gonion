package gonion

import (
	"net/http"
	"reflect"
)

type Binder struct {
	sources []BindRequest
}

type BindingContext struct {
	rw             http.ResponseWriter
	req            *http.Request
	additionalArgs map[string]string
	targetName     string
	targetType     reflect.Type
}

type BindRequest func(ctx *BindingContext) (Arg, bool)

func NewBinder() *Binder {
	sources := []BindRequest{
		formValueToPrimitiveBindRequest,
		argsBindRequest,
		cookieValueToPrimitiveBindRequest,
		handlerFuncBindRequest,
	}
	binder := &Binder{sources}
	return binder
}

func (b *Binder) Bind(ctx *BindingContext) (Arg, bool) {
	for _, br := range b.sources {
		if val, ok := br(ctx); ok {
			return reflect.ValueOf(val).Convert(ctx.targetType).Interface(), ok
		}
	}
	return nil, false
}

func formValueToPrimitiveBindRequest(ctx *BindingContext) (Arg, bool) {
	val := ctx.req.FormValue(ctx.targetName)
	if len(val) == 0 {
		return nil, false
	}
	return val, true
}

func argsBindRequest(ctx *BindingContext) (val Arg, ok bool) {
	val, ok = ctx.additionalArgs[ctx.targetName]
	return
}

func cookieValueToPrimitiveBindRequest(ctx *BindingContext) (Arg, bool) {
	c, err := ctx.req.Cookie(ctx.targetName)
	if err != nil {
		return nil, false
	}
	return c.Value, true
}

func handlerFuncBindRequest(ctx *BindingContext) (Arg, bool) {
	if reflect.ValueOf(ctx.rw).Type().AssignableTo(ctx.targetType) {
		return ctx.rw, true
	}
	if reflect.ValueOf(ctx.req).Type() == ctx.targetType {
		return ctx.req, true
	}
	return nil, false
}

// TODO: JSON-to-struct bind request
