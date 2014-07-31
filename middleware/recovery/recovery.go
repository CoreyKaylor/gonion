package recovery

import (
	"html/template"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
)

//SimpleRecovery is a an http.Handler that only reports a 500 status code
//without rendering the stacktrace.
type SimpleRecovery struct {
}

//StackTraceRecovery is intended for development and renders
//a stacktrace of where the panic occurred.
type StackTraceRecovery struct {
	template *template.Template
}

//Recovery is a factory method for SimpleRecovery
func Recovery(inner http.Handler) http.Handler {
	recovery := &SimpleRecovery{}
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		recovery.ServeHTTP(rw, r, inner)
	})
}

//WithStackTrace is a factory method for RecoveryWithStackTrace
func WithStackTrace(inner http.Handler) http.Handler {
	recovery := &StackTraceRecovery{
		template: createErrorTemplate(),
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		recovery.ServeHTTP(rw, r, inner)
	})
}

func createErrorTemplate() *template.Template {
	filename := getCurrentFile()
	dir := path.Dir(filename)
	recoveryFile := path.Join(dir, "recovery.html")
	return template.Must(template.ParseFiles(recoveryFile))
}

func getCurrentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	//hack: not sure why Caller is inconsistent between test and run
	if strings.Contains(filename, "recovery/_test") {
		_, filename, _, _ = runtime.Caller(3)
	}
	return filename
}

type panicError struct {
	ErrorMessage interface{}
	StackTrace   string
}

//ServeHTTP is the implementation of the standard http.Handler interface
//that will render a stacktrace.
func (recovery *StackTraceRecovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	handlePanic(func(err interface{}) {
		stack := debug.Stack()
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "text/html")
		pe := &panicError{err, string(stack)}
		recovery.template.Execute(rw, pe)
	}, next, rw, r)
}

//ServeHTTP is the implementation of the standard http.Handler interface
//that will only report a Internal Server Error
func (recovery *SimpleRecovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	handlePanic(func(err interface{}) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Internal Server Error"))
	}, next, rw, r)
}

func handlePanic(afterPanicFunc func(interface{}), next http.Handler, rw http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			afterPanicFunc(err)
		}
	}()
	next.ServeHTTP(rw, r)
}
