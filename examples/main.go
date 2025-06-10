package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"trpc.group/trpc-go/trpc-a2a-go/server"
	"trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

// Example: Integrating A2A multiplexer with your existing Gin server
func integrateWithExistingGinServer() {
	// Your existing Gin setup (similar to your company's setup)
	r := gin.Default()

	// Your existing routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API group for your existing APIs
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/users", getUsersHandler)
		apiV1.POST("/users", createUserHandler)
		// ... your existing routes
	}

	// Create A2A multiplexer for agents
	a2aMux := a2amux.NewMultiplexer("/api/agents") // Base path for all A2A agents

	// Optional: Use your company's logger
	// a2aMux.SetLogger(yourCompanyLogger)

	// Add your A2A agents
	setupA2AAgents(a2aMux)

	// Mount the A2A multiplexer on your Gin router
	r.Use(a2aMux.GinMiddleware())

	// Your existing server startup logic would go here
	// r.Run(":8080")
}

// Example: Setting up A2A agents
func setupA2AAgents(mux *a2amux.Multiplexer) {
	// Agent 1: Text Reverser
	reverserProcessor := &ReverserProcessor{
		texts: make([]string, 0),
	}
	reverserTM, err := taskmanager.NewMemoryTaskManager(reverserProcessor)
	if err != nil {
		log.Fatalf("Failed to create reverser task manager: %v", err)
	}
	reverserServer, err := server.NewA2AServer(createReverserAgentCard(), reverserTM)
	if err != nil {
		log.Fatalf("Failed to create reverser A2A server: %v", err)
	}

	err = mux.AddRoute(a2amux.AgentRoute{
		Path:   "reverser", // Will be accessible at /api/agents/reverser
		Server: reverserServer,
	})
	if err != nil {
		log.Fatalf("Failed to add reverser route: %v", err)
	}

	// Agent 2: Text Uppercaser
	uppercaserProcessor := &UppercaserProcessor{
		texts: make([]string, 0),
	}
	uppercaserTM, err := taskmanager.NewMemoryTaskManager(uppercaserProcessor)
	if err != nil {
		log.Fatalf("Failed to create uppercaser task manager: %v", err)
	}
	uppercaserServer, err := server.NewA2AServer(createUppercaserAgentCard(), uppercaserTM)
	if err != nil {
		log.Fatalf("Failed to create uppercaser A2A server: %v", err)
	}

	err = mux.AddRoute(a2amux.AgentRoute{
		Path:   "uppercaser", // Will be accessible at /api/agents/uppercaser
		Server: uppercaserServer,
		MethodMap: map[string]string{
			"message/send":     "tasks/send",
			"message/get":      "tasks/get",
			"custom/transform": "tasks/transform",
		},
	})
	if err != nil {
		log.Fatalf("Failed to add uppercaser route: %v", err)
	}

	// Agent 3: Translation Service (example)
	// translatorServer := setupTranslatorAgent()
	// mux.AddRoute(a2amux.AgentRoute{
	//     Path:   "translator",
	//     Server: translatorServer,
	// })

	// List all configured routes
	routes := mux.ListRoutes()
	log.Printf("A2A agents configured on routes: %v", routes)
}

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
func routerSetupExample() gin.Engine {
	r := gin.Default()

	// Your existing middleware
	// r.Use(cors.Default())
	// r.Use(yourAuthMiddleware())

	// Add A2A multiplexer
	a2aMux := a2amux.NewMultiplexer("/api/agents")
	setupA2AAgents(a2aMux)
	r.Use(a2aMux.GinMiddleware())

	// Your existing route groups
	apiV1 := r.Group("/api/v1")
	{
		// Your existing routes...
	}

	return *r
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

// Dummy handlers for example
func getUsersHandler(c *gin.Context) {
	c.JSON(200, gin.H{"users": []string{}})
}

func createUserHandler(c *gin.Context) {
	c.JSON(201, gin.H{"message": "user created"})
}

// Dummy processor implementations (you'd replace these with your actual ones)
type ReverserProcessor struct {
	texts []string
}

type UppercaserProcessor struct {
	texts []string
}

func createReverserAgentCard() interface{} {
	// Your agent card implementation
	return nil
}

func createUppercaserAgentCard() interface{} {
	// Your agent card implementation
	return nil
}
