package gonion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestIntegratingAllThePieces(t *testing.T) {
	g := New()
	g.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("usefunc->"))
	})
	g.Use().ChainLink(timeoutHandler)
	g.Use().Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("handlerfunc->"))
	}))
	g.Get("/", http.HandlerFunc(Index))
	g.Sub("/api", func(api *Composer) {
		api.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("api-key->"))
		})
		api.Get("/users/:id", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("subSuccess!"))
		}))
		api.Sub("/admin", func(admin *Composer) {
			admin.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("isadmin->"))
			})
			admin.Get("/super-important", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("importantstuff!"))
			}))
			admin.Post("/super-important", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("importantstuff!"))
			}))
			admin.Put("/super-important", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("importantstuff!"))
			}))
			admin.Patch("/super-important", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("importantstuff!"))
			}))
			admin.Delete("/super-important", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("importantstuff!"))
			}))
		})
	})
	routes := g.BuildRoutes()
	recorder := httptest.NewRecorder()
	route := routes.routeFor("*", "/")
	route.Handler.ServeHTTP(recorder, new(http.Request))
	assert.Equal(t, recorder.Body.String(), "usefunc->timeout->handlerfunc->Success!")

	recorder = httptest.NewRecorder()
	route = routes.routeFor("*", "/api/users/:id")
	route.Handler.ServeHTTP(recorder, new(http.Request))
	assert.Equal(t, recorder.Body.String(), "usefunc->timeout->handlerfunc->api-key->subSuccess!")

	recorder = httptest.NewRecorder()
	route = routes.routeFor("*", "/api/admin/super-important")
	route.Handler.ServeHTTP(recorder, new(http.Request))
	assert.Equal(t, recorder.Body.String(), "usefunc->timeout->handlerfunc->api-key->isadmin->importantstuff!")
}

func TestEndToEndWithRouter(t *testing.T) {
	g := New()
	g.Get("/hello", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Success!"))
	}))
	router := httprouter.New()
	g.EachRoute(func(route *Route) {
		router.Handle(route.Method, route.Pattern, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
			route.Handler.ServeHTTP(rw, r)
		})
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/hello", nil)
	router.ServeHTTP(recorder, request)
	assert.Equal(t, recorder.Body.String(), "Success!")
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
