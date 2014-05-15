package gonion

import (
	"net/http"
	"reflect"
)

type Binder struct {
	sources []BindRequest
}

type BindingContext struct {
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
	}
	binder := &Binder{sources}
	return binder
}

func (b *Binder) Bind(ctx *BindingContext) (Arg, bool) {
	for _, br := range b.sources {
		if val, ok := br(ctx); ok {
			return val, ok
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

// TODO: JSON-to-struct bind request
