package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	thinkingWord = "<think>"
	spaceCount   = 60
	modelName    = "xujiayin"
)

var fullContent = thinkingWord + strings.Repeat(" ", spaceCount)

type ChatCompletionRequest struct {
	Model               string                  `json:"model"`
	Messages            []ChatCompletionMessage `json:"messages"`
	Stream              bool                    `json:"stream,omitempty"`
	Temperature         *float64                `json:"temperature,omitempty"`
	TopP                *float64                `json:"top_p,omitempty"`
	N                   *int                    `json:"n,omitempty"`
	MaxTokens           *int                    `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                    `json:"max_completion_tokens,omitempty"`
	Stop                any                     `json:"stop,omitempty"`
	PresencePenalty     *float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64                `json:"frequency_penalty,omitempty"`
	LogitBias           map[string]int          `json:"logit_bias,omitempty"`
	User                string                  `json:"user,omitempty"`
	ResponseFormat      *ResponseFormat         `json:"response_format,omitempty"`
	Tools               []Tool                  `json:"tools,omitempty"`
	ToolChoice          any                     `json:"tool_choice,omitempty"`
	Seed                *int                    `json:"seed,omitempty"`
	Logprobs            *bool                   `json:"logprobs,omitempty"`
	TopLogprobs         *int                    `json:"top_logprobs,omitempty"`
}

type ChatCompletionMessage struct {
	Role       string     `json:"role"`
	Content    any        `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function *FunctionDef `json:"function,omitempty"`
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

type ResponsesRequest struct {
	Model  string          `json:"model"`
	Input  json.RawMessage `json:"input,omitempty"`
	Stream bool            `json:"stream,omitempty"`
}

type ResponsesResponse struct {
	ID        string                `json:"id"`
	Object    string                `json:"object"`
	CreatedAt int64                 `json:"created_at"`
	Status    string                `json:"status"`
	Model     string                `json:"model"`
	Output    []ResponsesOutputItem `json:"output"`
	Usage     *ResponsesUsage       `json:"usage"`
}

type ResponsesOutputItem struct {
	ID      string                `json:"id"`
	Type    string                `json:"type"`
	Status  string                `json:"status"`
	Role    string                `json:"role"`
	Content []ResponsesOutputText `json:"content"`
}

type ResponsesOutputText struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
	Logprobs    []any  `json:"logprobs"`
}

type ResponsesUsage struct {
	InputTokens         int                          `json:"input_tokens"`
	InputTokensDetails  ResponsesInputTokensDetails  `json:"input_tokens_details"`
	OutputTokens        int                          `json:"output_tokens"`
	OutputTokensDetails ResponsesOutputTokensDetails `json:"output_tokens_details"`
	TotalTokens         int                          `json:"total_tokens"`
}

type ResponsesInputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type ResponsesOutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type MessagesRequest struct {
	Model     string                 `json:"model"`
	Messages  []MessagesInputMessage `json:"messages"`
	MaxTokens *int                   `json:"max_tokens"`
	Stream    bool                   `json:"stream,omitempty"`
}

type MessagesInputMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type AnthropicMessageResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Model        string                  `json:"model"`
	Content      []AnthropicContentBlock `json:"content"`
	StopReason   any                     `json:"stop_reason"`
	StopSequence any                     `json:"stop_sequence"`
	Usage        AnthropicUsage          `json:"usage"`
}

type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicErrorResponse struct {
	Type  string               `json:"type"`
	Error anthropicErrorDetail `json:"error"`
}

type anthropicErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type sseStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
	encoder *json.Encoder
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", handleChatCompletions)
	mux.HandleFunc("POST /v1/responses", handleResponses)
	mux.HandleFunc("POST /v1/messages", handleMessages)
	mux.HandleFunc("GET /v1/models", handleListModels)
	mux.HandleFunc("GET /v1/models/{id}", handleGetModel)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := "0.0.0.0:8080"
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

func handleResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "invalid_request_error", "Only POST method is allowed")
		return
	}

	var req ResponsesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "Invalid JSON body: "+err.Error())
		return
	}
	if req.Model == "" {
		writeError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}

	id := generateAPIID("resp_")
	created := time.Now().Unix()
	messageID := generateAPIID("msg_")
	if req.Stream {
		handleResponsesStream(w, r.Context().Done(), id, messageID, created, req.Model)
		return
	}
	handleResponsesNonStream(w, id, messageID, created, req.Model)
}

func handleResponsesNonStream(w http.ResponseWriter, id, messageID string, created int64, model string) {
	resp := buildResponsesResponse(id, messageID, created, model, "completed", fullContent)
	writeJSON(w, http.StatusOK, resp)
}

func handleResponsesStream(w http.ResponseWriter, ctx contextDone, id, messageID string, created int64, model string) {
	stream, ok := newSSEStream(w)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	inProgress := buildResponsesResponse(id, messageID, created, model, "in_progress", "")
	stream.send("response.created", map[string]any{
		"type":     "response.created",
		"response": inProgress,
	})
	stream.send("response.in_progress", map[string]any{
		"type":     "response.in_progress",
		"response": inProgress,
	})

	inProgressItem := buildResponsesOutputItem(messageID, "in_progress", "")
	stream.send("response.output_item.added", map[string]any{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item":         inProgressItem,
	})
	stream.send("response.content_part.added", map[string]any{
		"type":          "response.content_part.added",
		"item_id":       messageID,
		"output_index":  0,
		"content_index": 0,
		"part":          buildResponsesOutputText(""),
	})

	sequence := 0
	sendDelta := func(delta string) {
		sequence++
		stream.send("response.output_text.delta", map[string]any{
			"type":            "response.output_text.delta",
			"delta":           delta,
			"item_id":         messageID,
			"output_index":    0,
			"content_index":   0,
			"sequence_number": sequence,
		})
	}

	sendDelta(thinkingWord)
	for i := 0; i < spaceCount; i++ {
		if !waitForSpace(ctx) {
			return
		}
		sendDelta(" ")
	}

	completed := buildResponsesResponse(id, messageID, created, model, "completed", fullContent)
	completedItem := completed.Output[0]
	stream.send("response.output_text.done", map[string]any{
		"type":            "response.output_text.done",
		"text":            fullContent,
		"item_id":         messageID,
		"output_index":    0,
		"content_index":   0,
		"sequence_number": sequence + 1,
	})
	stream.send("response.content_part.done", map[string]any{
		"type":          "response.content_part.done",
		"item_id":       messageID,
		"output_index":  0,
		"content_index": 0,
		"part":          completedItem.Content[0],
	})
	stream.send("response.output_item.done", map[string]any{
		"type":         "response.output_item.done",
		"output_index": 0,
		"item":         completedItem,
	})
	stream.send("response.completed", map[string]any{
		"type":     "response.completed",
		"response": completed,
	})
}

func buildResponsesResponse(id, messageID string, created int64, model, status, content string) ResponsesResponse {
	resp := ResponsesResponse{
		ID:        id,
		Object:    "response",
		CreatedAt: created,
		Status:    status,
		Model:     model,
		Output:    []ResponsesOutputItem{},
		Usage:     nil,
	}
	if status == "completed" {
		resp.Output = []ResponsesOutputItem{buildResponsesOutputItem(messageID, "completed", content)}
		resp.Usage = buildResponsesUsage()
	}
	return resp
}

func buildResponsesOutputItem(id, status, content string) ResponsesOutputItem {
	contentBlocks := []ResponsesOutputText{}
	if status == "completed" {
		contentBlocks = append(contentBlocks, buildResponsesOutputText(content))
	}
	return ResponsesOutputItem{
		ID:      id,
		Type:    "message",
		Status:  status,
		Role:    "assistant",
		Content: contentBlocks,
	}
}

func buildResponsesOutputText(content string) ResponsesOutputText {
	return ResponsesOutputText{
		Type:        "output_text",
		Text:        content,
		Annotations: []any{},
		Logprobs:    []any{},
	}
}

func buildResponsesUsage() *ResponsesUsage {
	return &ResponsesUsage{
		InputTokens:        0,
		InputTokensDetails: ResponsesInputTokensDetails{},
		OutputTokens:       3,
		OutputTokensDetails: ResponsesOutputTokensDetails{
			ReasoningTokens: 0,
		},
		TotalTokens: 3,
	}
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAnthropicError(w, http.StatusMethodNotAllowed, "invalid_request_error", "Only POST method is allowed")
		return
	}

	var req MessagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", "Invalid JSON body: "+err.Error())
		return
	}
	if req.Model == "" {
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", "messages is required and must be non-empty")
		return
	}
	if req.MaxTokens == nil {
		writeAnthropicError(w, http.StatusBadRequest, "invalid_request_error", "max_tokens is required")
		return
	}

	id := generateAPIID("msg_")
	if req.Stream {
		handleMessagesStream(w, r.Context().Done(), id, req.Model)
		return
	}
	handleMessagesNonStream(w, id, req.Model)
}

func handleMessagesNonStream(w http.ResponseWriter, id, model string) {
	resp := buildAnthropicMessage(id, model, fullContent, "end_turn", 3)
	writeJSON(w, http.StatusOK, resp)
}

func handleMessagesStream(w http.ResponseWriter, ctx contextDone, id, model string) {
	stream, ok := newSSEStream(w)
	if !ok {
		writeAnthropicError(w, http.StatusInternalServerError, "api_error", "Streaming not supported")
		return
	}

	start := buildAnthropicMessage(id, model, "", nil, 0)
	start.Content = []AnthropicContentBlock{}
	stream.send("message_start", map[string]any{
		"type":    "message_start",
		"message": start,
	})
	stream.send("content_block_start", map[string]any{
		"type":          "content_block_start",
		"index":         0,
		"content_block": AnthropicContentBlock{Type: "text", Text: ""},
	})

	sendDelta := func(text string) {
		stream.send("content_block_delta", map[string]any{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]string{
				"type": "text_delta",
				"text": text,
			},
		})
	}

	sendDelta(thinkingWord)
	for i := 0; i < spaceCount; i++ {
		if !waitForSpace(ctx) {
			return
		}
		sendDelta(" ")
	}

	stream.send("content_block_stop", map[string]any{
		"type":  "content_block_stop",
		"index": 0,
	})
	stream.send("message_delta", map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   "end_turn",
			"stop_sequence": nil,
		},
		"usage": AnthropicUsage{InputTokens: 0, OutputTokens: 3},
	})
	stream.send("message_stop", map[string]string{
		"type": "message_stop",
	})
}

func buildAnthropicMessage(id, model, content string, stopReason any, outputTokens int) AnthropicMessageResponse {
	return AnthropicMessageResponse{
		ID:    id,
		Type:  "message",
		Role:  "assistant",
		Model: model,
		Content: []AnthropicContentBlock{
			{Type: "text", Text: content},
		},
		StopReason:   stopReason,
		StopSequence: nil,
		Usage: AnthropicUsage{
			InputTokens:  0,
			OutputTokens: outputTokens,
		},
	}
}

type contextDone <-chan struct{}

func waitForSpace(done contextDone) bool {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-done:
		return false
	}
}

func newSSEStream(w http.ResponseWriter) (*sseStream, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	return &sseStream{
		w:       w,
		flusher: flusher,
		encoder: json.NewEncoder(w),
	}, true
}

func (s *sseStream) send(event string, payload any) {
	s.encoder.SetEscapeHTML(false)
	fmt.Fprintf(s.w, "event: %s\n", event)
	fmt.Fprint(s.w, "data: ")
	s.encoder.Encode(payload)
	fmt.Fprint(s.w, "\n")
	s.flusher.Flush()
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(value)
}

func writeAnthropicError(w http.ResponseWriter, status int, typ, message string) {
	writeJSON(w, status, anthropicErrorResponse{
		Type: "error",
		Error: anthropicErrorDetail{
			Type:    typ,
			Message: message,
		},
	})
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
					Content: fullContent,
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
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(resp)
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
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	send := func(content, role string) {
		chunk := ChatCompletionResponse{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   model,
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Delta: &ChatCompletionDelta{
						Role:    role,
						Content: content,
					},
					FinishReason: nil,
				},
			},
		}
		fmt.Fprint(w, "data: ")
		encoder.Encode(chunk)
		fmt.Fprint(w, "\n")
		flusher.Flush()
	}

	send("", "assistant")
	send(thinkingWord, "")
	for i := 0; i < spaceCount; i++ {
		time.Sleep(time.Second)
		send(" ", "")
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

	fmt.Fprint(w, "data: ")
	encoder.Encode(finishChunk)
	fmt.Fprint(w, "\n")
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func generateID() string {
	return generateAPIID("chatcmpl-")
}

func generateAPIID(prefix string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s%x", prefix, b)
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
				ID:      modelName,
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
	if id != modelName {
		writeError(w, http.StatusNotFound, "invalid_request_error", "Model '"+id+"' not found")
		return
	}
	resp := ModelObject{
		ID:      modelName,
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
