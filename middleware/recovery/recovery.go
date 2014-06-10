package recovery

import (
	"html/template"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
)

type Recovery struct {
}

type RecoveryWithStackTrace struct {
	template *template.Template
}

func NewRecovery() *Recovery {
	return &Recovery{}
}

func NewRecoveryWithStackTrace() *RecoveryWithStackTrace {
	return &RecoveryWithStackTrace{
		template: createErrorTemplate(),
	}
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

func (recovery *RecoveryWithStackTrace) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
	handlePanic(func(err interface{}) {
		stack := debug.Stack()
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Header().Set("Content-Type", "text/html")
		pe := &panicError{err, string(stack)}
		recovery.template.Execute(rw, pe)
	}, next, rw, r)
}

func (recovery *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.Handler) {
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
