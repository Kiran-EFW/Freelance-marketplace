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
	deepgramBaseURL = "https://api.deepgram.com/v1"
)

// Word represents a single transcribed word with timing information.
type Word struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence"`
}

// Transcription holds the result of a speech-to-text operation.
type Transcription struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
	Words      []Word  `json:"words"`
}

// TranscriptionChunk represents a partial transcription from a streaming session.
type TranscriptionChunk struct {
	Text       string  `json:"text"`
	IsFinal    bool    `json:"is_final"`
	Confidence float64 `json:"confidence"`
}

// DeepgramProvider defines the interface for speech-to-text operations.
type DeepgramProvider interface {
	Transcribe(ctx context.Context, audioData []byte, language string) (*Transcription, error)
	TranscribeStream(ctx context.Context, audioStream io.Reader, language string) (<-chan TranscriptionChunk, error)
}

// DeepgramClient provides speech-to-text transcription via the Deepgram API.
type DeepgramClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewDeepgramClient creates a new Deepgram API client.
func NewDeepgramClient(apiKey string) *DeepgramClient {
	return &DeepgramClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 300 * time.Second, // audio processing can be slow
		},
	}
}

// deepgramResponse represents the Deepgram API transcription response.
type deepgramResponse struct {
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string  `json:"transcript"`
				Confidence float64 `json:"confidence"`
				Words      []struct {
					Word       string  `json:"word"`
					Start      float64 `json:"start"`
					End        float64 `json:"end"`
					Confidence float64 `json:"confidence"`
				} `json:"words"`
			} `json:"alternatives"`
			DetectedLanguage string `json:"detected_language"`
		} `json:"channels"`
	} `json:"results"`
}

// supportedIndianLanguages maps language codes to Deepgram model identifiers
// for Indian languages.
var supportedIndianLanguages = map[string]bool{
	"hi": true, // Hindi
	"ml": true, // Malayalam
	"ta": true, // Tamil
	"kn": true, // Kannada
	"te": true, // Telugu
	"en": true, // English (default)
}

// Transcribe sends audio data to Deepgram for transcription.
func (d *DeepgramClient) Transcribe(ctx context.Context, audioData []byte, language string) (*Transcription, error) {
	if language == "" {
		language = "en"
	}

	if !supportedIndianLanguages[language] {
		log.Warn().Str("language", language).Msg("unsupported language, falling back to English")
		language = "en"
	}

	url := fmt.Sprintf("%s/listen?language=%s&punctuate=true&model=nova-2", deepgramBaseURL, language)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(audioData))
	if err != nil {
		return nil, fmt.Errorf("deepgram create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+d.apiKey)
	req.Header.Set("Content-Type", "audio/wav")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deepgram http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("deepgram read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deepgram API returned status %d: %s", resp.StatusCode, string(body))
	}

	var dgResp deepgramResponse
	if err := json.Unmarshal(body, &dgResp); err != nil {
		return nil, fmt.Errorf("deepgram unmarshal response: %w", err)
	}

	result := &Transcription{
		Language: language,
	}

	if len(dgResp.Results.Channels) > 0 && len(dgResp.Results.Channels[0].Alternatives) > 0 {
		alt := dgResp.Results.Channels[0].Alternatives[0]
		result.Text = alt.Transcript
		result.Confidence = alt.Confidence

		for _, w := range alt.Words {
			result.Words = append(result.Words, Word{
				Word:       w.Word,
				Start:      w.Start,
				End:        w.End,
				Confidence: w.Confidence,
			})
		}

		if dgResp.Results.Channels[0].DetectedLanguage != "" {
			result.Language = dgResp.Results.Channels[0].DetectedLanguage
		}
	}

	log.Debug().
		Str("language", result.Language).
		Float64("confidence", result.Confidence).
		Int("word_count", len(result.Words)).
		Msg("deepgram transcription completed")

	return result, nil
}

// TranscribeStream opens a streaming transcription session. Audio data is read
// from the provided io.Reader, and partial transcription results are delivered
// on the returned channel.
func (d *DeepgramClient) TranscribeStream(ctx context.Context, audioStream io.Reader, language string) (<-chan TranscriptionChunk, error) {
	if language == "" {
		language = "en"
	}

	ch := make(chan TranscriptionChunk, 64)

	go func() {
		defer close(ch)

		// Read audio in chunks and send for transcription.
		// In a production implementation this would use WebSocket streaming
		// to Deepgram's real-time API endpoint.
		buf := make([]byte, 8192)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := audioStream.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])

				transcription, tErr := d.Transcribe(ctx, chunk, language)
				if tErr != nil {
					log.Error().Err(tErr).Msg("deepgram stream transcription error")
					continue
				}

				if transcription.Text != "" {
					select {
					case ch <- TranscriptionChunk{
						Text:       transcription.Text,
						IsFinal:    false,
						Confidence: transcription.Confidence,
					}:
					case <-ctx.Done():
						return
					}
				}
			}

			if err == io.EOF {
				// Send final marker
				select {
				case ch <- TranscriptionChunk{IsFinal: true}:
				case <-ctx.Done():
				}
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("deepgram stream read error")
				return
			}
		}
	}()

	return ch, nil
}
