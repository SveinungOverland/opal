package simpleHttp

import (
	"strings"
)

type HandleFunc func(req *Request, res *Response)

type Router struct {
	baseURL string
	routes  []*Route
}

type RouteNode struct {
	value     string
	subRoutes map[string]*RouteNode
	Get       *Route
	Post      *Route
	Put       *Route
	Delete    *Route
}

type Route struct {
	URL       string
	method    string
	functions []HandleFunc
}

// --- INITIALIZERS ----
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

func emptyRouteNode() *RouteNode {
	return &RouteNode{
		subRoutes: make(map[string]*RouteNode, 0),
	}
}

// ---- METHODS ----
func (r *Router) Get(path string, funcs ...HandleFunc) {
	appendRoute(r, "GET", path, funcs)
}

func (r *Router) Post(path string, funcs ...HandleFunc) {
	appendRoute(r, "POST", path, funcs)
}

func appendRoute(r *Router, method string, path string, funcs []HandleFunc) {
	route := &Route{method: method, URL: strings.TrimRight(path, "/"), functions: funcs}
	r.routes = append(r.routes, route)
}

func (r *Route) Exec(req *Request, res *Response) *Response {
	for _, routeMethod := range r.functions {
		routeMethod(req, res)
	}
	return res
}

// ---- HELPERS ----
func baseURLMatch(r *Router, baseURL string) bool {
	if r.baseURL == "/" && baseURL == "/" {
		return true
	}

	baseURL = strings.TrimRight(baseURL, "/")
	return strings.HasPrefix(baseURL, r.baseURL)
}

func setMethodToRoute(route *Route, node *RouteNode) {
	method := route.method
	switch method {
	case "GET":
		node.Get = route
	case "POST":
		node.Post = route
	case "DELETE":
		node.Delete = route
	case "PUT":
		node.Put = route
	}
}

func findRouteByRequest(root *RouteNode, req *Request) *Route {
	subPaths := strings.Split(strings.TrimRight(req.RequestURI, "/"), "/")[1:]

	if len(subPaths) == 0 {
		return findRouteByMethod(root, req.Method)
	}

	// Iterate through paths and find the route-node
	curRouteNode := root
	for _, path := range subPaths {

		if _, ok := curRouteNode.subRoutes[path]; !ok {
			return nil
		}
		curRouteNode = curRouteNode.subRoutes[path]
	}

	// Return route
	if curRouteNode != nil {
		return findRouteByMethod(curRouteNode, req.Method)
	} else {
		return nil
	}
}

func findRouteByMethod(node *RouteNode, method string) *Route {
	switch method {
	case "GET":
		return node.Get
	case "POST":
		return node.Post
	case "DELETE":
		return node.Delete
	case "PUT":
		return node.Put
	}
	return nil
}
