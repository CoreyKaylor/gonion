package recovery

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var panicHandler http.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	panic("oh no!")
})

var noPanicHandler http.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
})

func TestRecoveryWithStackTrace(t *testing.T) {
	Convey("When using recovery with stacktrace", t, func() {
		Convey("And a panic occurs", func() {
			Convey("Recovery should include the stacktrace", func() {
				recovery := NewRecoveryWithStackTrace()
				recorder := httptest.NewRecorder()
				recovery.ServeHTTP(recorder, new(http.Request), panicHandler)
				body := recorder.Body.String()
				Convey("With a status code of InternalServerError", func() {
					So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
				})
				Convey("body should contain the stacktrace", func() {
					containsStackTraceClass := strings.Contains(body, "class=\"stacktrace\"")
					So(containsStackTraceClass, ShouldBeTrue)
				})
				Convey("body should contain the panic error message", func() {
					containsPanicMessage := strings.Contains(body, "oh no!")
					So(containsPanicMessage, ShouldBeTrue)
				})
			})
		})
		Convey("No panic occurs", func() {
			recovery := NewRecoveryWithStackTrace()
			recorder := httptest.NewRecorder()
			recovery.ServeHTTP(recorder, new(http.Request), noPanicHandler)
			body := recorder.Body.String()
			Convey("Status code should be OK", func() {
				So(recorder.Code, ShouldEqual, http.StatusOK)
			})
			Convey("body should NOT contain any stacktrace", func() {
				containsStackTraceClass := strings.Contains(body, "class=\"stacktrace\"")
				So(containsStackTraceClass, ShouldBeFalse)
			})
		})
	})
}

func TestStandardRecovery(t *testing.T) {
	Convey("When using standard recovery", t, func() {
		Convey("And a panic occurs", func() {
			Convey("Recovery should", func() {
				recovery := NewRecovery()
				recorder := httptest.NewRecorder()
				recovery.ServeHTTP(recorder, new(http.Request), panicHandler)
				body := recorder.Body.String()
				Convey("Have a status code of InternalServerError", func() {
					So(recorder.Code, ShouldEqual, http.StatusInternalServerError)
				})
				Convey("body should NOT contain the stacktrace", func() {
					containsStackTraceClass := strings.Contains(body, "class=\"stacktrace\"")
					So(containsStackTraceClass, ShouldBeFalse)
				})
				Convey("body should have text Internal Server Error", func() {
					So(body, ShouldEqual, "Internal Server Error")
				})
			})
		})
		Convey("No panic occurs", func() {
			recovery := NewRecoveryWithStackTrace()
			recorder := httptest.NewRecorder()
			recovery.ServeHTTP(recorder, new(http.Request), noPanicHandler)
			body := recorder.Body.String()
			Convey("Status code should be OK", func() {
				So(recorder.Code, ShouldEqual, http.StatusOK)
			})
			Convey("body should NOT contain any stacktrace", func() {
				containsStackTraceClass := strings.Contains(body, "class=\"stacktrace\"")
				So(containsStackTraceClass, ShouldBeFalse)
			})
		})
	})
}
