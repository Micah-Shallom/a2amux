package a2amux

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"trpc.group/trpc-go/trpc-a2a-go/server"
)

type wrappedHandler struct {
	server    *server.A2AServer
	methodMap map[string]string
	logger    Logger
}

func (w *wrappedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	originalHandler := w.server.Handler()

	if w.shouldMapMethod(r) {
		w.handleWithMethodMapping(rw, r, originalHandler)
	} else {
		originalHandler.ServeHTTP(rw, r)
	}
}

func (w *wrappedHandler) shouldMapMethod(r *http.Request) bool {
	return r.Method == http.MethodPost &&
		(strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "/")
}

func (w *wrappedHandler) handleWithMethodMapping(rw http.ResponseWriter, r *http.Request, originalHandler http.Handler) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, `{"jsonrpc":"2.0","error":{"code":-32700,"message":"Failed to read request body"},"id":null}`, http.StatusBadRequest)
		return
	}
	r.Body.Close()

	var jsonrpcReq map[string]interface{}
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		w.forwardRequest(rw, r, body, originalHandler)
		return
	}

	// Extract and map JSON-RPC method
	var originalMethod string
	if methodInterface, exists := jsonrpcReq["method"]; exists {
		if method, ok := methodInterface.(string); ok {
			originalMethod = method
			if mappedMethod, exists := w.methodMap[method]; exists {
				w.logger.Printf("Mapping method '%s' -> '%s'", method, mappedMethod)
				jsonrpcReq["method"] = mappedMethod
			}
		}
	}

	// Inject method into message.metadata if missing
	if params, exists := jsonrpcReq["params"].(map[string]interface{}); exists {
		if message, exists := params["message"].(map[string]interface{}); exists {
			if metadata, exists := message["metadata"].(map[string]interface{}); exists {
				if _, hasMethod := metadata["method"]; !hasMethod {
					metadata["method"] = originalMethod
					message["metadata"] = metadata
					params["message"] = message
					jsonrpcReq["params"] = params
					w.logger.Printf("Injected method '%s' into message.metadata", originalMethod)
				}
			} else {
				// Create metadata if it doesn't exist
				message["metadata"] = map[string]interface{}{"method": originalMethod}
				params["message"] = message
				jsonrpcReq["params"] = params
				w.logger.Printf("Created metadata with method '%s'", originalMethod)
			}
		}
	}

	// Re-marshal the modified request
	modifiedBody, err := json.Marshal(jsonrpcReq)
	if err != nil {
		http.Error(rw, "Failed to marshal modified request", http.StatusInternalServerError)
		return
	}

	w.forwardRequest(rw, r, modifiedBody, originalHandler)
}

func (w *wrappedHandler) forwardRequest(rw http.ResponseWriter, r *http.Request, body []byte, originalHandler http.Handler) {
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	originalHandler.ServeHTTP(rw, r)
}

func (m *Multiplexer) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a gin context for compatibility
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		middleware := m.GinMiddleware()
		middleware(c)
	})
}
