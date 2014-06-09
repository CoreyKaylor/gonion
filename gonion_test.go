package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppInitialization(t *testing.T) {
	Convey("When building the runtime", t, func() {
		Convey("The route handler is built", func() {
			g := New()
			g.Handle("GET", "/index2", get_index2)
			runtime := g.buildRuntime()
			route := runtime.routeFor("/index2")
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
		g.UseWrappingHandler(wrapperHandler)
		g.UseFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("no-wrap"))
		})
		g.UseWrappingHandler(wrapperHandler)
		g.UseWrappingHandler(wrapperHandler)

		g.Handle("GET", "/index2", get_index2)
		runtime := g.buildRuntime()
		route := runtime.routeFor("/index2")
		recorder := httptest.NewRecorder()
		route.Handler.ServeHTTP(recorder, new(http.Request))

		response := recorder.Body.String()
		So(response, ShouldEqual, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
	})
}

func TestContextualHandlers(t *testing.T) {
	g := New()
	g.CreateContext(func() interface{} {
		return &MyContext{}
	})
	g.UseContextualHandler(MyM((*MyContext).Middle))
	g.GetC("/", MyM((*MyContext).Get))
	runtime := g.buildRuntime()
	route := runtime.routeFor("/")
	recorder := httptest.NewRecorder()
	route.Handler.ServeHTTP(recorder, new(http.Request))

	response := recorder.Body.String()
	if response != "middlecontextgetcontext" {
		t.FailNow()
	}
}

type MyM func(*MyContext, http.ResponseWriter, *http.Request)

func (m MyM) Wrap() func(interface{}, http.ResponseWriter, *http.Request) {
	return func(i interface{}, rw http.ResponseWriter, req *http.Request) {
		m(i.(*MyContext), rw, req)
	}
}

type MyContext struct {
}

func (c *MyContext) Middle(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("middlecontext"))
}

func (c *MyContext) Get(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("getcontext"))
}

func (runtime *Runtime) routeFor(pattern string) *Route {
	for _, r := range runtime.Routes {
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

func BenchmarkSimpleInvocation(b *testing.B) {
	g := New()
	g.UseWrappingHandler(func(rw http.ResponseWriter, r *http.Request, handler http.Handler) {
		handler.ServeHTTP(rw, r)
	})
	g.Handle("GET", "/index3", get_index3)
	runtime := g.buildRuntime()
	route := runtime.routeFor("/index3")
	b.ReportAllocs()
	b.ResetTimer()
	recorder := httptest.NewRecorder()
	request := new(http.Request)
	for i := 0; i < b.N; i++ {
		route.Handler.ServeHTTP(recorder, request)
	}
}
