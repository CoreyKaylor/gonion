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
			g.Handle("GET", "/index2", http.HandlerFunc(getIndex2))
			routes := g.BuildRoutes()
			route := routes.routeFor("*", "/index2")
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
		g.Handle("GET", "/index2", http.HandlerFunc(getIndex2))
		routes := g.BuildRoutes()
		route := routes.routeFor("*", "/index2")
		recorder := httptest.NewRecorder()
		route.Handler.ServeHTTP(recorder, new(http.Request))

		response := recorder.Body.String()
		So(response, ShouldEqual, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
	})
}

func TestMiddlewareConstraints(t *testing.T) {
	Convey("When constraining middleware", t, func() {
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
		Convey("to POST only", func() {
			g.Only().Post().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("POSTONLY"))
			})
			Convey("GET route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GET")
			})
			Convey("POST route should apply middleware", func() {
				route := g.BuildRoutes().routeFor("POST", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "POSTONLYPOST")
			})
		})
		Convey("to GET only", func() {
			g.Only().Get().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("GETONLY"))
			})
			Convey("GET route should apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GETONLYGET")
			})
			Convey("POST route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("POST", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "POST")
			})
		})
		Convey("to PUT only", func() {
			g.Only().Put().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("PUTONLY"))
			})
			Convey("GET route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GET")
			})
			Convey("PUT route should apply middleware", func() {
				route := g.BuildRoutes().routeFor("PUT", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "PUTONLYPUT")
			})
		})
		Convey("to PATCH only", func() {
			g.Only().Patch().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("PATCHONLY"))
			})
			Convey("GET route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GET")
			})
			Convey("PATCH route should apply middleware", func() {
				route := g.BuildRoutes().routeFor("PATCH", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "PATCHONLYPATCH")
			})
		})
		Convey("to DELETE only", func() {
			g.Only().Delete().Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("DELETEONLY"))
			})
			Convey("GET route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GET")
			})
			Convey("DELETE route should apply middleware", func() {
				route := g.BuildRoutes().routeFor("DELETE", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "DELETEONLYDELETE")
			})
		})
		Convey("to arbitrary condition", func() {
			g.Only().When(func() bool {
				return false
			}).Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("nada"))
			})
			Convey("GET route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("GET", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "GET")
			})
			Convey("DELETE route should not apply middleware", func() {
				route := g.BuildRoutes().routeFor("DELETE", "/")
				recorder := httptest.NewRecorder()
				route.Handler.ServeHTTP(recorder, nil)
				So(recorder.Body.String(), ShouldEqual, "DELETE")
			})
		})

	})
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
