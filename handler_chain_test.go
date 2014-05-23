package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerInvocation(t *testing.T) {
	Convey("When invoking handlers", t, func() {
		registry := NewChainBuilder()
		rw := httptest.NewRecorder()
		req := new(http.Request)
		Convey("For a func matching http HandlerFunc", func() {
			var handled bool
			handler := func(rw http.ResponseWriter, req *http.Request) {
				handled = true
			}
			invoker, _ := registry.GetInvoker(handler, nil)
			context := Context{rw, req}
			invoker(context)
			So(handled, ShouldBeTrue)
		})
	})
}

func TestBuildingChain(t *testing.T) {
	Convey("When building a chain of handlers", t, func() {
		var result string
		registry := NewChainBuilder()
		chain := registry.build(func(rw http.ResponseWriter, req *http.Request) {
			result += "blah"
		}, func(rw http.ResponseWriter, req *http.Request) {
			result += "blah2"
		})
		Convey("The chain will build a composed function executing all handlers in order", func() {
			chain(httptest.NewRecorder(), new(http.Request))
			So(result, ShouldEqual, "blahblah2")
		})
	})
}

func TestMiddlewareRegistry(t *testing.T) {
	Convey("When registering middleware for all routes", t, func() {
		m := NewMiddlewareRegistry()
		m.AppliesToAllRoutes("test", func() {
		})
		Convey("It should return true always", func() {
			routes := make([]*RouteModel, 0, 10)
			for i := 0; i < 5; i++ {
				routes = append(routes, &RouteModel{})
			}
			for i := 0; i < 5; i++ {
				warez := m.MiddlewareFor(routes[i])
				if len(warez) != 1 {
					t.FailNow()
				}
			}
		})
	})
}
