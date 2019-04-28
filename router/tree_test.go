package router

import (
	"testing"
)

func TestSearchForRoot(t *testing.T) {
	r := getSearchTestRouter("/")
	root := r.Root()
	testSearch(t, root, "/", true, "/")
}

func TestSearchDeep(t *testing.T) {
	r := getSearchTestRouter("/")
	root := r.Root()
	testSearch(t, root, "/asdf/1234", true, "1234")
	testSearch(t, root, "/bbbb/aaaa/", true, "aaaa")
	testSearch(t, root, "/bbb", false, "/")
}

func TestSearchParams(t *testing.T) {
	r := getSearchTestRouter("/")
	root := r.Root()
	testSearchParams(t, root, "/asdf/12345", true, testParam{"id", "12345"})
	testSearchParams(t, root, "/asdf/1234-12344-4321/asdf", false)
	testSearchParams(t, root, "/bbbb/123.123/asdf.asdf", true, testParam{"lat", "123.123"}, testParam{"lng", "asdf.asdf"})
	testSearchParams(t, root, "/bbbb/123.123/asdf.asdf/asdf", false, testParam{"lat", "123.123"}, testParam{"lng", "asdf.asdf"})
}

func TestStaticSearch(t *testing.T) {
	r := getSearchTestRouter("/")
	r.Static("/", "./testPath")

	// Test static route from root
	testStaticPath(t, r.Root(), "/myTestFile.txt", "./testPath/myTestFile.txt", "text/plain")

	// Test static route from path not equal to root
	r = NewRouter("/")
	r.Static("/staticPath", "./")
	testStaticPath(t, r.Root(), "/staticPath/styles.css", "./styles.css", "text/css")

}

func TestCreateOrFindRoute(t *testing.T) {
	route := newRoute("/")

	// Create new route
	newRoute, created := createOrFindRoute(route, "testpath")
	if !created {
		t.Error("Returned bool value 'created' should not be false!")
	}
	if newRoute.value != "testpath" {
		t.Errorf("New route has incorrect value. Expected %s, got %s", "testpath", newRoute.value)
	}

	// Create new nested route
	newRoute, created = createOrFindRoute(route, "testpath/test/asdf")
	if !created {
		t.Error("Returned bool value 'created' should not be false!")
	}
	if newRoute.value != "asdf" {
		t.Errorf("New route has incorrect value. Expected %s, got %s", "asdf", newRoute.value)
	}

	// Find nested route
	foundRoute, created := createOrFindRoute(route, "testpath/test/")
	if created {
		t.Error("Returned bool value 'created' should not be 'true'!")
	}
	if foundRoute.value != "test" {
		t.Errorf("Found route has incorrect value. Expected %s, got %s", "test", newRoute.value)
	}
}

// ---- HELPERS ----
func getSearchTestRouter(basePath string) *Router {
	r := NewRouter(basePath)

	r.Post("/")
	r.Get("/asdf/1234")
	r.Put("/asdf/:id")
	r.Delete("BBBB/aAaA/")
	r.Patch("/bbbb/:lat/:lng")
	return r
}

func testSearch(t *testing.T, root *Route, path string, shouldMatch bool, routeValue string) {
	match, route, _, _ := root.Search(path)
	if match != shouldMatch {
		t.Errorf("The search did not find a matching route for %s! Stopped at %s", path, routeValue)
		return
	}

	if route.value != routeValue {
		t.Errorf("Matching route's value is incorrect: Expected %s, got %s", routeValue, route.value)
	}
}

type testParam struct {
	id    string
	value string
}

func testSearchParams(t *testing.T, root *Route, path string, shouldMatch bool, expectedParams ...testParam) {
	match, _, params, _ := root.Search(path)
	if match != shouldMatch {
		t.Errorf("The search did not find a matching route for %s!", path)
		return
	}

	for _, p := range expectedParams {
		val, ok := params[p.id]
		if !ok {
			t.Errorf("Could not find param %s in path %s", p.id, path)
		}
		if val != p.value {
			t.Errorf("Incorrect param value in search. Expected %s, got %s", p.value, val)
		}
	}

}

func testStaticPath(t *testing.T, root *Route, path, fullFilePath string, mimeType string) {
	match, _, _, fh := search(root, path)
	if match {
		t.Error("Found matching route on static path!")
	}
	if fh == nil {
		t.Error("Got nil as fileHandler on static path!")
		return
	}

	if fh.FullPath() != fullFilePath {
		t.Errorf("Incorrect file path in filehandler. Expected %s, got %s", fullFilePath, fh.FullPath())
	}
	if fh.MimeType != mimeType {
		t.Errorf("Incorrect mimetype in filehandler. Expected %s, got %s", mimeType, fh.MimeType)
	}
}
