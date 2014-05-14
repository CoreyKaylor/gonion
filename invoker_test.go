package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerInvocation(t *testing.T) {
	Convey("When invoking handlers", t, func() {
		registry := NewInvocationRegistry()
		Convey("For a func with no input or output", func() {
			var handled bool
			handler := func() {
				handled = true
			}

			invoker := registry.GetInvoker(handler)
			invoker()
			So(handled, ShouldBeTrue)
		})
		Convey("For a func with single arity", func() {
			var x string
			handler := func(name string) {
				x = name
			}
			invoker := registry.GetInvoker(handler)
			invoker("Corey")
			So(x, ShouldEqual, "Corey")
		})
		Convey("For a func matching http HandlerFunc", func() {
			var handled bool
			handler := func(rw http.ResponseWriter, req *http.Request) {
				handled = true
			}
			invoker := registry.GetInvoker(handler)
			invoker(httptest.NewRecorder(), new(http.Request))
			So(handled, ShouldBeTrue)
		})
		Convey("Falls back to reflection for unknown funcs", func() {
			invoker := registry.GetInvoker(func(a, lot, of, args int) {})
			So(invoker, ShouldNotBeNil)
		})
	})
}

func TestBuiltLocators(t *testing.T) {
	Convey("When locating invokers that match a built-in locator", t, func() {
		registry := NewInvocationRegistry()
		registry.Locators = registry.Locators[1:]
		Convey("Uses the right locator for single string argument", func() {
			invoker := registry.GetInvoker(func(foo string) {})
			So(invoker, ShouldNotBeNil)
		})

		Convey("Uses the right locator for http HandlerFunc", func() {
			invoker := registry.GetInvoker(func(rw http.ResponseWriter, r *http.Request) {})
			So(invoker, ShouldNotBeNil)
		})
	})
}