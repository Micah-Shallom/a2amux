package a2amux

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Multiplexer struct {
	routes   map[string]*wrappedHandler
	logger   Logger
	basePath string
}

// NewMultiplexer creates a new A2A multiplexer
func NewMultiplexer(basePath string) *Multiplexer {
	// Ensure basePath starts with / and doesn't end with /
	if basePath == "" {
		basePath = ""
	} else {
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
	}

	return &Multiplexer{
		routes:   make(map[string]*wrappedHandler),
		logger:   DefaultLogger{},
		basePath: basePath,
	}
}

// set up custom logger here
func (m *Multiplexer) SetLogger(logger Logger) {
	m.logger = logger
}

// AddRoute adds an A2A server to a specific route
func (m *Multiplexer) AddRoute(route AgentRoute) error {
	// Clean the path - remove leading/trailing slashes
	routePath := strings.Trim(route.Path, "/")

	// Default method map if not provided
	if route.MethodMap == nil {
		route.MethodMap = map[string]string{
			"message/send": "tasks/send", //based on new A2A specs that hasnt been updated
			// "message/get":  "tasks/get",
		}
	}

	wrapper := &wrappedHandler{
		server:    route.Server,
		methodMap: route.MethodMap,
		logger:    m.logger,
	}

	m.routes[routePath] = wrapper
	m.logger.Printf("Added A2A route: %s%s", m.basePath, "/"+routePath)
	return nil
}

// RemoveRoute removes a route from the multiplexer
func (m *Multiplexer) RemoveRoute(path string) {
	routePath := strings.Trim(path, "/")
	delete(m.routes, routePath)
	m.logger.Printf("Removed A2A route: %s%s", m.basePath, "/"+routePath)
}

func (m *Multiplexer) ListRoutes() []string {
	routes := make([]string, 0, len(m.routes))
	for route := range m.routes {
		routes = append(routes, m.basePath+"/"+route)
	}
	return routes
}

// GinMiddleware returns a Gin middleware that handles A2A requests
func (m *Multiplexer) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Remove base path from the request path
		requestPath := c.Request.URL.Path
		if m.basePath != "" && strings.HasPrefix(requestPath, m.basePath) {
			requestPath = strings.TrimPrefix(requestPath, m.basePath)
		}

		// Remove leading slash
		requestPath = strings.TrimPrefix(requestPath, "/")

		m.logger.Printf("A2A Multiplexer received request: %s %s (mapped to: %s)",
			c.Request.Method, c.Request.URL.Path, requestPath)

		// Find the agent route
		parts := strings.Split(requestPath, "/")
		if len(parts) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"jsonrpc": "2.0",
				"error": gin.H{
					"code":    -32601,
					"message": "Agent not found",
				},
				"id": nil,
			})
			c.Abort()
			return
		}

		agentName := parts[0]
		handler, exists := m.routes[agentName]
		if !exists {
			m.logger.Printf("No route found for agent: %s (available: %v)", agentName, m.getRouteNames())
			c.JSON(http.StatusNotFound, gin.H{
				"jsonrpc": "2.0",
				"error": gin.H{
					"code":    -32601,
					"message": "Agent not found",
				},
				"id": nil,
			})
			c.Abort()
			return
		}

		// Determine the new path for the A2A server
		var newPath string
		if len(parts) == 1 {
			// /agent -> /
			newPath = "/"
		} else {
			// /agent/something -> /something
			newPath = "/" + strings.Join(parts[1:], "/")
		}

		// Handle agent card requests
		if newPath == "/.well-known/agent.json" && c.Request.Method == http.MethodGet {
			m.logger.Printf("Serving agent card for %s: %s -> %s", agentName, c.Request.URL.Path, newPath)
			c.Request.URL.Path = newPath
			handler.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// Handle JSON-RPC requests
		m.logger.Printf("Routing request for %s: %s -> %s", agentName, c.Request.URL.Path, newPath)
		c.Request.URL.Path = newPath
		handler.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func (m *Multiplexer) getRouteNames() []string {
	names := make([]string, 0, len(m.routes))
	for name := range m.routes {
		names = append(names, name)
	}
	return names
}
