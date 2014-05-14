package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMiddlewareRegistry(t *testing.T) {
	Convey("When registering middleware for all routes", t, func() {
		m := NewMiddlewareRegistry()
		m.AppliesToAllRoutes("test", func() {
		})
		Convey("It should return true always", func() {
			routes := make([]*Route, 0, 10)
			for i := 0; i < 5; i++ {
				routes = append(routes, &Route{})
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
