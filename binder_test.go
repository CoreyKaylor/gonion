package gonion

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestBinding(t *testing.T) {
	Convey("When binding to a request", t, func() {
		binder := NewBinder()
		ctx := &BindingContext{nil, nil, nil, "name", reflect.TypeOf("")}

		Convey("Query string parameter can be resolved by the form binder", func() {
			ctx.req, _ = http.NewRequest("GET", "http://something.com?name=foo", nil)
			arg, ok := binder.Bind(ctx)
			So(ok, ShouldBeTrue)
			So(arg, ShouldEqual, "foo")
		})

		Convey("Form POST data can be resolved by the form binder", func() {
			ctx.req, _ = http.NewRequest("POST", "http://test.com",
				strings.NewReader("name=foo"))
			ctx.req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
			arg, ok := binder.Bind(ctx)
			So(ok, ShouldBeTrue)
			So(arg, ShouldEqual, "foo")
		})

		Convey("Cookie value can be bound", func() {
			req, _ := http.NewRequest("GET", "http://test.com", nil)
			req.AddCookie(&http.Cookie{Name: "name", Value: "foo"})
			ctx.req = req
			arg, ok := binder.Bind(ctx)
			So(ok, ShouldBeTrue)
			So(arg, ShouldEqual, "foo")
		})
	})
}
