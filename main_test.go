package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponsesNonStream(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"xujiayin","input":"hi"}`))
	recorder := httptest.NewRecorder()

	handleResponses(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response ResponsesResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "response" || response.Status != "completed" {
		t.Fatalf("unexpected response metadata: %+v", response)
	}
	if got := response.Output[0].Content[0].Text; got != fullContent {
		t.Fatalf("content = %q, want %q", got, fullContent)
	}
}

func TestMessagesNonStream(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"xujiayin","max_tokens":64,"messages":[{"role":"user","content":"hi"}]}`))
	recorder := httptest.NewRecorder()

	handleMessages(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response AnthropicMessageResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Type != "message" || response.StopReason != "end_turn" {
		t.Fatalf("unexpected message metadata: %+v", response)
	}
	if got := response.Content[0].Text; got != fullContent {
		t.Fatalf("content = %q, want %q", got, fullContent)
	}
}

func TestResponsesStreamStarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"xujiayin","input":"hi","stream":true}`)).WithContext(ctx)
	recorder := httptest.NewRecorder()

	handleResponses(recorder, req)

	body := recorder.Body.String()
	for _, want := range []string{
		"event: response.created",
		"event: response.output_text.delta",
		`"delta":"<think>"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("stream body does not contain %q: %s", want, body)
		}
	}
}

func TestMessagesStreamStarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"xujiayin","max_tokens":64,"messages":[{"role":"user","content":"hi"}],"stream":true}`)).WithContext(ctx)
	recorder := httptest.NewRecorder()

	handleMessages(recorder, req)

	body := recorder.Body.String()
	for _, want := range []string{
		"event: message_start",
		"event: content_block_delta",
		`"text":"<think>"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("stream body does not contain %q: %s", want, body)
		}
	}
}
