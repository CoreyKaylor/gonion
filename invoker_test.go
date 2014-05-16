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

			invoker := registry.findInvoker(handler)
			invoker()
			So(handled, ShouldBeTrue)
		})
		Convey("For a func with single arity", func() {
			var x string
			handler := func(name string) {
				x = name
			}
			invoker := registry.findInvoker(handler)
			invoker("Corey")
			So(x, ShouldEqual, "Corey")
		})
		Convey("For a func matching http HandlerFunc", func() {
			var handled bool
			handler := func(rw http.ResponseWriter, req *http.Request) {
				handled = true
			}
			invoker := registry.findInvoker(handler)
			invoker(httptest.NewRecorder(), new(http.Request))
			So(handled, ShouldBeTrue)
		})
		Convey("Falls back to reflection for unknown funcs", func() {
			invoker := registry.findInvoker(func(a, lot, of, args int) {})
			So(invoker, ShouldNotBeNil)
		})
	})
}

func TestBuiltLocators(t *testing.T) {
	Convey("When locating invokers that match a built-in locator", t, func() {
		registry := NewInvocationRegistry()
		registry.Locators = registry.Locators[1:]
		Convey("Uses the right locator for single string argument", func() {
			invoker := registry.findInvoker(func(foo string) {})
			So(invoker, ShouldNotBeNil)
		})

		Convey("Uses the right locator for http HandlerFunc", func() {
			invoker := registry.findInvoker(func(rw http.ResponseWriter, r *http.Request) {})
			So(invoker, ShouldNotBeNil)
		})
	})
}

func TestBuildingChain(t *testing.T) {
	Convey("When building a chain of handlers", t, func() {
		var result string
		registry := NewInvocationRegistry()
		chain := registry.buildInvocationChain(func() {
			result += "blah"
		}, func() {
			result += "blah2"
		})
		Convey("The chain will build a composed function executing all handlers in order", func() {
			chain(httptest.NewRecorder(), new(http.Request), make(map[string]string))
			So(result, ShouldEqual, "blahblah2")
		})
	})
}

type TestInput string

func TestBindingArguments(t *testing.T) {
	Convey("When invoking a handler containing bindable arguments", t, func() {
		registry := NewInvocationRegistry()
		var x TestInput
		handler := func(name TestInput) {
			x = name
		}

		invoker := registry.GetInvoker(handler)
		req, _ := http.NewRequest("GET", "http://test.com?TestInput=foo", nil)
		invoker(httptest.NewRecorder(), req, make(map[string]string))
		So(x, ShouldEqual, "foo")
	})
}
