package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Micah-Shallom/a2amux"
	"github.com/gin-gonic/gin"
	"trpc.group/trpc-go/trpc-a2a-go/server"
)

// Integration with your company's main.go pattern
func integrationExample() {
	// This would go in your router.Setup() function

	// Create A2A multiplexer
	a2aMux := a2amux.NewMultiplexer("/api/agents")

	// Setup your A2A agents
	setupA2AAgents(a2aMux)

	// In your router setup, add the middleware
	// r.Use(a2aMux.GinMiddleware())
}

// Example of how to modify your existing router.Setup() function
func routerSetupExample() *gin.Engine {
	r := gin.Default()

	// Your existing middleware
	// r.Use(cors.Default())
	// r.Use(yourAuthMiddleware())

	// Add A2A multiplexer
	a2aMux := a2amux.NewMultiplexer("/api/agents")
	setupA2AAgents(a2aMux)
	r.Use(a2aMux.GinMiddleware())

	return r
}

// Alternative: Mount on specific path group
func alternativeGroupMounting() {
	r := gin.Default()

	// Create multiplexer without base path
	a2aMux := a2amux.NewMultiplexer("")
	setupA2AAgents(a2aMux)

	// Mount only on /agents/* routes
	agentsGroup := r.Group("/agents")
	agentsGroup.Use(a2aMux.GinMiddleware())

	// Now accessible at:
	// GET /agents/reverser/.well-known/agent.json
	// POST /agents/reverser/
}

// Example using with standard HTTP (if you ever need it)
func standardHTTPExample() {
	mux := a2amux.NewMultiplexer("/api/agents")
	setupA2AAgents(mux)

	http.Handle("/", mux.HTTPHandler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func main() {
	integrateWithExistingGinServer()
}

func integrateWithExistingGinServer() {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	a2aMux := a2amux.NewMultiplexer("/api/agents")
	setupA2AAgents(a2aMux)
	r.Use(a2aMux.GinMiddleware())
	r.Run(":8080")
}

func setupA2AAgents(mux *a2amux.Multiplexer) {
	flag.Parse()

	// Configure reverser agent
	reverserConfig := a2amux.AgentConfig{
		Path:      "reverser",
		AgentCard: createReverserAgentCard(),
		Processor: &ReverserProcessor{
			texts: make([]string, 0),
		},
	}
	if err := mux.AddAgent(reverserConfig); err != nil {
		log.Fatalf("Failed to add reverser agent: %v", err)
	}

	// Configure uppercaser agent
	uppercaserConfig := a2amux.AgentConfig{
		Path:      "uppercaser",
		AgentCard: createUppercaserAgentCard(),
		Processor: &UppercaserProcessor{
			texts: make([]string, 0),
		},
		MethodMap: map[string]string{
			"message/send": "tasks/send",
		},
	}
	if err := mux.AddAgent(uppercaserConfig); err != nil {
		log.Fatalf("Failed to add uppercaser agent: %v", err)
	}

	routes := mux.ListRoutes()
	log.Printf("A2A agents configured on routes: %v", routes)
}

func createReverserAgentCard() server.AgentCard {
	host := flag.String("reverser-host", "localhost", "Host to listen on")
	port := flag.Int("reverser-port", 9000, "Port to listen on")
	return reverserAgentCard(*host, *port)
}

func createUppercaserAgentCard() server.AgentCard {
	host := flag.String("uppercaser-host", "localhost", "Host to listen on")
	port := flag.Int("uppercaser-port", 9001, "Port to listen on")
	return uppercaserAgentCard(*host, *port)
}
