package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	claudeDefaultBaseURL = "https://api.anthropic.com/v1"
	claudeModel          = "claude-sonnet-4-20250514"
	claudeAPIVersion     = "2023-06-01"
	claudeMaxRetries     = 3
	claudeInitialBackoff = 500 * time.Millisecond
)

// Message represents a single message in a Claude conversation.
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // text content
}

// Tool describes a tool that Claude can invoke during a conversation.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// ToolCall represents a tool invocation returned by Claude.
type ToolCall struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// Response holds the result of a Claude API call.
type Response struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ClaudeClient provides methods for interacting with the Claude API.
type ClaudeClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClaudeProvider defines the interface for Claude AI operations.
type ClaudeProvider interface {
	Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
	AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (*Response, error)
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// NewClaudeClient creates a new Claude API client.
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:  apiKey,
		baseURL: claudeDefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// claudeRequest is the request body sent to the Claude messages API.
type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	Messages  []claudeMessage  `json:"messages"`
	Tools     []claudeToolDef  `json:"tools,omitempty"`
	System    string           `json:"system,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []claudeContentBlock
}

type claudeContentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Source    *claudeImageSource `json:"source,omitempty"`
}

type claudeImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png", etc.
	Data      string `json:"data"`
}

type claudeToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// claudeResponse is the response body from the Claude messages API.
type claudeResponse struct {
	ID      string               `json:"id"`
	Type    string               `json:"type"`
	Role    string               `json:"role"`
	Content []claudeResponseBlock `json:"content"`
	Model   string               `json:"model"`
	StopReason string            `json:"stop_reason"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type claudeResponseBlock struct {
	Type  string         `json:"type"`
	Text  string         `json:"text,omitempty"`
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
}

type claudeErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a conversation with optional tool use to Claude and returns the response.
func (c *ClaudeClient) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	claudeMessages := make([]claudeMessage, len(messages))
	for i, m := range messages {
		claudeMessages[i] = claudeMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 4096,
		Messages:  claudeMessages,
	}

	if len(tools) > 0 {
		req.Tools = make([]claudeToolDef, len(tools))
		for i, t := range tools {
			req.Tools[i] = claudeToolDef{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			}
		}
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("claude chat: %w", err)
	}

	return c.parseResponse(resp), nil
}

// AnalyzeImage sends an image to Claude for vision analysis.
func (c *ClaudeClient) AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (*Response, error) {
	contentBlocks := []claudeContentBlock{
		{
			Type: "image",
			Source: &claudeImageSource{
				Type:      "base64",
				MediaType: "image/jpeg",
				Data:      imageBase64,
			},
		},
		{
			Type: "text",
			Text: prompt,
		},
	}

	req := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 4096,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: contentBlocks,
			},
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("claude analyze image: %w", err)
	}

	return c.parseResponse(resp), nil
}

// GenerateContent generates text content (e.g., SEO descriptions, landing pages)
// from a prompt.
func (c *ClaudeClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}

	resp, err := c.Chat(ctx, messages, nil)
	if err != nil {
		return "", fmt.Errorf("claude generate content: %w", err)
	}

	return resp.Content, nil
}

// doRequest performs the HTTP request to the Claude API with retry and backoff.
func (c *ClaudeClient) doRequest(ctx context.Context, reqBody claudeRequest) (*claudeResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < claudeMaxRetries; attempt++ {
		if attempt > 0 {
			backoff := claudeInitialBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Anthropic-Version", claudeAPIVersion)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			log.Warn().Err(lastErr).Int("attempt", attempt+1).Msg("claude API request failed, retrying")
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("claude API returned status %d: %s", resp.StatusCode, string(respBody))
			log.Warn().Int("status", resp.StatusCode).Int("attempt", attempt+1).Msg("claude API returned retryable error")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errResp claudeErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				return nil, fmt.Errorf("claude API error (%s): %s", errResp.Error.Type, errResp.Error.Message)
			}
			return nil, fmt.Errorf("claude API returned status %d: %s", resp.StatusCode, string(respBody))
		}

		var result claudeResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		log.Debug().
			Str("model", result.Model).
			Int("input_tokens", result.Usage.InputTokens).
			Int("output_tokens", result.Usage.OutputTokens).
			Msg("claude API call completed")

		return &result, nil
	}

	return nil, fmt.Errorf("claude API request failed after %d retries: %w", claudeMaxRetries, lastErr)
}

// parseResponse converts the internal Claude API response to the public Response type.
func (c *ClaudeClient) parseResponse(resp *claudeResponse) *Response {
	result := &Response{}

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}

	return result
}
