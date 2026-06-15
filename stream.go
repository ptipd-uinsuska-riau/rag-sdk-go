package ragsdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ChatStream mengirim pesan dan menerima jawaban via SSE streaming.
func (c *Client) ChatStream(ctx context.Context, req ChatRequest, handler StreamHandler) error {
	if req.Language == "" {
		req.Language = c.language
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ragsdk: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/chat/stream", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ragsdk: create request: %w", err)
	}

	c.setHeaders(httpReq)

	// Use a client with extended timeout for streaming
	streamClient := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		handler.OnError(err)
		return fmt.Errorf("ragsdk: do request: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		io.ReadAll(resp.Body) // Drain body to allow connection reuse
		handler.OnError(err)
		return err
	}

	var messageID string

	scanner := bufio.NewScanner(resp.Body)
	// SSE chunks can exceed the 64KB default token size.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		// Check context cancellation
		select {
		case <-ctx.Done():
			handler.OnError(ctx.Err())
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			handler.OnDone(messageID)
			return nil
		}

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.SessionID != "" {
			handler.OnSessionId(event.SessionID)
		}

		switch event.Type {
		case "chunk":
			handler.OnChunk(event.Content)
		case "sources":
			handler.OnSources(event.Sources)
		case "fallback":
			// rag-be still sends 'done' (with message_id) + [DONE] afterwards,
			// so surface the text and keep reading to the end.
			handler.OnChunk(event.Content)
		case "done":
			messageID = event.MessageID
		case "error":
			err := fmt.Errorf("ragsdk: stream error: %s", event.Message)
			handler.OnError(err)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		handler.OnError(err)
		return fmt.Errorf("ragsdk: read stream: %w", err)
	}

	handler.OnDone(messageID)
	return nil
}

// SimpleStreamHandler adalah implementasi sederhana StreamHandler.
type SimpleStreamHandler struct {
	ChunkFn     func(text string)
	SourcesFn   func(sources []ChatSource)
	SessionIdFn func(sessionID string)
	DoneFn      func(messageID string)
	ErrorFn     func(err error)
}

func (h *SimpleStreamHandler) OnChunk(text string) {
	if h.ChunkFn != nil {
		h.ChunkFn(text)
	}
}

func (h *SimpleStreamHandler) OnSources(sources []ChatSource) {
	if h.SourcesFn != nil {
		h.SourcesFn(sources)
	}
}

func (h *SimpleStreamHandler) OnSessionId(sessionID string) {
	if h.SessionIdFn != nil {
		h.SessionIdFn(sessionID)
	}
}

func (h *SimpleStreamHandler) OnDone(messageID string) {
	if h.DoneFn != nil {
		h.DoneFn(messageID)
	}
}

func (h *SimpleStreamHandler) OnError(err error) {
	if h.ErrorFn != nil {
		h.ErrorFn(err)
	}
}
