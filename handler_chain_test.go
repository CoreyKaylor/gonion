package gonion

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhenMiddlewareAppliesToAllRoutes(t *testing.T) {
	m := newMiddlewareRegistry()
	m.appliesToAllRoutes(func(http.Handler) http.Handler {
		return nil
	})
	routes := make([]*RouteModel, 0, 10)
	for i := 0; i < 5; i++ {
		routes = append(routes, &RouteModel{})
	}
	for i := 0; i < 5; i++ {
		middlewareLength := len(m.middlewareFor(routes[i]))
		assert.Equal(t, 1, middlewareLength)
	}
}
