package recovery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var panicHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	panic("oh no!")
})

var noPanicHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
})

func TestRecoveryWithStackTrace_WithPanic(t *testing.T) {
	recovery := WithStackTrace(panicHandler)
	recorder := httptest.NewRecorder()
	recovery.ServeHTTP(recorder, new(http.Request))
	body := recorder.Body.String()
	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
	assert.True(t, strings.Contains(body, "class=\"stacktrace\""))
	assert.True(t, strings.Contains(body, "oh no!"))
}

func TestRecoveryWithStackTrace_WithoutPanic(t *testing.T) {
	recovery := WithStackTrace(noPanicHandler)
	recorder := httptest.NewRecorder()
	recovery.ServeHTTP(recorder, new(http.Request))
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.False(t, strings.Contains(recorder.Body.String(), "class=\"stacktrace\""))
}

func TestStandardRecovery_WithPanic(t *testing.T) {
	recovery := Recovery(panicHandler)
	recorder := httptest.NewRecorder()
	recovery.ServeHTTP(recorder, new(http.Request))
	body := recorder.Body.String()
	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
	assert.False(t, strings.Contains(body, "class=\"stacktrace\""))
	assert.Equal(t, body, "Internal Server Error")
}

func TestStandardRecovery_WithoutPanic(t *testing.T) {
	recovery := Recovery(noPanicHandler)
	recorder := httptest.NewRecorder()
	recovery.ServeHTTP(recorder, new(http.Request))
	assert.Equal(t, recorder.Code, http.StatusOK)
	assert.False(t, strings.Contains(recorder.Body.String(), "class=\"stacktrace\""))
}
