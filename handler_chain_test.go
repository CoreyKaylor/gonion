package gonion

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMiddlewareRegistry(t *testing.T) {
	Convey("When registering middleware for all routes", t, func() {
		m := NewMiddlewareRegistry()
		m.AppliesToAllRoutes(func(http.Handler) http.Handler {
			return nil
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
