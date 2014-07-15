package gonion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAppInitialization(t *testing.T) {
	Convey("When building the routes", t, func() {
		Convey("The route handler is built", func() {
			g := New()
			g.Handle("GET", "/index2", get_index2)
			routes := g.BuildRoutes()
			route := routes.routeFor("/index2")
			recorder := httptest.NewRecorder()
			route.Handler.ServeHTTP(recorder, new(http.Request))
			response := recorder.Body.String()
			So(response, ShouldEqual, "Success!")
		})
	})
}

type wrapper struct {
	http.ResponseWriter
}

func (rw *wrapper) Write(b []byte) (int, error) {
	bytes, _ := rw.ResponseWriter.Write([]byte("wrapper"))
	moreBytes, _ := rw.ResponseWriter.Write(b)
	return bytes + moreBytes, nil
}

func TestChainWrappingSemantics(t *testing.T) {
	Convey("When middleware wraps the writer it should use the wrapped writer for chain", t, func() {
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
		g.Handle("GET", "/index2", get_index2)
		routes := g.BuildRoutes()
		route := routes.routeFor("/index2")
		recorder := httptest.NewRecorder()
		route.Handler.ServeHTTP(recorder, new(http.Request))

		response := recorder.Body.String()
		So(response, ShouldEqual, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
	})
}

func (routes Routes) routeFor(pattern string) *Route {
	for _, r := range routes {
		if r.Pattern == pattern {
			return r
		}
	}
	return nil
}

func get_index2(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Success!"))
}
