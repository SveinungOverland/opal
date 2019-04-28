package router

import (
	"github.com/SveinungOverland/opal/http"
)

// HandleFunc is function that represents the handler for a HTTP-Endpoint
type HandleFunc func(req *http.Request, res *http.Response)

// Router manages routes for given http-endpoints
type Router struct {
	basePath string
	root     *Route
}

// NewRouter creates a new router to build routes with.
func NewRouter(basePath string) *Router {
	return &Router{
		basePath: basePath,
		root:     NewRoot(),
	}
}

// ------ ROUTE BUILDERS --------

// Get initializes a GET-endpoint at given path.
func (r *Router) Get(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "GET", funcs)
}

// Post initializes a POST-endpoint at given path.
func (r *Router) Post(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "POST", funcs)
}

// Put initializes a PUT-endpoint at given path.
func (r *Router) Put(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "PUT", funcs)
}

// Delete initializes a DELETE-endpoint at given path.
func (r *Router) Delete(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "DELETE", funcs)
}

// Patch initializes a PATCH-endpoint at given path.
func (r *Router) Patch(path string, funcs ...HandleFunc) {
	createFullRoute(r.root, path, "PATCH", funcs)
}

// Static initializes a static route for handling searches for static files.
func (r *Router) Static(path string, relativePath string) {
	leafRoute, _ := createOrFindRoute(r.root, path)
	leafRoute.static = true
	leafRoute.staticPath = relativePath
}

// Root returns the root of the router.
func (r *Router) Root() *Route {
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
