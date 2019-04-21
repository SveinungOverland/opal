package router

import (
	"fmt"
	"strings"
)

// creates or finds a route based a path relative to a provided root-route.
// New routes will be created and appended to the tree if a wanted path does not exist.
// fullPath is the full path (for example "/aaaaa/bbbbb/")
// It returns the leaf route of provided path, and a bool indicating if the route was created
func createOrFindRoute(root *Route, fullPath string) (*Route, bool) {
	// Create sub paths
	subPaths := pathToSubPaths(fullPath)
	if len(subPaths) == 0 {
		return nil, false
	}

	// Check if appended route has a root-value "/"
	if subPaths[0] == "/" {
		return root, false
	}

	// Build route
	curRoute := root
	created := false
	for _, path := range subPaths {

		// Check if path is a parameter
		if strings.HasPrefix(path, ":") {
			// If the route already has a paramter subscribed, panic
			if curRoute.paramRoute != nil {
				panic(fmt.Sprintf("Path '%s' conflicts with existing route due to '%s'", fullPath, path))
			}
			paramRoute := newRoute(path)
			curRoute.paramRoute = paramRoute
			curRoute = paramRoute
			continue
		}

		// Get subroute with matching path
		subRoute, has := curRoute.subRoutes[path]

		// If the subroute, create new route
		if !has {
			subRoute = newRoute(path)
			curRoute.subRoutes[path] = subRoute
			created = true
		}

		// Check if the subroute is static
		if subRoute.static {
			panic("Can not add new route on static route!")
		}

		curRoute = subRoute
	}

	return curRoute, created
}

// Search searches after a route relative to a provided root route with a given path.
// It returns "match", which indicates if a route was found.
// "r" is the route found. If match is true, then r is the matching route. Otherwise it is farthest route found based on the path.
func search(root *Route, path string) (match bool, r *Route, params map[string]string, fh *FileHandler) {
	subPaths := pathToSubPaths(path)
	params = make(map[string]string)

	if len(subPaths) == 0 {
		return false, nil, nil, nil
	}

	// Check if provided path is referencing the root
	if subPaths[0] == "/" {
		return true, root, nil, nil
	}

	// Find matching route
	curNode := root
	for i, path := range subPaths {
		// Check if any subroutes matches the path
		if route, ok := curNode.subRoutes[path]; ok {
			curNode = route

			// Check if route is static
			if curNode.static {
				return false, curNode, params, newFileHandler(curNode.staticPath, strings.Join(subPaths[i+1:], "/"))
			}

			continue
		}

		// Otherwise check if a param-route exist
		if curNode.paramRoute != nil {
			curNode = curNode.paramRoute
			params[strings.TrimLeft(curNode.value, ":")] = path // Add param value
			continue
		}

		// No match found!
		return false, curNode, params, nil
	}

	return true, curNode, params, nil
}

// ------ HELPERS ---------

func pathToSubPaths(path string) []string {
	path = strings.TrimSpace(path)

	// Check if path is root path
	if path == "/" || path == "" {
		return []string{"/"}
	}

	// Trim slashes on the edges
	path = strings.Trim(path, "/")
	subPaths := strings.Split(path, "/")
	for i, path := range subPaths {
		subPaths[i] = strings.ToLower(path)
	}

	return subPaths
}
