![Opal_Maskot](https://user-images.githubusercontent.com/31648998/56459646-a53d1a00-6396-11e9-8b5a-7715a2796813.png)

# Opal - A HTTP2 Webframework in Go 

[![Go Report Card](https://goreportcard.com/badge/github.com/SveinungOverland/Opal)](https://goreportcard.com/report/github.com/SveinungOverland/Opal)
[![Build Status](https://travis-ci.com/SveinungOverland/opal.svg?token=qzzDg7qxp9Cyq4d1SzcF&branch=master)](https://travis-ci.com/SveinungOverland/opal)

Opal is a simple HTTP2 web-framework implemented in Go (Golang), made for fast and reliable development. Opal is a powerful package for quickly writing modular web applications/services in Go.

## Content
1. [Installation](#installation)
2. [Documentation](#documentation)
3. [Examples](#examples)
4. [Implementations](#implementations)
5. [Todo](#todo)
6. [Dependencies](#dependencies)
7. [Tests](#tests)
7. [Authors](#authors)

## Installation
To install Opal just install from the Github repo.
```
go get github.com/SveinungOverland/opal
```
## Documentation
GENERATE DOCS MAYBE?
## Examples
### Basic Usage
```go
srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
r := router.NewRouter("/")

// A simple GET-endpoint
r.Get("/", func(req *http.Request, res *http.Response) {
  res.String(200, "Hello World! :D")
})

// A simple PUT-endpoint
r.Put("/:id", func(req *http.Request, res *http.Response) {
  id := req.Param("id") // Read path parameter
  res.String(200, id)
})

srv.Register(r) // Register router
srv.Listen(443)
```

### Server Push
```go
srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
r := router.NewRouter("/")

// A simple endpoint that returns an index.html
r.Get("/", func(req *http.Request, res *http.Response) {
    res.HTML("./index.html", nil)
    res.Push("/static/app.js") // Push app.js
    res.Push("/static/index.css") // Push index.css
})

// Includes access to static files (app.js and index.css)
r.Static("/static", "./MY_STATIC_PATH")

srv.Register(r)
srv.Listen(443)
```
### Middlewares
```go
// Authorization middleware
func auth(req *http.Request, res *http.Response) {
  token := req.Query("token")
  if token != "MY_SECRET_PASSWORD" {
    req.Finish() // Stops rests of the endpoint flow
    res.Unauthorized()
  }
}

srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
r := router.NewRouter("/")

// This endpoint is protected
r.Post("/todo", auth, func(req *http.Request, res *http.Response) {
  task := string(req.Body)
  // Create new todo-item
  todo := http.JSON {
   "todo": task,
   "done": false,
  }
  res.JSON(201, todo)
}

srv.Register(r)
srv.Listen(443)
```

### Static Content
To serve static content, for example a React or Vue build, just provide the build path in router.Static().
```go
srv, err := opal.NewTLSServer("./server.crt", "./server.key", nil)
r := router.NewRouter("/")

r.Static("/", "./build") // Serves the entire build folder on root path

srv.Register(r)
srv.Listen(443)
```

## Implementations
Implemented a robust HTTP2-library managing multiple clients with REST-support, Server-Push, and support for serving static files.

#### Core of the HTTP/2 Protocol
Implemented most of the HTTP2-protocol, specified by [RFC7540](https://tools.ietf.org/html/rfc7540)
 * HTTP/2 Connection Preface, [RFC7540 Section 3.5](https://tools.ietf.org/html/rfc7540#section-3.5)
 * TLS Support, [RFC7540 Section 3.3](https://tools.ietf.org/html/rfc7540#section-3.3)
 * Stream multiplexing, [RFC7540 Section 5](https://tools.ietf.org/html/rfc7540#section-5)
    - Stream states, [RFC7540 Section 5.1](https://tools.ietf.org/html/rfc7540#section-5.1)
    - Flow control, [RFC7540 Section 5.2](https://tools.ietf.org/html/rfc7540#section-5.2)
    - Concurrency
 * Frame management, [RFC7540 Section 4](https://tools.ietf.org/html/rfc7540#section-4)
 * Server Push, [RFC7540 Section 8.2](https://tools.ietf.org/html/rfc7540#section-8.2)
 
#### HPACK - Header compression
Created a robust and solid HPACK library, [RFC7541](https://tools.ietf.org/html/rfc7541)

#### HTTP Router library
A high preformance HTTP-Router with parameter- and filehandling-functionality.

## Todo
* Add support for HTTP/1.1 ( Had no time :( )
* Implement Stream Priority, [RFC7540 Section 5.3](https://tools.ietf.org/html/rfc7540#section-5.3)


## Dependencies

* __crypto/tls__ - A TLS-library from the standard-library. [Docs](https://golang.org/pkg/crypto/tls/)
* __github.com/go-test/deep__ - A library for test-support. [Github Repo](github.com/go-test/deep)
* __github.com/fatih/color__ - A color-library for changing colors in the console. [Github Repo](github.com/fatih/color)


## Tests

## Authors
<a href="https://github.com/Andorr" target="_blank"><img src="https://avatars2.githubusercontent.com/u/31648998?s=400&v=4" width=40 title="Andorr"/></a>
<a href="https://github.com/SveinungOverland" target="_blank"><img src="https://avatars0.githubusercontent.com/u/39273837?s=460&v=4" width=40 title="SveinungOverland"/></a>
