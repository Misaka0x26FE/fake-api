package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const helloWorld = "Hello, World!"

type ChatCompletionRequest struct {
	Model            string                       `json:"model"`
	Messages         []ChatCompletionMessage      `json:"messages"`
	Stream           bool                         `json:"stream,omitempty"`
	Temperature      *float64                     `json:"temperature,omitempty"`
	TopP             *float64                     `json:"top_p,omitempty"`
	N                *int                         `json:"n,omitempty"`
	MaxTokens        *int                         `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                       `json:"max_completion_tokens,omitempty"`
	Stop             any                          `json:"stop,omitempty"`
	PresencePenalty  *float64                     `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64                     `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int               `json:"logit_bias,omitempty"`
	User             string                       `json:"user,omitempty"`
	ResponseFormat   *ResponseFormat              `json:"response_format,omitempty"`
	Tools            []Tool                       `json:"tools,omitempty"`
	ToolChoice       any                          `json:"tool_choice,omitempty"`
	Seed             *int                         `json:"seed,omitempty"`
	Logprobs         *bool                        `json:"logprobs,omitempty"`
	TopLogprobs      *int                         `json:"top_logprobs,omitempty"`
}

type ChatCompletionMessage struct {
	Role       string       `json:"role"`
	Content    any          `json:"content"`
	Name       string       `json:"name,omitempty"`
	ToolCalls  []ToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string            `json:"type"`
	Function *FunctionDef      `json:"function,omitempty"`
}

type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ResponseFormat struct {
	Type       string         `json:"type"`
	JSONSchema *JSONSchemaDef `json:"json_schema,omitempty"`
}

type JSONSchemaDef struct {
	Name   string `json:"name"`
	Schema any    `json:"schema"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

type ChatCompletionChoice struct {
	Index        int                    `json:"index"`
	Message      *ChatCompletionMessage `json:"message,omitempty"`
	Delta        *ChatCompletionDelta   `json:"delta,omitempty"`
	FinishReason *string                `json:"finish_reason"`
}

type ChatCompletionDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", handleChatCompletions)
	mux.HandleFunc("GET /v1/models", handleListModels)
	mux.HandleFunc("GET /v1/models/{id}", handleGetModel)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("fake-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "invalid_request_error", "Only POST method is allowed")
		return
	}

	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "Invalid JSON body: "+err.Error())
		return
	}

	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "messages is required and must be non-empty")
		return
	}

	id := generateID()
	created := time.Now().Unix()

	if req.Stream {
		handleStream(w, id, created, req.Model)
	} else {
		handleNonStream(w, id, created, req.Model)
	}
}

func handleNonStream(w http.ResponseWriter, id string, created int64, model string) {
	finishReason := "stop"
	resp := ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: &ChatCompletionMessage{
					Role:    "assistant",
					Content: helloWorld,
				},
				FinishReason: &finishReason,
			},
		},
		Usage: Usage{
			PromptTokens:     0,
			CompletionTokens: 3,
			TotalTokens:      3,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleStream(w http.ResponseWriter, id string, created int64, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	chunks := []ChatCompletionResponse{
		{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Delta: &ChatCompletionDelta{
						Role:    "assistant",
						Content: "",
					},
					FinishReason: nil,
				},
			},
		},
		{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Delta: &ChatCompletionDelta{
						Content: helloWorld,
					},
					FinishReason: nil,
				},
			},
		},
	}

	finishReason := "stop"
	finishChunk := ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   model,
		Choices: []ChatCompletionChoice{
			{
				Index:        0,
				Delta:        &ChatCompletionDelta{},
				FinishReason: &finishReason,
			},
		},
	}

	for _, chunk := range chunks {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		time.Sleep(50 * time.Millisecond)
	}

	data, _ := json.Marshal(finishChunk)
	fmt.Fprintf(w, "data: %s\n\n", data)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("chatcmpl-%x", b)
}

type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelListResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}

func handleListModels(w http.ResponseWriter, r *http.Request) {
	resp := ModelListResponse{
		Object: "list",
		Data: []ModelObject{
			{
				ID:      "hello",
				Object:  "model",
				Created: 1700000000,
				OwnedBy: "fake-api",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleGetModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id != "hello" {
		writeError(w, http.StatusNotFound, "invalid_request_error", "Model '"+id+"' not found")
		return
	}
	resp := ModelObject{
		ID:      "hello",
		Object:  "model",
		Created: 1700000000,
		OwnedBy: "fake-api",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, typ, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    typ,
		},
	})
}