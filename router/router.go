package router

import (
	"opal/http"
)

/* type Req interface {
	Param(string) string
	Header(string) string
	Body() []byte
}

type Res interface {
	SetHeader(string, string)
	SetBody([]byte)
	SetStatus(int)
} */

// HandleFunc is function that represents the handler for a HTTP-Endpoint
type HandleFunc func(req *http.Request, res *http.Response)

type router struct {
	basePath string
	root     *Route
}

// NewRouter creates a new router to build routes with.
func NewRouter(basePath string) *router {
	return &router{
		basePath: basePath,
		root:     NewRoot(),
	}
}

// ------ ROUTE BUILDERS --------

func (r *router) Get(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "GET", funcs)
}

func (r *router) Post(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "POST", funcs)
}

func (r *router) Put(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "PUT", funcs)
}

func (r *router) Delete(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "DELETE", funcs)
}

func (r *router) Patch(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "PATCH", funcs)
}

func (r *router) Static(path string, relativePath string) {
	leafRoute, _ := createOrFindRoute(r.root, path)
	leafRoute.static = true
	leafRoute.staticPath = relativePath
}

func (r *router) Root() *Route {
	return r.root
}

// ------- HELPERS ---------

func createFullRoute(root *Route, fullPath string, method string, funcs []HandleFunc) {
	route, _ := createOrFindRoute(root, fullPath)
	if route == nil {
		panic("Invalid path!")
	}

	route.addHandlers(method, funcs)
}
