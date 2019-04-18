package router

// Change back to Request and Response when ready ^^
type HandleFunc func(req interface{} /* *Request */, res interface{} /* *Response */)

type router struct {
	basePath string
	root     *route
}

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

func (r *router) Root() *route {
	return r.root
}

// ------- HELPERS ---------

func createFullRoute(root *route, fullPath string, method string, funcs []HandleFunc) {
	route, _ := createOrFindRoute(root, fullPath)
	if route == nil {
		panic("Invalid path!")
	}

	route.addHandlers(method, funcs)
}
