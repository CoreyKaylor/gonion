package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func (r *RouteRegistry) findByPattern(pattern string) *RouteModel {
	for _, route := range r.Routes {
		if route.Pattern == pattern {
			return route
		}
	}
	return nil
}

func TestRouteByFunc(t *testing.T) {
	Convey("When adding routes by func", t, func() {
		routes := NewRouteRegistry()
		Convey("When adding a basic index route", func() {
			routes.AddFunc(get_index)
			route := routes.findByPattern("/")
			Convey("The route pattern is derived from function name", func() {
				So(route, ShouldNotBeNil)
			})
			Convey("The HTTP METHOD is derived from function name prefix", func() {
				So(route.Method, ShouldEqual, "GET")
			})
		})
		Convey("When adding a route for func with multiple name parts", func() {
			routes.AddFunc(get_something_great_going_on_here)
			Convey("The '_'s are replaced with '/'", func() {
				So(routes.findByPattern("/something/great/going/on/here"), ShouldNotBeNil)
			})
		})
		Convey("When trying to add a func route with incorrect arguments", func() {
			Convey("It should panic", func() {
				So(func() { routes.AddFunc("hello") }, ShouldPanic)
			})
		})
	})
}

func get_index() {
}

func get_something_great_going_on_here() {
}
