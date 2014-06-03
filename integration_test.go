package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegratingAllThePieces(t *testing.T) {
	Convey("When using all API functions", t, func() {
		g := New()
		g.UseFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("usefunc->"))
		})
		g.UseMiddlewareConstructor(timeoutHandler)
		g.Use(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("handlerfunc->"))
		}))
		g.Get("/", Index)
		g.Sub("/api", func(api *App) {
			api.UseFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("api-key->"))
			})
			api.Get("/users/:id", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("subSuccess!"))
			})
			api.Sub("/admin", func(admin *App) {
				admin.UseFunc(func(rw http.ResponseWriter, r *http.Request) {
					rw.Write([]byte("isadmin->"))
				})
				admin.Get("/super-important", func(rw http.ResponseWriter, r *http.Request) {
					rw.Write([]byte("importantstuff!"))
				})
			})
		})
		runtime := g.buildRuntime()
		recorder := httptest.NewRecorder()
		Convey("Routes defined for the root path don't inherit sub-route middlware", func() {
			route := runtime.routeFor("/")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->Success!")
		})
		Convey("Sub-routes do inherit root middlware in addition to its own", func() {
			route := runtime.routeFor("/api/users/:id")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->api-key->subSuccess!")
		})
		Convey("Sub-Sub-routes do inherit root middlware in addition to its own", func() {
			route := runtime.routeFor("/api/admin/super-important")
			route.Handler.ServeHTTP(recorder, new(http.Request))
			So(recorder.Body.String(), ShouldEqual, "usefunc->timeout->handlerfunc->api-key->isadmin->importantstuff!")
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
