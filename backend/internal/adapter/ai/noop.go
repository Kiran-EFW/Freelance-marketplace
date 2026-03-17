package ai

import (
	"context"
	"io"

	"github.com/rs/zerolog/log"
)

// --- NoopClaudeClient ---

// NoopClaudeClient is a no-op implementation of ClaudeProvider for development
// and testing. It returns realistic mock data without making API calls.
type NoopClaudeClient struct{}

// NewNoopClaudeClient creates a new no-op Claude client.
func NewNoopClaudeClient() *NoopClaudeClient {
	return &NoopClaudeClient{}
}

// Chat returns a mock response with realistic content.
func (n *NoopClaudeClient) Chat(_ context.Context, messages []Message, tools []Tool) (*Response, error) {
	log.Debug().Int("messages", len(messages)).Int("tools", len(tools)).Msg("noop claude chat")

	resp := &Response{
		Content: "This is a mock response from the noop Claude client. In production, this would be a real AI-generated response based on the conversation context.",
	}

	if len(tools) > 0 {
		resp.ToolCalls = []ToolCall{
			{
				Name: tools[0].Name,
				Input: map[string]any{
					"mock": true,
				},
			},
		}
	}

	return resp, nil
}

// AnalyzeImage returns a mock image analysis response.
func (n *NoopClaudeClient) AnalyzeImage(_ context.Context, _ string, prompt string) (*Response, error) {
	log.Debug().Str("prompt", prompt).Msg("noop claude analyze image")
	return &Response{
		Content: "Mock image analysis: The image appears to contain a document. All text is legible and the document appears authentic. This is a placeholder response for development.",
	}, nil
}

// GenerateContent returns mock generated content.
func (n *NoopClaudeClient) GenerateContent(_ context.Context, prompt string) (string, error) {
	log.Debug().Str("prompt", prompt).Msg("noop claude generate content")
	return "Mock generated content: This is placeholder SEO-optimized content for the Seva service marketplace. In production, this would be AI-generated content tailored to the specific prompt.", nil
}

// --- NoopDeepgramClient ---

// NoopDeepgramClient is a no-op implementation of DeepgramProvider for development
// and testing.
type NoopDeepgramClient struct{}

// NewNoopDeepgramClient creates a new no-op Deepgram client.
func NewNoopDeepgramClient() *NoopDeepgramClient {
	return &NoopDeepgramClient{}
}

// Transcribe returns a mock transcription result.
func (n *NoopDeepgramClient) Transcribe(_ context.Context, audioData []byte, language string) (*Transcription, error) {
	log.Debug().Int("audio_bytes", len(audioData)).Str("language", language).Msg("noop deepgram transcribe")

	if language == "" {
		language = "en"
	}

	return &Transcription{
		Text:       "This is a mock transcription of the provided audio data.",
		Confidence: 0.95,
		Language:   language,
		Words: []Word{
			{Word: "This", Start: 0.0, End: 0.2, Confidence: 0.99},
			{Word: "is", Start: 0.2, End: 0.3, Confidence: 0.98},
			{Word: "a", Start: 0.3, End: 0.35, Confidence: 0.99},
			{Word: "mock", Start: 0.35, End: 0.6, Confidence: 0.97},
			{Word: "transcription", Start: 0.6, End: 1.1, Confidence: 0.96},
		},
	}, nil
}

// TranscribeStream returns a channel with a single mock transcription chunk.
func (n *NoopDeepgramClient) TranscribeStream(_ context.Context, _ io.Reader, language string) (<-chan TranscriptionChunk, error) {
	log.Debug().Str("language", language).Msg("noop deepgram transcribe stream")

	ch := make(chan TranscriptionChunk, 2)
	go func() {
		defer close(ch)
		ch <- TranscriptionChunk{
			Text:       "Mock streaming transcription chunk.",
			IsFinal:    false,
			Confidence: 0.93,
		}
		ch <- TranscriptionChunk{
			Text:    "",
			IsFinal: true,
		}
	}()

	return ch, nil
}

// --- NoopGoogleTTSClient ---

// NoopGoogleTTSClient is a no-op implementation of GoogleTTSProvider for
// development and testing.
type NoopGoogleTTSClient struct{}

// NewNoopGoogleTTSClient creates a new no-op Google TTS client.
func NewNoopGoogleTTSClient() *NoopGoogleTTSClient {
	return &NoopGoogleTTSClient{}
}

// Synthesize returns mock audio bytes (empty MP3-like header).
func (n *NoopGoogleTTSClient) Synthesize(_ context.Context, text string, language string, voice string) ([]byte, error) {
	log.Debug().Str("language", language).Str("voice", voice).Int("text_len", len(text)).Msg("noop google tts synthesize")

	// Return a minimal valid byte slice that represents "audio" for testing.
	mockAudio := make([]byte, 256)
	// MP3 sync word (very simplified mock)
	mockAudio[0] = 0xFF
	mockAudio[1] = 0xFB

	return mockAudio, nil
}

// SupportedLanguages returns the list of supported TTS languages.
func (n *NoopGoogleTTSClient) SupportedLanguages() []string {
	return []string{"en-IN", "en-US", "hi-IN", "ml-IN", "ta-IN", "kn-IN", "te-IN"}
}

// --- NoopGoogleVisionClient ---

// NoopGoogleVisionClient is a no-op implementation of GoogleVisionProvider for
// development and testing.
type NoopGoogleVisionClient struct{}

// NewNoopGoogleVisionClient creates a new no-op Google Vision client.
func NewNoopGoogleVisionClient() *NoopGoogleVisionClient {
	return &NoopGoogleVisionClient{}
}

// DetectText returns a mock OCR result.
func (n *NoopGoogleVisionClient) DetectText(_ context.Context, imageData []byte) (*OCRResult, error) {
	log.Debug().Int("image_bytes", len(imageData)).Msg("noop google vision detect text")
	return &OCRResult{
		Text:     "GOVERNMENT OF INDIA\nName: John Doe\nDOB: 01/01/1990\nAadhaar: XXXX XXXX 1234",
		Language: "en",
		Blocks: []TextBlock{
			{Text: "GOVERNMENT OF INDIA", Confidence: 0.98},
			{Text: "Name: John Doe", Confidence: 0.95},
			{Text: "DOB: 01/01/1990", Confidence: 0.97},
			{Text: "Aadhaar: XXXX XXXX 1234", Confidence: 0.94},
		},
	}, nil
}

// DetectLabels returns mock image classification labels.
func (n *NoopGoogleVisionClient) DetectLabels(_ context.Context, imageData []byte) ([]Label, error) {
	log.Debug().Int("image_bytes", len(imageData)).Msg("noop google vision detect labels")
	return []Label{
		{Description: "Document", Score: 0.95},
		{Description: "Identity card", Score: 0.90},
		{Description: "Text", Score: 0.88},
	}, nil
}

// DetectFaces returns a mock face detection result.
func (n *NoopGoogleVisionClient) DetectFaces(_ context.Context, imageData []byte) ([]Face, error) {
	log.Debug().Int("image_bytes", len(imageData)).Msg("noop google vision detect faces")
	return []Face{
		{
			Confidence: 0.99,
			BoundingBox: []Vertex{
				{X: 100, Y: 100},
				{X: 300, Y: 100},
				{X: 300, Y: 350},
				{X: 100, Y: 350},
			},
			Joy:      "VERY_LIKELY",
			Sorrow:   "VERY_UNLIKELY",
			Anger:    "VERY_UNLIKELY",
			Surprise: "UNLIKELY",
		},
	}, nil
}

// CompareFaces returns a mock high similarity score.
func (n *NoopGoogleVisionClient) CompareFaces(_ context.Context, face1, face2 []byte) (float64, error) {
	log.Debug().Int("face1_bytes", len(face1)).Int("face2_bytes", len(face2)).Msg("noop google vision compare faces")
	return 0.92, nil
}

// --- NoopGoogleTranslateClient ---

// NoopGoogleTranslateClient is a no-op implementation of GoogleTranslateProvider
// for development and testing.
type NoopGoogleTranslateClient struct{}

// NewNoopGoogleTranslateClient creates a new no-op Google Translate client.
func NewNoopGoogleTranslateClient() *NoopGoogleTranslateClient {
	return &NoopGoogleTranslateClient{}
}

// Translate returns the original text prefixed with the target language code.
func (n *NoopGoogleTranslateClient) Translate(_ context.Context, text string, sourceLang, targetLang string) (string, error) {
	log.Debug().Str("source", sourceLang).Str("target", targetLang).Msg("noop google translate")
	return "[" + targetLang + "] " + text, nil
}

// DetectLanguage returns English with high confidence.
func (n *NoopGoogleTranslateClient) DetectLanguage(_ context.Context, text string) (string, float64, error) {
	log.Debug().Int("text_len", len(text)).Msg("noop google translate detect language")
	return "en", 0.98, nil
}

// TranslateBatch returns the original texts prefixed with the target language code.
func (n *NoopGoogleTranslateClient) TranslateBatch(_ context.Context, texts []string, sourceLang, targetLang string) ([]string, error) {
	log.Debug().Str("source", sourceLang).Str("target", targetLang).Int("count", len(texts)).Msg("noop google translate batch")
	results := make([]string, len(texts))
	for i, t := range texts {
		results[i] = "[" + targetLang + "] " + t
	}
	return results, nil
}
