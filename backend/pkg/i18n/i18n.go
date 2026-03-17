package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/rs/zerolog/log"
)

// Translator provides SMS template formatting and language utilities for the
// Seva platform.
type Translator struct {
	configDir string
	templates map[string]map[string]string // language -> templateKey -> template
	mu        sync.RWMutex
}

// NewTranslator creates a new Translator that reads SMS templates from the
// given configuration directory.
func NewTranslator(configDir string) *Translator {
	return &Translator{
		configDir: configDir,
		templates: make(map[string]map[string]string),
	}
}

// LoadSMSTemplates loads SMS templates for the specified language from a JSON
// file at {configDir}/sms_templates/{language}.json.
//
// Expected file format:
//
//	{
//	    "otp_verification": "Your Seva verification code is {{code}}. Valid for {{minutes}} minutes.",
//	    "job_notification": "New job request: {{category}} in {{location}}. Open app to respond."
//	}
func (t *Translator) LoadSMSTemplates(language string) error {
	path := filepath.Join(t.configDir, "sms_templates", language+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("i18n load templates for %s: %w", language, err)
	}

	var templates map[string]string
	if err := json.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("i18n parse templates for %s: %w", language, err)
	}

	t.mu.Lock()
	t.templates[language] = templates
	t.mu.Unlock()

	log.Info().
		Str("language", language).
		Int("templates", len(templates)).
		Msg("SMS templates loaded")

	return nil
}

// FormatSMS formats an SMS message using the named template in the specified
// language, substituting the provided variables.
//
// Variables in templates use the {{key}} syntax:
//
//	FormatSMS("otp_verification", "en", map[string]string{"code": "1234", "minutes": "5"})
//	// -> "Your Seva verification code is 1234. Valid for 5 minutes."
func (t *Translator) FormatSMS(templateKey string, language string, vars map[string]string) (string, error) {
	t.mu.RLock()
	langTemplates, ok := t.templates[language]
	t.mu.RUnlock()

	if !ok {
		// Try to load templates on demand.
		if err := t.LoadSMSTemplates(language); err != nil {
			// Fall back to English.
			t.mu.RLock()
			langTemplates, ok = t.templates["en"]
			t.mu.RUnlock()
			if !ok {
				return "", fmt.Errorf("i18n: no templates loaded for language %q or fallback", language)
			}
		} else {
			t.mu.RLock()
			langTemplates = t.templates[language]
			t.mu.RUnlock()
		}
	}

	tmpl, ok := langTemplates[templateKey]
	if !ok {
		return "", fmt.Errorf("i18n: template %q not found for language %q", templateKey, language)
	}

	// Substitute variables.
	result := tmpl
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}

	return result, nil
}

// supportedLanguages lists the language codes supported by the Seva platform.
var supportedLanguages = []string{
	"en",  // English
	"hi",  // Hindi
	"ml",  // Malayalam
	"ta",  // Tamil
	"kn",  // Kannada
	"te",  // Telugu
	"bn",  // Bengali
	"mr",  // Marathi
	"gu",  // Gujarati
	"pa",  // Punjabi
}

// SupportedLanguages returns the list of language codes supported by the
// platform.
func SupportedLanguages() []string {
	langs := make([]string, len(supportedLanguages))
	copy(langs, supportedLanguages)
	return langs
}

// scriptRanges maps Unicode script ranges to language codes for simple script-based
// language detection.
var scriptRanges = []struct {
	RangeStart rune
	RangeEnd   rune
	Language   string
}{
	{0x0900, 0x097F, "hi"}, // Devanagari -> Hindi
	{0x0D00, 0x0D7F, "ml"}, // Malayalam
	{0x0B80, 0x0BFF, "ta"}, // Tamil
	{0x0C80, 0x0CFF, "kn"}, // Kannada
	{0x0C00, 0x0C7F, "te"}, // Telugu
	{0x0980, 0x09FF, "bn"}, // Bengali
	{0x0A80, 0x0AFF, "gu"}, // Gujarati
	{0x0A00, 0x0A7F, "pa"}, // Gurmukhi -> Punjabi
}

// DetectLanguage performs simple script-based language detection. It examines
// the Unicode script of each character and returns the most likely language
// code. Returns "en" if no Indic script is detected.
func DetectLanguage(text string) string {
	if text == "" {
		return "en"
	}

	scriptCounts := make(map[string]int)

	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}

		for _, sr := range scriptRanges {
			if r >= sr.RangeStart && r <= sr.RangeEnd {
				scriptCounts[sr.Language]++
				break
			}
		}
	}

	// Find the script with the most characters.
	maxCount := 0
	detected := "en"
	for lang, count := range scriptCounts {
		if count > maxCount {
			maxCount = count
			detected = lang
		}
	}

	return detected
}

// transliterationMap provides basic transliterations between Devanagari and Latin
// scripts. This is a simplified mapping for common characters.
var devanagariToLatin = map[rune]string{
	'अ': "a", 'आ': "aa", 'इ': "i", 'ई': "ee", 'उ': "u", 'ऊ': "oo",
	'ए': "e", 'ऐ': "ai", 'ओ': "o", 'औ': "au",
	'क': "ka", 'ख': "kha", 'ग': "ga", 'घ': "gha",
	'च': "cha", 'छ': "chha", 'ज': "ja", 'झ': "jha",
	'ट': "ta", 'ठ': "tha", 'ड': "da", 'ढ': "dha",
	'त': "ta", 'थ': "tha", 'द': "da", 'ध': "dha", 'न': "na",
	'प': "pa", 'फ': "pha", 'ब': "ba", 'भ': "bha", 'म': "ma",
	'य': "ya", 'र': "ra", 'ल': "la", 'व': "va",
	'श': "sha", 'ष': "sha", 'स': "sa", 'ह': "ha",
	'ं': "n", 'ः': "h",
	'ा': "a", 'ि': "i", 'ी': "ee", 'ु': "u", 'ू': "oo",
	'े': "e", 'ै': "ai", 'ो': "o", 'ौ': "au",
	'्': "",
}

// Transliterate performs basic transliteration between scripts. Currently
// supports Devanagari to Latin conversion. For production-quality transliteration,
// consider using a proper library or API.
func Transliterate(text string, fromScript, toScript string) string {
	if fromScript == "devanagari" && toScript == "latin" {
		var result strings.Builder
		for _, r := range text {
			if latin, ok := devanagariToLatin[r]; ok {
				result.WriteString(latin)
			} else {
				result.WriteRune(r)
			}
		}
		return result.String()
	}

	// Unsupported script pair; return original text.
	log.Warn().
		Str("from", fromScript).
		Str("to", toScript).
		Msg("unsupported transliteration pair")

	return text
}
