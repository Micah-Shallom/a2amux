package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"trpc.group/trpc-go/trpc-a2a-go/protocol"
	"trpc.group/trpc-go/trpc-a2a-go/server"
	"trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

// UppercaserProcessor handles text uppercasing tasks
type UppercaserProcessor struct {
	texts []string
}

func (p *UppercaserProcessor) Process(
	ctx context.Context,
	taskID string,
	message protocol.Message,
	handle taskmanager.TaskHandle,
) error {
	// Log the incoming message for debugging
	log.Printf("Uppercaser processing task %s with message: %+v", taskID, message)

	// Extract text from the message
	text := extractText(message)
	if text == "" {
		errMsg := protocol.NewMessage(
			protocol.MessageRoleAgent,
			[]protocol.Part{protocol.NewTextPart("input message must contain text")},
		)
		_ = handle.UpdateStatus(protocol.TaskStateFailed, &errMsg)
		return fmt.Errorf("input message must contain text")
	}

	// Store the text
	p.texts = append(p.texts, text)
	log.Printf("Uppercaser task %s: Stored text: %s", taskID, text)

	// Convert text to uppercase
	uppercased := strings.ToUpper(text)

	// Create response message
	responseMessage := protocol.NewMessage(
		protocol.MessageRoleAgent,
		[]protocol.Part{protocol.NewTextPart("Text uppercased successfully")},
	)

	// Create artifact
	artifact := protocol.Artifact{
		Name:        stringPtr("Uppercased Text"),
		Description: stringPtr("The input text converted to uppercase"),
		Index:       0,
		Parts:       []protocol.Part{protocol.NewTextPart(uppercased)},
		LastChunk:   boolPtr(true),
	}

	// Update task status to completed
	if err := handle.UpdateStatus(protocol.TaskStateCompleted, &responseMessage); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Add the artifact
	if err := handle.AddArtifact(artifact); err != nil {
		log.Printf("Error adding artifact for uppercaser task %s: %v", taskID, err)
	}

	return nil
}

// uppercaserAgentCard creates the agent card for the uppercaser agent
func uppercaserAgentCard(host string, port int) server.AgentCard {
	return server.AgentCard{
		Name:        "Text Uppercaser Agent",
		Description: stringPtr("An agent that converts input text to uppercase"),
		URL:         fmt.Sprintf("http://%s:%d/uppercaser/", host, port),
		Version:     "1.0.0",
		Provider: &server.AgentProvider{
			Organization: "tRPC-A2A-Go Examples",
		},
		Capabilities: server.AgentCapabilities{
			Streaming:              false,
			StateTransitionHistory: true,
		},
		DefaultInputModes:  []string{string(protocol.PartTypeText)},
		DefaultOutputModes: []string{string(protocol.PartTypeText)},
		Skills: []server.AgentSkill{
			{
				ID:          "text_uppercasing",
				Name:        "Text Uppercasing",
				Description: stringPtr("Converts the input text to uppercase"),
				Tags:        []string{"text", "processing"},
				Examples:    []string{"Hello, world!"},
				InputModes:  []string{string(protocol.PartTypeText)},
				OutputModes: []string{string(protocol.PartTypeText)},
			},
		},
	}
}
