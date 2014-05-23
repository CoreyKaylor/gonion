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
			g.Handle(get_index2)
			runtime := g.buildRuntime()
			route := runtime.routeFor("/index2")
			recorder := httptest.NewRecorder()
			route.Handler(recorder, new(http.Request))
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
		wrapperHandler := func(rw http.ResponseWriter, r *http.Request, nextHandler NextHandler) {
			nextHandler(&wrapper{rw}, r)
		}
		g.Use(wrapperHandler)
		g.Use(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte("no-wrap"))
		})
		g.Use(wrapperHandler)
		g.Use(wrapperHandler)

		g.Handle(get_index2)
		runtime := g.buildRuntime()
		route := runtime.routeFor("/index2")
		recorder := httptest.NewRecorder()
		route.Handler(recorder, new(http.Request))

		response := recorder.Body.String()
		So(response, ShouldEqual, "wrapperno-wrapwrapperwrapperwrapperwrapperwrapperwrapperwrapperSuccess!")
	})
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

func BenchmarkSimpleInvocation(b *testing.B) {
	g := New()
	g.Handle(get_index2)
	runtime := g.buildRuntime()
	route := runtime.routeFor("/index2")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		recorder := httptest.NewRecorder()
		route.Handler(recorder, new(http.Request))
		response := string(recorder.Body.Bytes())
		if response != "Success!" {
			b.FailNow()
		}
	}
}
