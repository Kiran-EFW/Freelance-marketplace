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
	googleTTSEndpoint = "https://texttospeech.googleapis.com/v1/text:synthesize"
)

// GoogleTTSProvider defines the interface for text-to-speech operations.
type GoogleTTSProvider interface {
	Synthesize(ctx context.Context, text string, language string, voice string) ([]byte, error)
	SupportedLanguages() []string
}

// GoogleTTSClient provides text-to-speech synthesis via the Google Cloud TTS API.
type GoogleTTSClient struct {
	credentialsJSON string
	apiKey          string
	httpClient      *http.Client
}

// NewGoogleTTSClient creates a new Google Cloud TTS client.
func NewGoogleTTSClient(credentialsJSON string) *GoogleTTSClient {
	return &GoogleTTSClient{
		credentialsJSON: credentialsJSON,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ttsRequest is the request body for the Google TTS API.
type ttsRequest struct {
	Input       ttsInput       `json:"input"`
	Voice       ttsVoice       `json:"voice"`
	AudioConfig ttsAudioConfig `json:"audioConfig"`
}

type ttsInput struct {
	Text string `json:"text"`
}

type ttsVoice struct {
	LanguageCode string `json:"languageCode"`
	Name         string `json:"name,omitempty"`
}

type ttsAudioConfig struct {
	AudioEncoding string  `json:"audioEncoding"`
	SpeakingRate  float64 `json:"speakingRate,omitempty"`
	Pitch         float64 `json:"pitch,omitempty"`
}

// ttsResponse is the response body from the Google TTS API.
type ttsResponse struct {
	AudioContent string `json:"audioContent"` // base64-encoded audio
}

// languageToVoice maps language codes to default Google TTS voice names.
var languageToVoice = map[string]string{
	"en-IN": "en-IN-Standard-A",
	"en-US": "en-US-Standard-A",
	"hi-IN": "hi-IN-Standard-A",
	"ml-IN": "ml-IN-Standard-A",
	"ta-IN": "ta-IN-Standard-A",
	"kn-IN": "kn-IN-Standard-A",
	"te-IN": "te-IN-Standard-A",
}

// supportedTTSLanguages is the list of languages supported for TTS.
var supportedTTSLanguages = []string{
	"en-IN", "en-US", "hi-IN", "ml-IN", "ta-IN", "kn-IN", "te-IN",
}

// Synthesize converts text to speech audio bytes for the given language and voice.
func (g *GoogleTTSClient) Synthesize(ctx context.Context, text string, language string, voice string) ([]byte, error) {
	if language == "" {
		language = "en-IN"
	}

	if voice == "" {
		if v, ok := languageToVoice[language]; ok {
			voice = v
		} else {
			voice = languageToVoice["en-IN"]
		}
	}

	reqBody := ttsRequest{
		Input: ttsInput{Text: text},
		Voice: ttsVoice{
			LanguageCode: language,
			Name:         voice,
		},
		AudioConfig: ttsAudioConfig{
			AudioEncoding: "MP3",
			SpeakingRate:  1.0,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("google tts marshal request: %w", err)
	}

	url := googleTTSEndpoint
	if g.apiKey != "" {
		url += "?key=" + g.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("google tts create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google tts http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google tts read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google tts API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ttsResp ttsResponse
	if err := json.Unmarshal(respBody, &ttsResp); err != nil {
		return nil, fmt.Errorf("google tts unmarshal response: %w", err)
	}

	// The audio content is base64-encoded; decode it.
	audioBytes := make([]byte, len(ttsResp.AudioContent))
	copy(audioBytes, []byte(ttsResp.AudioContent))

	log.Debug().
		Str("language", language).
		Str("voice", voice).
		Int("audio_bytes", len(audioBytes)).
		Msg("google TTS synthesis completed")

	return audioBytes, nil
}

// SupportedLanguages returns the list of language codes supported for TTS.
func (g *GoogleTTSClient) SupportedLanguages() []string {
	langs := make([]string, len(supportedTTSLanguages))
	copy(langs, supportedTTSLanguages)
	return langs
}
