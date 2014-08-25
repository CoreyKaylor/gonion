package gonion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppInitialization(t *testing.T) {
	g := New()
	g.Handle("GET", "/index2", http.HandlerFunc(getIndex2))
	routes := g.BuildRoutes()
	route := routes.routeFor("*", "/index2")
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, new(http.Request))
	response := recorder.Body.String()
	assert.Equal(t, response, "Success!")
}

type wrapper struct {
	http.ResponseWriter
}

func (rw *wrapper) Write(b []byte) (int, error) {
	bytes, _ := rw.ResponseWriter.Write([]byte("wrapper"))
	moreBytes, _ := rw.ResponseWriter.Write(b)
	return bytes + moreBytes, nil
}

func TestWrappingChainWithNewWriter(t *testing.T) {
	g := New()
	wrapperHandler := func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			inner.ServeHTTP(&wrapper{rw}, r)
		})
	}
	g.Use().ChainLink(wrapperHandler)
	g.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("no-wrap"))
	})
	g.Use().ChainLink(wrapperHandler)
	g.Use().ChainLink(wrapperHandler)
	g.Handle("GET", "/index2", http.HandlerFunc(getIndex2))
	routes := g.BuildRoutes()
	route := routes.routeFor("*", "/index2")
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, new(http.Request))

	response := recorder.Body.String()
	assert.Equal(t, response, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
}

func oneOfEachRoute() *Composer {
	g := New()
	addRoute := func(method string) {
		g.Handle(method, "/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(method))
		}))
	}
	addRoute("POST")
	addRoute("GET")
	addRoute("PUT")
	addRoute("PATCH")
	addRoute("DELETE")
	return g
}

func assertRouteConstraintResponse(t *testing.T, c *Composer, method string, expectedResponse string) {
	route := c.BuildRoutes().routeFor(method, "/")
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, nil)
	assert.Equal(t, recorder.Body.String(), expectedResponse)
}

func TestConstrainingMiddleware_GetAppliesGetOnlyMiddleware(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().Get().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("GETONLY"))
	})
	assertRouteConstraintResponse(t, g, "GET", "GETONLYGET")
	assertRouteConstraintResponse(t, g, "POST", "POST")
}

func TestConstrainingMiddleware_PutAppliesPutOnlyMiddleware(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().Put().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("PUTONLY"))
	})
	assertRouteConstraintResponse(t, g, "GET", "GET")
	assertRouteConstraintResponse(t, g, "PUT", "PUTONLYPUT")
}

func TestConstrainingMiddleware_PatchAppliesPatchOnlyMiddleware(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().Patch().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("PATCHONLY"))
	})
	assertRouteConstraintResponse(t, g, "GET", "GET")
	assertRouteConstraintResponse(t, g, "PATCH", "PATCHONLYPATCH")
}

func TestConstrainingMiddleware_DeleteAppliesDeleteOnlyMiddleware(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().Delete().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("DELETEONLY"))
	})
	assertRouteConstraintResponse(t, g, "GET", "GET")
	assertRouteConstraintResponse(t, g, "DELETE", "DELETEONLYDELETE")
}

func TestConstrainingMiddleware_PostAppliesPostOnlyMiddleware(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().Post().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("POSTONLY"))
	})
	assertRouteConstraintResponse(t, g, "POST", "POSTONLYPOST")
	assertRouteConstraintResponse(t, g, "GET", "GET")
}

func TestConstrainingMiddleware_WithArbitraryCondition(t *testing.T) {
	g := oneOfEachRoute()
	g.Only().When(func() bool {
		return false
	}).Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("nada"))
	})
	assertRouteConstraintResponse(t, g, "GET", "GET")
	assertRouteConstraintResponse(t, g, "DELETE", "DELETE")
}

func (routes Routes) routeFor(method string, pattern string) *Route {
	for _, r := range routes {
		if r.Pattern == pattern && (method == "*" || method == r.Method) {
			return r
		}
	}
	return nil
}

func getIndex2(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Success!"))
}
