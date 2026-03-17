package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/seva-platform/backend/internal/config"
	"github.com/seva-platform/backend/internal/worker"
)

// IVRHandler handles incoming voice calls via Twilio IVR webhooks.
// It provides a DTMF-driven menu for users on basic phones to find
// and book service providers.
type IVRHandler struct {
	cfg         *config.Config
	db          *pgxpool.Pool
	asynqClient *asynq.Client
}

// NewIVRHandler creates a new IVRHandler.
func NewIVRHandler(cfg *config.Config, db *pgxpool.Pool) *IVRHandler {
	// Create an asynq client for enqueuing background tasks (e.g., sending SMS).
	var asynqClient *asynq.Client
	if cfg.RedisURL != "" {
		asynqClient = asynq.NewClient(asynq.RedisClientOpt{Addr: parseRedisAddr(cfg.RedisURL)})
	}
	return &IVRHandler{cfg: cfg, db: db, asynqClient: asynqClient}
}

// parseRedisAddr extracts the host:port from a redis:// URL for asynq.
func parseRedisAddr(redisURL string) string {
	// Strip the scheme.
	addr := redisURL
	if idx := strings.Index(addr, "://"); idx >= 0 {
		addr = addr[idx+3:]
	}
	// Strip any auth portion (user:pass@).
	if idx := strings.LastIndex(addr, "@"); idx >= 0 {
		addr = addr[idx+1:]
	}
	// Strip the path (e.g., /0).
	if idx := strings.Index(addr, "/"); idx >= 0 {
		addr = addr[:idx]
	}
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}

// RegisterRoutes mounts IVR webhook routes on the given Fiber router group.
// Expected mount point: /webhooks/ivr
func (h *IVRHandler) RegisterRoutes(rg fiber.Router) {
	rg.Post("/incoming", h.HandleIncomingCall)
	rg.Post("/gather", h.HandleGatherInput)
	rg.Post("/status", h.HandleCallStatus)
}

// ---------------------------------------------------------------------------
// Service categories for DTMF selection
// ---------------------------------------------------------------------------

var serviceCategories = []struct {
	Digit int
	EN    string
	HI    string
	KN    string
	Slug  string
}{
	{1, "Coconut tree climbing", "Nariyal ped chadhai", "Tenginakaayi mara hattuvudu", "coconut-climbing"},
	{2, "Plumbing", "Plumbing", "Plumbing", "plumbing"},
	{3, "Electrical work", "Bijli ka kaam", "Vidyut kelasa", "electrical"},
	{4, "House cleaning", "Ghar safai", "Mane shuchigolisu", "cleaning"},
	{5, "Gardening", "Bagwaani", "Thota kelasa", "gardening"},
	{6, "Painting", "Painting", "Painting", "painting"},
	{7, "Pest control", "Keede maarna", "Keeta niyantrana", "pest-control"},
	{8, "Carpentry", "Karigari", "Maranigiri", "carpentry"},
	{9, "Other services", "Anya sevaayein", "Itara sevegalu", "other"},
}

// ---------------------------------------------------------------------------
// TwiML builder helpers
// ---------------------------------------------------------------------------

func twimlSay(text, language, voice string) string {
	if voice == "" {
		voice = "Polly.Aditi" // AWS Polly Indian English / Hindi voice
	}
	if language == "" {
		language = "en-IN"
	}
	return fmt.Sprintf(`<Say language="%s" voice="%s">%s</Say>`, language, voice, xmlEscape(text))
}

func twimlGather(action string, numDigits int, timeout int, content string) string {
	return fmt.Sprintf(`<Gather action="%s" numDigits="%d" timeout="%d">%s</Gather>`,
		action, numDigits, timeout, content)
}

// ---------------------------------------------------------------------------
// Handle incoming call
// ---------------------------------------------------------------------------

// HandleIncomingCall handles POST /webhooks/ivr/incoming from Twilio.
// It plays a welcome message and asks the user to select a language.
func (h *IVRHandler) HandleIncomingCall(c *fiber.Ctx) error {
	from := c.FormValue("From")
	callSid := c.FormValue("CallSid")

	log.Info().
		Str("from", from).
		Str("call_sid", callSid).
		Msg("ivr: incoming call")

	// Build TwiML: Welcome + language selection.
	welcomeEN := twimlSay("Welcome to Seva. ", "en-IN", "Polly.Aditi")
	welcomeHI := twimlSay("Seva mein aapka swaagat hai. ", "hi-IN", "Polly.Aditi")

	langPrompt := twimlSay(
		"Press 1 for English. 2 ke liye Hindi. 3 ke liye Kannada.",
		"en-IN", "Polly.Aditi",
	)

	gatherContent := welcomeEN + welcomeHI + langPrompt
	gather := twimlGather("/webhooks/ivr/gather?step=language", 1, 10, gatherContent)

	// If no input, repeat the prompt.
	fallback := twimlSay("We did not receive any input. Goodbye.", "en-IN", "Polly.Aditi")

	twiml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  %s
  <Hangup/>
</Response>`, gather, fallback)

	c.Set("Content-Type", "application/xml")
	return c.Status(fiber.StatusOK).SendString(twiml)
}

// ---------------------------------------------------------------------------
// Handle DTMF input
// ---------------------------------------------------------------------------

// HandleGatherInput handles POST /webhooks/ivr/gather from Twilio.
// The "step" query parameter determines which stage of the IVR flow we are in.
func (h *IVRHandler) HandleGatherInput(c *fiber.Ctx) error {
	digits := c.FormValue("Digits")
	from := c.FormValue("From")
	step := c.Query("step", "language")

	log.Info().
		Str("from", from).
		Str("digits", digits).
		Str("step", step).
		Msg("ivr: gather input received")

	var twiml string

	switch step {
	case "language":
		twiml = h.handleLanguageSelection(digits)
	case "category":
		lang := c.Query("lang", "en")
		twiml = h.handleCategorySelection(digits, lang)
	case "confirm":
		lang := c.Query("lang", "en")
		category := c.Query("cat", "")
		twiml = h.handleConfirmation(digits, lang, category, from)
	default:
		twiml = h.buildErrorTwiML("Invalid step")
	}

	c.Set("Content-Type", "application/xml")
	return c.Status(fiber.StatusOK).SendString(twiml)
}

// handleLanguageSelection processes the language selection digit and presents
// the service category menu.
func (h *IVRHandler) handleLanguageSelection(digits string) string {
	lang := "en"
	ttsLang := "en-IN"

	switch digits {
	case "1":
		lang = "en"
		ttsLang = "en-IN"
	case "2":
		lang = "hi"
		ttsLang = "hi-IN"
	case "3":
		lang = "kn"
		ttsLang = "kn-IN"
	default:
		return h.buildErrorTwiML("Invalid language selection")
	}

	// Build category menu.
	var menuText string
	for _, cat := range serviceCategories {
		var name string
		switch lang {
		case "hi":
			name = cat.HI
		case "kn":
			name = cat.KN
		default:
			name = cat.EN
		}
		menuText += fmt.Sprintf("Press %d for %s. ", cat.Digit, name)
	}

	var prompt string
	switch lang {
	case "hi":
		prompt = "Apni seva chunein. "
	case "kn":
		prompt = "Nimma seva aayke maadi. "
	default:
		prompt = "Please select a service category. "
	}

	sayContent := twimlSay(prompt+menuText, ttsLang, "Polly.Aditi")
	gatherAction := fmt.Sprintf("/webhooks/ivr/gather?step=category&lang=%s", lang)
	gather := twimlGather(gatherAction, 1, 15, sayContent)

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  %s
  <Hangup/>
</Response>`, gather, twimlSay("No input received. Goodbye.", ttsLang, "Polly.Aditi"))
}

// handleCategorySelection processes the category selection and asks for location
// confirmation.
func (h *IVRHandler) handleCategorySelection(digits, lang string) string {
	ttsLang := "en-IN"
	switch lang {
	case "hi":
		ttsLang = "hi-IN"
	case "kn":
		ttsLang = "kn-IN"
	}

	// Find the selected category.
	var selectedSlug, selectedName string
	digitInt := 0
	if len(digits) == 1 && digits[0] >= '1' && digits[0] <= '9' {
		digitInt = int(digits[0] - '0')
	}

	for _, cat := range serviceCategories {
		if cat.Digit == digitInt {
			selectedSlug = cat.Slug
			switch lang {
			case "hi":
				selectedName = cat.HI
			case "kn":
				selectedName = cat.KN
			default:
				selectedName = cat.EN
			}
			break
		}
	}

	if selectedSlug == "" {
		return h.buildErrorTwiML("Invalid category selection")
	}

	// Confirm the selection and ask to proceed.
	var confirmText string
	switch lang {
	case "hi":
		confirmText = fmt.Sprintf("Aapne %s chuna. Aage badhne ke liye 1 dabayein. Vapas jaane ke liye 2 dabayein.", selectedName)
	case "kn":
		confirmText = fmt.Sprintf("Neevu %s aayke maadiddeera. Mundvariyalu 1 otti. Hinde hogalu 2 otti.", selectedName)
	default:
		confirmText = fmt.Sprintf("You selected %s. Press 1 to confirm and find providers. Press 2 to go back.", selectedName)
	}

	sayContent := twimlSay(confirmText, ttsLang, "Polly.Aditi")
	gatherAction := fmt.Sprintf("/webhooks/ivr/gather?step=confirm&lang=%s&cat=%s", lang, selectedSlug)
	gather := twimlGather(gatherAction, 1, 10, sayContent)

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  <Hangup/>
</Response>`, gather)
}

// handleConfirmation processes the confirmation response and either connects
// the user to a provider or takes them back to the category menu.
func (h *IVRHandler) handleConfirmation(digits, lang, category, from string) string {
	ttsLang := "en-IN"
	switch lang {
	case "hi":
		ttsLang = "hi-IN"
	case "kn":
		ttsLang = "kn-IN"
	}

	if digits == "2" {
		// Go back to category selection.
		return h.handleLanguageSelection(map[string]string{"en": "1", "hi": "2", "kn": "3"}[lang])
	}

	if digits != "1" {
		return h.buildErrorTwiML("Invalid input")
	}

	log.Info().
		Str("from", from).
		Str("category", category).
		Str("lang", lang).
		Msg("ivr: provider search confirmed")

	// Search for providers in the selected category.
	type providerResult struct {
		Name       string
		Phone      string
		TrustScore float64
	}

	var providers []providerResult

	if h.db != nil {
		ctx := context.Background()
		rows, err := h.db.Query(ctx,
			`SELECT u.name, u.phone, pp.trust_score
			 FROM provider_profiles pp
			 JOIN users u ON pp.user_id = u.id
			 WHERE pp.is_active = true
			   AND EXISTS (SELECT 1 FROM unnest(pp.skills) s WHERE LOWER(s) LIKE '%' || $1 || '%')
			 ORDER BY pp.trust_score DESC
			 LIMIT 3`,
			strings.ToLower(category),
		)
		if err != nil {
			log.Error().Err(err).Str("category", category).Msg("ivr: failed to query providers")
		} else {
			defer rows.Close()
			for rows.Next() {
				var p providerResult
				if err := rows.Scan(&p.Name, &p.Phone, &p.TrustScore); err != nil {
					log.Error().Err(err).Msg("ivr: failed to scan provider row")
					continue
				}
				providers = append(providers, p)
			}
			if err := rows.Err(); err != nil {
				log.Error().Err(err).Msg("ivr: error iterating provider rows")
			}
		}
	}

	if len(providers) == 0 {
		// No providers found for this category.
		log.Info().
			Str("from", from).
			Str("category", category).
			Msg("ivr: no providers found for category")

		var responseText string
		switch lang {
		case "hi":
			responseText = "Maaf kijiye, aapke area mein is seva ke liye koi provider abhi uplabdh nahi hai. Humne aapka anurodh not kar liya hai. Dhanyavaad!"
		case "kn":
			responseText = "Kshamisibeku, nimma pradeshada ee sevege yavude provider inna illavilla. Namma nimma vinantiya gamnisi iddeve. Dhanyavaadagalu!"
		default:
			responseText = "We don't have providers for that service in your area yet. We've noted your request. Goodbye."
		}

		return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  <Hangup/>
</Response>`, twimlSay(responseText, ttsLang, "Polly.Aditi"))
	}

	// Providers found -- enqueue an SMS task with provider details.
	log.Info().
		Str("from", from).
		Str("category", category).
		Int("provider_count", len(providers)).
		Msg("ivr: providers found, enqueuing SMS with details")

	// Build the SMS message body with top provider details.
	smsBody := fmt.Sprintf("Seva: We found %d provider(s) for %s:\n", len(providers), category)
	for i, p := range providers {
		smsBody += fmt.Sprintf("%d. %s - %s (Rating: %.0f/100)\n", i+1, p.Name, p.Phone, p.TrustScore)
	}
	smsBody += "Call them directly or book via Seva. Reply STOP to opt out."

	// Enqueue the SMS task asynchronously.
	if h.asynqClient != nil && from != "" {
		smsTask, err := worker.NewSendSMSTask(from, smsBody)
		if err != nil {
			log.Error().Err(err).Str("phone", from).Msg("ivr: failed to create SMS task")
		} else {
			if _, err := h.asynqClient.Enqueue(smsTask); err != nil {
				log.Error().Err(err).Str("phone", from).Msg("ivr: failed to enqueue SMS task")
			} else {
				log.Info().Str("phone", from).Msg("ivr: SMS task enqueued with provider details")
			}
		}
	}

	// Respond with TwiML telling the user providers were found.
	var responseText string
	switch lang {
	case "hi":
		responseText = fmt.Sprintf("Humne aapke area mein %d provider paaye hain. Aapko jaldi hi SMS se unki jaankari milegi. Seva ka upyog karne ke liye dhanyavaad!", len(providers))
	case "kn":
		responseText = fmt.Sprintf("Naavu nimma pradeshada %d provider galannu kandu hididdeve. Nimge bega SMS moolaka avara vivaragalu baruttave. Seva balasidakke dhanyavaadagalu!", len(providers))
	default:
		responseText = fmt.Sprintf("We found %d providers near you. We'll send you an SMS with their details shortly. Thank you for using Seva!", len(providers))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  <Hangup/>
</Response>`, twimlSay(responseText, ttsLang, "Polly.Aditi"))
}

// buildErrorTwiML returns a TwiML response for error conditions.
func (h *IVRHandler) buildErrorTwiML(reason string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  %s
  <Hangup/>
</Response>`, twimlSay("Sorry, there was an error. Please try again later. "+reason, "en-IN", "Polly.Aditi"))
}

// ---------------------------------------------------------------------------
// Call status webhook
// ---------------------------------------------------------------------------

// HandleCallStatus handles POST /webhooks/ivr/status from Twilio.
// It logs call status changes for analytics.
func (h *IVRHandler) HandleCallStatus(c *fiber.Ctx) error {
	callSid := c.FormValue("CallSid")
	callStatus := c.FormValue("CallStatus")
	from := c.FormValue("From")
	duration := c.FormValue("CallDuration")

	log.Info().
		Str("call_sid", callSid).
		Str("status", callStatus).
		Str("from", from).
		Str("duration", duration).
		Msg("ivr: call status update")

	return c.Status(fiber.StatusOK).SendString("OK")
}
