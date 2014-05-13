package gonion

import (
	"reflect"
	"runtime"
	"strings"
)

//Storage for route information
type RouteRegistry struct {
	Routes []*Route
}

type Handler interface{}

type Route struct {
	Method  string
	Pattern string
	Handler Handler
}

func (r *RouteRegistry) AddFunc(i interface{}) {
	route := buildRouteForFunc(i)
	r.Routes = append(r.Routes, route)
}

func buildRouteForFunc(i interface{}) *Route {
	//TODO pull out url policy
	funcType := reflect.TypeOf(i)
	if funcType.Kind() != reflect.Func {
		panic("gonion: handler for route must be a func")
	}
	funcName := runtime.FuncForPC(
		reflect.ValueOf(i).Pointer()).Name()

	funcName = strings.Split(funcName, ".")[1]
	funcName = strings.ToLower(funcName)
	parts := strings.Split(funcName, "_")
	method := strings.ToUpper(parts[0])

	var pattern string
	for _, part := range parts[1:] {
		pattern = pattern + "/"
		if part != "index" {
			pattern = pattern + part
		}
	}
	return &Route{method, pattern, i}
}

//Creates a new RouteRegistry for storing route information
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{make([]*Route, 0, 10)}
}
