package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	googleTranslateEndpoint = "https://translation.googleapis.com/language/translate/v2"
	googleDetectEndpoint    = "https://translation.googleapis.com/language/translate/v2/detect"
)

// GoogleTranslateProvider defines the interface for translation operations.
type GoogleTranslateProvider interface {
	Translate(ctx context.Context, text string, sourceLang, targetLang string) (string, error)
	DetectLanguage(ctx context.Context, text string) (string, float64, error)
	TranslateBatch(ctx context.Context, texts []string, sourceLang, targetLang string) ([]string, error)
}

// GoogleTranslateClient provides translation via the Google Cloud Translation API.
type GoogleTranslateClient struct {
	credentialsJSON string
	apiKey          string
	httpClient      *http.Client
}

// NewGoogleTranslateClient creates a new Google Cloud Translation client.
func NewGoogleTranslateClient(credentialsJSON string) *GoogleTranslateClient {
	return &GoogleTranslateClient{
		credentialsJSON: credentialsJSON,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// translateRequest is the request body for the Google Translate API.
type translateRequest struct {
	Q      interface{} `json:"q"`      // string or []string
	Source string      `json:"source,omitempty"`
	Target string      `json:"target"`
	Format string      `json:"format"`
}

// translateResponse is the response body from the Google Translate API.
type translateResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText         string `json:"translatedText"`
			DetectedSourceLanguage string `json:"detectedSourceLanguage,omitempty"`
		} `json:"translations"`
	} `json:"data"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// detectRequest is the request body for the language detection API.
type detectRequest struct {
	Q string `json:"q"`
}

// detectResponse is the response body from the language detection API.
type detectResponse struct {
	Data struct {
		Detections [][]struct {
			Language   string  `json:"language"`
			Confidence float64 `json:"confidence"`
			IsReliable bool    `json:"isReliable"`
		} `json:"detections"`
	} `json:"data"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Translate translates text from the source language to the target language.
// If sourceLang is empty, the API will auto-detect the source language.
func (g *GoogleTranslateClient) Translate(ctx context.Context, text string, sourceLang, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	reqBody := translateRequest{
		Q:      text,
		Source: sourceLang,
		Target: targetLang,
		Format: "text",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("google translate marshal request: %w", err)
	}

	url := googleTranslateEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("google translate create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google translate http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("google translate read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google translate API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var transResp translateResponse
	if err := json.Unmarshal(respBody, &transResp); err != nil {
		return "", fmt.Errorf("google translate unmarshal response: %w", err)
	}

	if transResp.Error != nil {
		return "", fmt.Errorf("google translate API error (%d): %s", transResp.Error.Code, transResp.Error.Message)
	}

	if len(transResp.Data.Translations) == 0 {
		return "", fmt.Errorf("google translate: no translations returned")
	}

	translated := transResp.Data.Translations[0].TranslatedText

	log.Debug().
		Str("source", sourceLang).
		Str("target", targetLang).
		Int("input_len", len(text)).
		Int("output_len", len(translated)).
		Msg("google translate completed")

	return translated, nil
}

// DetectLanguage detects the language of the given text and returns the language
// code and confidence score.
func (g *GoogleTranslateClient) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	if text == "" {
		return "", 0, fmt.Errorf("google translate detect language: empty text")
	}

	reqBody := detectRequest{Q: text}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("google translate detect marshal request: %w", err)
	}

	url := googleDetectEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("google translate detect create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("google translate detect http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("google translate detect read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("google translate detect API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var detectResp detectResponse
	if err := json.Unmarshal(respBody, &detectResp); err != nil {
		return "", 0, fmt.Errorf("google translate detect unmarshal response: %w", err)
	}

	if detectResp.Error != nil {
		return "", 0, fmt.Errorf("google translate detect API error (%d): %s", detectResp.Error.Code, detectResp.Error.Message)
	}

	if len(detectResp.Data.Detections) == 0 || len(detectResp.Data.Detections[0]) == 0 {
		return "", 0, fmt.Errorf("google translate detect: no detections returned")
	}

	detection := detectResp.Data.Detections[0][0]

	log.Debug().
		Str("language", detection.Language).
		Float64("confidence", detection.Confidence).
		Msg("google translate language detection completed")

	return detection.Language, detection.Confidence, nil
}

// TranslateBatch translates multiple texts from the source language to the target
// language in a single API call.
func (g *GoogleTranslateClient) TranslateBatch(ctx context.Context, texts []string, sourceLang, targetLang string) ([]string, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := translateRequest{
		Q:      texts,
		Source: sourceLang,
		Target: targetLang,
		Format: "text",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("google translate batch marshal request: %w", err)
	}

	url := googleTranslateEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("google translate batch create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google translate batch http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google translate batch read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google translate batch API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var transResp translateResponse
	if err := json.Unmarshal(respBody, &transResp); err != nil {
		return nil, fmt.Errorf("google translate batch unmarshal response: %w", err)
	}

	if transResp.Error != nil {
		return nil, fmt.Errorf("google translate batch API error (%d): %s", transResp.Error.Code, transResp.Error.Message)
	}

	results := make([]string, len(transResp.Data.Translations))
	for i, t := range transResp.Data.Translations {
		results[i] = t.TranslatedText
	}

	log.Debug().
		Str("source", sourceLang).
		Str("target", targetLang).
		Int("count", len(results)).
		Msg("google translate batch completed")

	return results, nil
}
