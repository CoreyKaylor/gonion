package gonion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIntegratingAllThePieces(t *testing.T) {
	Convey("When using all API functions", t, func() {
		g := New()
		g.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("usefunc->"))
		})
		g.Use().ConstructorFunc(timeoutHandler)
		g.Use().HandlerFunc(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("handlerfunc->"))
		}))
		g.Get("/", Index)
		g.Sub("/api", func(api *Composer) {
			api.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("api-key->"))
			})
			api.Get("/users/:id", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("subSuccess!"))
			})
			api.Sub("/admin", func(admin *Composer) {
				admin.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
					rw.Write([]byte("isadmin->"))
				})
				admin.Get("/super-important", func(rw http.ResponseWriter, r *http.Request) {
					rw.Write([]byte("importantstuff!"))
				})
			})
		})
		routes := g.BuildRoutes()
		recorder := httptest.NewRecorder()
		Convey("Routes defined for the root path don't inherit sub-route middlware", func() {
			route := routes.routeFor("/")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->Success!")
		})
		Convey("Sub-routes do inherit root middlware in addition to its own", func() {
			route := routes.routeFor("/api/users/:id")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->api-key->subSuccess!")
		})
		Convey("Sub-Sub-routes do inherit root middlware in addition to its own", func() {
			route := routes.routeFor("/api/admin/super-important")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->api-key->isadmin->importantstuff!")
		})
	})
}

func TestEndToEndWithRouter(t *testing.T) {
	Convey("When using a routing package", t, func() {
		g := New()
		g.Get("/hello", func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("Success!"))
		})
		routes := g.BuildRoutes()
		router := httprouter.New()
		for _, route := range routes {
			router.Handle(route.Method, route.Pattern, func(rw http.ResponseWriter, r *http.Request, m map[string]string) {
				route.Handler.ServeHTTP(rw, r)
			})
		}
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/hello", nil)
		router.ServeHTTP(recorder, request)
		Convey("Everything should flow through to the registered handler", func() {
			So(recorder.Body.String(), ShouldEqual, "Success!")
		})
	})
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("timeout->"))
		h.ServeHTTP(rw, r)
	})
}

func Index(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Success!"))
}
