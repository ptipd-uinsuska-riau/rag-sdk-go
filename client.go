// Package ragsdk menyediakan SDK untuk integrasi dengan RAG AI Chatbot Universitas.
//
// Autentikasi menggunakan Vendor App Token dari sistem SDM.
//
//	client := ragsdk.NewClient("https://rag.universitas.ac.id/v1", "your-app-token")
//	resp, err := client.Chat(ctx, ragsdk.ChatRequest{Message: "Info pendaftaran"})
package ragsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client untuk berkomunikasi dengan RAG service.
type Client struct {
	baseURL    string
	appToken   string
	httpClient *http.Client
	language   string
}

// NewClient membuat RAG client baru.
func NewClient(ragBaseURL, appToken string, opts ...Option) *Client {
	c := &Client{
		baseURL:  strings.TrimRight(ragBaseURL, "/"),
		appToken: appToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		language: "id",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithLanguage mengatur bahasa default ('id' atau 'en').
func WithLanguage(lang string) Option {
	return func(c *Client) { c.language = lang }
}

// WithHTTPClient mengatur custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout mengatur timeout HTTP client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// Chat mengirim pesan dan menerima jawaban lengkap (non-streaming).
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Language == "" {
		req.Language = c.language
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ragsdk: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ragsdk: create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ragsdk: do request: %w", err)
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("ragsdk: decode response: %w", err)
	}
	return &out, nil
}

// Feedback mengirim rating untuk jawaban chatbot.
func (c *Client) Feedback(ctx context.Context, messageID string, rating Rating, comment string) error {
	body, err := json.Marshal(FeedbackRequest{Rating: ratingToInt(rating), FeedbackText: comment})
	if err != nil {
		return fmt.Errorf("ragsdk: marshal feedback: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/chat/%s/feedback", c.baseURL, messageID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ragsdk: create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ragsdk: do request: %w", err)
	}
	defer resp.Body.Close()

	return checkStatus(resp)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", c.language)
	req.Header.Set("X-App-Token", c.appToken)
}

func checkStatus(resp *http.Response) error {
	switch {
	case resp.StatusCode == 401:
		return ErrUnauthorized
	case resp.StatusCode == 429:
		return ErrRateLimited
	case resp.StatusCode >= 500:
		return ErrServerError
	case resp.StatusCode >= 400:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ragsdk: API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
