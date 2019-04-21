package router

import (
	"testing"
)

// Tests if routes are initialized and all the methods are configured properly
func TestRouterMethods(t *testing.T) {
	r := NewRouter("/")

	// Initialize all routes
	var actual string

	r.Post("/", func(req *IRequest, res *IResponse) {
		actual += "POST"
	})

	r.Get("/aaaa", func(req *IRequest, res *IResponse) {
		actual += "GET"
	})

	r.Put("/bbbb", func(req *IRequest, res *IResponse) {
		actual += "PUT"
	})

	r.Delete("/bbbb/dddd/", func(req *IRequest, res *IResponse) {
		actual += "DELETE"
	})

	r.Patch("/bbbb/DDDD/", func(req *IRequest, res *IResponse) {
		actual += "PATCH"
	})

	// Get and run all the configured methods
	root := r.Root()
	route := root
	runHandlers(route.GetHandlers("POST"))
	route = getSubRoute(t, root, "aaaa")
	runHandlers(route.GetHandlers("GET"))
	route = getSubRoute(t, root, "bbbb")
	runHandlers(route.GetHandlers("PUT"))
	route = getSubRoute(t, route, "dddd")
	runHandlers(route.GetHandlers("DELETE"))
	route = getSubRoute(t, getSubRoute(t, root, "bbbb"), "dddd")
	runHandlers(route.GetHandlers("PATCH"))

	// Check if all routes were found and all methods were run
	expected := "POSTGETPUTDELETEPATCH"
	if expected != actual {
		t.Error("Route methods were not run or the routes were not found!")
	}
}

// ----- HELPERS ------

func runHandlers(hs []HandleFunc) {
	for _, h := range hs {
		h(nil, nil)
	}
}

func getSubRoute(t *testing.T, r *Route, subPath string) *Route {
	subRoute := r.subRoutes[subPath]
	if subRoute == nil {
		t.Errorf("Path '%s' does not exist, but it should (?)", subPath)
	}
	return subRoute
}
