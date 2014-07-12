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
		wrapperHandler := func(rw http.ResponseWriter, r *http.Request, next http.Handler) {
			next.ServeHTTP(&wrapper{rw}, r)
		}
		g.Use().WrappingFunc(wrapperHandler)
		g.Use().Func(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("no-wrap"))
		})
		g.Use().WrappingFunc(wrapperHandler)
		g.Use().WrappingFunc(wrapperHandler)
		g.Handle("GET", "/index2", get_index2)
		routes := g.BuildRoutes()
		route := routes.routeFor("/index2")
		recorder := httptest.NewRecorder()
		route.Handler.ServeHTTP(recorder, new(http.Request))

		response := recorder.Body.String()
		So(response, ShouldEqual, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
	})
}

func TestContextualHandlers(t *testing.T) {
	g := New()
	g.Use().ContextHandler(MyContextFunc((*MyContext).Middle)).CreateContext(func() interface{} {
		return &MyContext{}
	})
	g.GetC("/", MyContextFunc((*MyContext).Get))
	routes := g.BuildRoutes()
	route := routes.routeFor("/")
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, new(http.Request))

	response := recorder.Body.String()
	if response != "middlecontextgetcontext" {
		t.FailNow()
	}
}

type MyContextFunc func(*MyContext, http.ResponseWriter, *http.Request)

func (m MyContextFunc) ServeHTTP(i interface{}, rw http.ResponseWriter, r *http.Request) {
	m(i.(*MyContext), rw, r)
}

type MyContext struct {
}

func (c *MyContext) Middle(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("middlecontext"))
}

func (c *MyContext) Get(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("getcontext"))
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

func get_index3(rw http.ResponseWriter, r *http.Request) {
}
