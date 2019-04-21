package router

import (
	"fmt"
)

// Route represents a node in a http-router, and contains http-method-implementation
type Route struct {
	value      string
	subRoutes  map[string]*Route
	paramRoute *Route

	static     bool
	staticPath string

	Get    []HandleFunc
	Post   []HandleFunc
	Put    []HandleFunc
	Delete []HandleFunc
	Patch  []HandleFunc
}

func newRoute(pathValue string) *Route {
	return &Route{
		value:     pathValue,
		subRoutes: make(map[string]*Route),
	}
}

func NewRoot() *Route {
	return newRoute("/")
}

func (r *Route) AppendRouter(router *Router) {
	leafRoute, _ := createOrFindRoute(r, router.basePath)
	leafRoute.merge(router.root)
}

func (r *Route) Search(path string) (match bool, route *Route, params map[string]string, fh *FileHandler) {
	return search(r, path)
}

func (r *Route) addHandlers(method string, funcs []HandleFunc) {
	switch method {
	case "GET":
		r.Get = funcs
	case "POST":
		r.Post = funcs
	case "PUT":
		r.Put = funcs
	case "DELETE":
		r.Delete = funcs
	case "PATCH":
		r.Patch = funcs
	}
}

func (r *Route) GetHandlers(method string) []HandleFunc {
	switch method {
	case "GET":
		return r.Get
	case "POST":
		return r.Post
	case "PUT":
		return r.Put
	case "DELETE":
		return r.Delete
	case "PATCH":
		return r.Patch
	}
	return nil
}

func (r *Route) merge(route *Route) {
	// Overwrite config
	r.static = route.static
	r.Get = route.Get
	r.Post = route.Post
	r.Put = route.Put
	r.Delete = route.Delete
	r.Patch = route.Patch
	r.subRoutes = route.subRoutes
}

// ---- HELPERS ------

func (r *Route) String() string {
	return r.string(0)
}

func (r *Route) string(depth int) string {
	var s string
	for d := 0; d < depth; d++ {
		s += " "
	}
	if depth != 0 {
		s += "/"
	}

	var methods []string
	if len(r.Get) > 0 {
		methods = append(methods, "GET")
	}
	if len(r.Post) > 0 {
		methods = append(methods, "POST")
	}
	if len(r.Put) > 0 {
		methods = append(methods, "PUT")
	}
	if len(r.Delete) > 0 {
		methods = append(methods, "DELETE")
	}
	if len(r.Patch) > 0 {
		methods = append(methods, "PATCH")
	}
	s += fmt.Sprintf("%s  %v", r.value, methods)
	if r.static {
		s += "    STATIC"
	}
	s += "\n"

	for _, v := range r.subRoutes {
		s += v.string(depth + 3)
	}
	if r.paramRoute != nil {
		s += r.paramRoute.string(depth + 3)
	}
	return s
}
