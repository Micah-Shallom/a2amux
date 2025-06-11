package a2amux

import (
	"fmt"

	"trpc.group/trpc-go/trpc-a2a-go/server"
	"trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

type AgentConfig struct {
	Path      string
	AgentCard server.AgentCard
	Server    *server.A2AServer
	Processor taskmanager.TaskProcessor
	MethodMap map[string]string
}

func (m *Multiplexer) AddAgent(config AgentConfig) error {
	if config.Path == "" {
		return fmt.Errorf("agent config path cannot be empty")
	}
	if config.Processor == nil {
		return fmt.Errorf("agent config processor cannot be nil")
	}

	tm, err := taskmanager.NewMemoryTaskManager(config.Processor)
	if err != nil {
		return fmt.Errorf("failed to create task manager for agent '%s': %w", config.Path, err)
	}

	a2aServer, err := server.NewA2AServer(config.AgentCard, tm)
	if err != nil {
		return fmt.Errorf("failed to create A2A server for agent '%s': %w", config.Path, err)
	}

	route := AgentConfig{
		Path:      config.Path,
		Server:    a2aServer,
		MethodMap: config.MethodMap,
	}

	return m.AddRoute(route)
}
