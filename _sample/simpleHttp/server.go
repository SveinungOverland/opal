package simpleHttp

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
)

type Server struct {
	rootNode *RouteNode
	cert     tls.Certificate
	isTLS    bool
}

func NewServer() *Server {
	return &Server{
		rootNode: emptyRouteNode(),
		isTLS:    false,
	}
}

func NewTLSServer(certPath string, privateKeyPath string) *Server {
	cert, err := tls.LoadX509KeyPair(certPath, privateKeyPath)
	handleError(err)

	return &Server{
		rootNode: emptyRouteNode(),
		cert:     cert,
		isTLS:    true,
	}
}

// Listen listens
func (s *Server) Listen(port int16) {
	fmt.Printf("Starting server on port %d \n", port)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", port))
	handleError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	handleError(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		c := s.createConn(conn)
		go c.serve()
	}
}

func (s *Server) createConn(conn net.Conn) *Conn {
	c := &Conn{
		server: s,
		conn:   conn,
		isTLS:  false,
	}

	if s.isTLS {
		config := &tls.Config{Certificates: []tls.Certificate{s.cert}}
		c.tlsConn = tls.Server(conn, config)
		c.isTLS = true
	}

	return c
}

func (s *Server) Register(baseURL string, r *Router) {

	// Initialize base path
	subPaths := strings.Split(strings.TrimRight(baseURL, "/"), "/")[1:]
	routeNode := s.rootNode

	for _, path := range subPaths {
		// Maintain existing node or create new
		var nextRouteNode *RouteNode
		if val, ok := routeNode.subRoutes[path]; ok {
			nextRouteNode = val
		} else {
			nextRouteNode = emptyRouteNode()
			nextRouteNode.value = path
		}
		routeNode.subRoutes[path] = nextRouteNode
		routeNode = routeNode.subRoutes[path]
	}

	// Add routes to nodes
	for _, route := range r.routes {
		curRouteNode := routeNode

		// Route is at curret node
		if route.URL == "" {
			setMethodToRoute(route, curRouteNode)
		} else {
			// Find or create route nodes based on paths
			subPaths = strings.Split(route.URL, "/")[1:]

			for _, path := range subPaths {
				var nextRouteNode *RouteNode
				if val, ok := curRouteNode.subRoutes[path]; ok {
					nextRouteNode = val
				} else {
					nextRouteNode = emptyRouteNode()
					nextRouteNode.value = path
				}
				curRouteNode.subRoutes[path] = nextRouteNode
				curRouteNode = curRouteNode.subRoutes[path]
			}
			setMethodToRoute(route, curRouteNode)
		}
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
