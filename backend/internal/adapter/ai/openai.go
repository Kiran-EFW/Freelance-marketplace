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
	openaiDefaultBaseURL = "https://api.openai.com/v1"
	openaiModel          = "gpt-4o"
	openaiMaxRetries     = 3
	openaiInitialBackoff = 500 * time.Millisecond
)

// OpenAIClient provides methods for interacting with the OpenAI API.
// It implements the same ClaudeProvider interface so it can be used as
// a drop-in replacement for chat and vision capabilities.
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI API client.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: openaiDefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// -- Internal request/response types --

type openaiRequest struct {
	Model      string          `json:"model"`
	Messages   []openaiMessage `json:"messages"`
	MaxTokens  int             `json:"max_tokens,omitempty"`
	Tools      []openaiTool    `json:"tools,omitempty"`
	ToolChoice string          `json:"tool_choice,omitempty"`
}

type openaiMessage struct {
	Role       string `json:"role"`
	Content    any    `json:"content"` // string or []openaiContentPart
	ToolCalls  []openaiToolCallResp `json:"tool_calls,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type openaiContentPart struct {
	Type     string             `json:"type"`
	Text     string             `json:"text,omitempty"`
	ImageURL *openaiImageURL    `json:"image_url,omitempty"`
}

type openaiImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type openaiTool struct {
	Type     string         `json:"type"` // "function"
	Function openaiFunction `json:"function"`
}

type openaiFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// -- Response types --

type openaiResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openaiChoice struct {
	Index        int                 `json:"index"`
	Message      openaiRespMessage   `json:"message"`
	FinishReason string              `json:"finish_reason"`
}

type openaiRespMessage struct {
	Role      string               `json:"role"`
	Content   *string              `json:"content"`
	ToolCalls []openaiToolCallResp `json:"tool_calls,omitempty"`
}

type openaiToolCallResp struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

type openaiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Chat sends a conversation with optional tool use to OpenAI and returns the response.
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	oaiMessages := make([]openaiMessage, len(messages))
	for i, m := range messages {
		oaiMessages[i] = openaiMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openaiRequest{
		Model:     openaiModel,
		MaxTokens: 4096,
		Messages:  oaiMessages,
	}

	if len(tools) > 0 {
		req.Tools = make([]openaiTool, len(tools))
		for i, t := range tools {
			req.Tools[i] = openaiTool{
				Type: "function",
				Function: openaiFunction{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.InputSchema,
				},
			}
		}
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}

	return c.parseResponse(resp), nil
}

// AnalyzeImage sends an image to OpenAI GPT-4o for vision analysis.
func (c *OpenAIClient) AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (*Response, error) {
	contentParts := []openaiContentPart{
		{
			Type: "image_url",
			ImageURL: &openaiImageURL{
				URL:    "data:image/jpeg;base64," + imageBase64,
				Detail: "auto",
			},
		},
		{
			Type: "text",
			Text: prompt,
		},
	}

	req := openaiRequest{
		Model:     openaiModel,
		MaxTokens: 4096,
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: contentParts,
			},
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai analyze image: %w", err)
	}

	return c.parseResponse(resp), nil
}

// GenerateContent generates text content from a prompt.
func (c *OpenAIClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}

	resp, err := c.Chat(ctx, messages, nil)
	if err != nil {
		return "", fmt.Errorf("openai generate content: %w", err)
	}

	return resp.Content, nil
}

// doRequest performs the HTTP request to the OpenAI API with retry and backoff.
func (c *OpenAIClient) doRequest(ctx context.Context, reqBody openaiRequest) (*openaiResponse, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < openaiMaxRetries; attempt++ {
		if attempt > 0 {
			backoff := openaiInitialBackoff * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request: %w", err)
			log.Warn().Err(lastErr).Int("attempt", attempt+1).Msg("openai API request failed, retrying")
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("openai API returned status %d: %s", resp.StatusCode, string(respBody))
			log.Warn().Int("status", resp.StatusCode).Int("attempt", attempt+1).Msg("openai API returned retryable error")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errResp openaiErrorResponse
			if err := json.Unmarshal(respBody, &errResp); err == nil {
				return nil, fmt.Errorf("openai API error (%s): %s", errResp.Error.Type, errResp.Error.Message)
			}
			return nil, fmt.Errorf("openai API returned status %d: %s", resp.StatusCode, string(respBody))
		}

		var result openaiResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		log.Debug().
			Str("model", result.Model).
			Int("prompt_tokens", result.Usage.PromptTokens).
			Int("completion_tokens", result.Usage.CompletionTokens).
			Msg("openai API call completed")

		return &result, nil
	}

	return nil, fmt.Errorf("openai API request failed after %d retries: %w", openaiMaxRetries, lastErr)
}

// parseResponse converts the OpenAI API response to the shared Response type.
func (c *OpenAIClient) parseResponse(resp *openaiResponse) *Response {
	result := &Response{}

	if len(resp.Choices) == 0 {
		return result
	}

	choice := resp.Choices[0]
	if choice.Message.Content != nil {
		result.Content = *choice.Message.Content
	}

	for _, tc := range choice.Message.ToolCalls {
		var input map[string]any
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
			log.Warn().Err(err).Str("function", tc.Function.Name).Msg("failed to parse tool call arguments")
			continue
		}
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			Name:  tc.Function.Name,
			Input: input,
		})
	}

	return result
}
