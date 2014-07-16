# Gonion

Gonion is a library for composing your routes and middleware. It is NOT a full-fledged framework, but rather allows you to compose
things in a similar way that a framework might provide. Even though "idiomatic" go seems to encourage repeating yourself over
and over again Gonion stays idiomatic without the repetition and loss of expressiveness.

Gonion has 0 reflection dependencies and adds 0 runtime alloc's to your routes and middleware.

Gonion is extremely fast!

~~~
Benchmark_Simple			20000000		124 ns/op			14 B/op			0 allocs/op
Benchmark_Middleware	10000000		180 ns/op			14 B/op			0 allocs/op
~~~

## Getting Started

Assuming you have installed gonion via `go get github.com/CoreyKaylor/gonion`

~~~ go
package main

import (
  "net/http"

  "github.com/CoreyKaylor/gonion"
	"github.com/julienschmidt/httprouter"
)

func main() {
	g := gonion.New()
	g.Get("/hello", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("Hello World!"))
	}))
	routes := g.BuildRoutes()
	router := httprouter.New() //use router of your choice
	for _, route := range routes {
		router.Handle(route.Method, route.Pattern, func(rw http.ResponseWriter, r *http.Request, m map[string]string) {
			route.Handler.ServeHTTP(rw, r)
		})
	}
	http.ListenAndServe(":3000", router)
}
~~~

Then 
`go run server.go`

You should now be able to open your browser to http://localhost:3000/

## Middleware

Gonion handlers and middleware all take the form of the standard 'net/http' http.Handler

~~~ go
g.Use(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
	rw.Write([]byte("middleware"))
}))
~~~

When you need to wrap the downstream chain, rather than adding new signatures beyond the http.Handler gonion uses the 
middleware constructor method or in gonion terms the 'ChainLink'. 

~~~ go
type ChainLink func(http.Handler) http.Handler
~~~

An example usage of a ChainLink

~~~ go
func wrappingHandler(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		//useless example, but this would be a good place to wrap the writer for something like gzip compression
		rw.Write([]byte("wrapping"))
		inner.ServeHTTP(rw, r)
	})
}

//and it's corresponding registration
g.Use().ChainLink(wrappingHandler)
~~~

Often it's useful to only apply middleware for 'POST' only routes. This removes the needless runtime checks for whether
the current requests method is truly POST.

~~~ go
g.Only().Post().Use().ChainLink(wrappingHandler)
~~~

Typically middleware applies only to routes at a particular path.

~~~ go
g.Sub("/api", func(api *Composer){
	//middleware will only apply to routes with the prefix of /api
	api.Use().ChainLink(apiKeyHandler)
	api.Get("/users", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		rw.Write([]byte("Full route is /api/users"))
	}))
})
~~~
