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
			response := string(recorder.Body.Bytes())
			So(response, ShouldEqual, "Success!")
		})
	})
}

func (runtime *Runtime) routeFor(pattern string) *RuntimeRoute {
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
