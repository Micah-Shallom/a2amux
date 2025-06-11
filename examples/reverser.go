package main

import (
	"context"
	"fmt"
	"log"

	"trpc.group/trpc-go/trpc-a2a-go/protocol"
	"trpc.group/trpc-go/trpc-a2a-go/server"
	"trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

// ReverserProcessor handles text reversal tasks
type ReverserProcessor struct {
	texts []string
}

func (p *ReverserProcessor) Process(
	ctx context.Context,
	taskID string,
	message protocol.Message,
	handle taskmanager.TaskHandle,
) error {
	// Log the incoming message for debugging
	log.Printf("Reverser processing task %s with message: %+v", taskID, message)

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
	log.Printf("Reverser task %s: Stored text: %s", taskID, text)

	// Reverse the text for the artifact
	reversed := reverseString(text)

	// Create response message
	responseMessage := protocol.NewMessage(
		protocol.MessageRoleAgent,
		[]protocol.Part{protocol.NewTextPart("Text reversed successfully")},
	)

	// Create artifact
	artifact := protocol.Artifact{
		Name:        stringPtr("Reversed Text"),
		Description: stringPtr("The input text reversed"),
		Index:       0,
		Parts:       []protocol.Part{protocol.NewTextPart(reversed)},
		LastChunk:   boolPtr(true),
	}

	// Update task status to completed
	if err := handle.UpdateStatus(protocol.TaskStateCompleted, &responseMessage); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Add the artifact
	if err := handle.AddArtifact(artifact); err != nil {
		log.Printf("Error adding artifact for reverser task %s: %v", taskID, err)
	}

	return nil
}

// reverserAgentCard creates the agent card for the reverser agent
func reverserAgentCard(host string, port int) server.AgentCard {
	return server.AgentCard{
		Name:        "Text Reverser Agent",
		Description: stringPtr("An agent that reverses input text"),
		URL:         fmt.Sprintf("http://%s:%d/reverser/", host, port),
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
				ID:          "text_reversal",
				Name:        "Text Reversal",
				Description: stringPtr("Reverses the input text"),
				Tags:        []string{"text", "processing"},
				Examples:    []string{"Hello, world!"},
				InputModes:  []string{string(protocol.PartTypeText)},
				OutputModes: []string{string(protocol.PartTypeText)},
			},
		},
	}
}


func extractText(message protocol.Message) string {
	for _, part := range message.Parts {
		if textPart, ok := part.(protocol.TextPart); ok {
			return textPart.Text
		}
	}
	return ""
}

// extractMethod extracts the method from message metadata
func extractMethod(message protocol.Message) string {
	if message.Metadata != nil {
		if method, ok := message.Metadata["method"].(string); ok {
			return method
		}
	}
	return ""
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
