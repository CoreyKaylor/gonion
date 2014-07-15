package gonion

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func Benchmark_Simple(b *testing.B) {
	g := New()
	g.Get("/simple", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello")
	}))
	routes := g.BuildRoutes()
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Pattern, func(rw http.ResponseWriter, r *http.Request, m map[string]string) {
			route.Handler.ServeHTTP(rw, r)
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/simple", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(recorder, request)
	}
}

func Benchmark_Middleware(b *testing.B) {
	g := New()
	g.Sub("/a", func(a *Composer) {
		a.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		})
		a.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
		})
		a.Sub("/c", func(c *Composer) {
			c.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
			})
			c.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
			})
			c.Get("/action", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(rw, "hello")
			}))
		})
	})
	routes := g.BuildRoutes()
	router := httprouter.New()
	for _, route := range routes {
		router.Handle(route.Method, route.Pattern, func(rw http.ResponseWriter, r *http.Request, m map[string]string) {
			route.Handler.ServeHTTP(rw, r)
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/a/c/action", nil)
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(recorder, request)
	}
}
