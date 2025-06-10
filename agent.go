package a2amux

import "trpc.group/trpc-go/trpc-a2a-go/server"

type AgentRoute struct {
	Path      string
	Server    *server.A2AServer
	MethodMap map[string]string
}
