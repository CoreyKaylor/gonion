package contextual

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.google.com/p/go.net/context"
)

func TestContextualHandlers(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := Handlers(
		ChainLinkFunc(ContextualOneChainLink),
		Last(HandlerFunc(ContextualTwo)),
	)
	handler.ServeHTTP(recorder, new(http.Request))
	if recorder.Body.String() != "hello" {
		t.Fail()
	}
}

var key = context.NewKey("test")

func ContextualOne(ctx context.Context, rw http.ResponseWriter, r *http.Request, next Handler) {
	next.ServeHTTP(context.WithValue(ctx, key, "hello"), rw, r)
}

func ContextualOneChainLink(inner Handler) Handler {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
		ContextualOne(ctx, rw, r, inner)
	})
}

func ContextualTwo(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	message := ctx.Value(key).(string)
	rw.Write([]byte(message))
}
