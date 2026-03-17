// Package content provides the content and education service implementation
// with real database operations for managing articles, guides, and educational
// resources for both providers and customers (Vision.md Section 19).
package content

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Article represents an educational content article on the Seva platform.
type Article struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`    // provider_guide, customer_tip, maintenance, pricing, legal
	Audience    string    `json:"audience"`     // provider, customer, both
	Tags        []string  `json:"tags"`
	Language    string    `json:"language"`
	AuthorName  string    `json:"author_name"`
	IsPublished bool      `json:"is_published"`
	ViewCount   int       `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContentService implements content management with direct database operations.
type ContentService struct {
	db *pgxpool.Pool
}

// NewContentService returns a ready-to-use ContentService. It ensures the
// articles table exists and seeds initial content.
func NewContentService(db *pgxpool.Pool) *ContentService {
	svc := &ContentService{db: db}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := svc.ensureTableExists(ctx); err != nil {
		log.Error().Err(err).Msg("content: failed to ensure articles table exists")
	}

	if err := svc.seedArticles(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to seed initial articles")
	}

	log.Info().Msg("content: service initialized")
	return svc
}

// ensureTableExists creates the articles table and indexes if they do not exist.
func (s *ContentService) ensureTableExists(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS articles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		slug TEXT UNIQUE NOT NULL,
		title TEXT NOT NULL,
		summary TEXT NOT NULL DEFAULT '',
		body TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT 'general',
		audience TEXT NOT NULL DEFAULT 'both',
		tags TEXT[] DEFAULT '{}',
		language TEXT NOT NULL DEFAULT 'en',
		author_name TEXT NOT NULL DEFAULT 'Seva Team',
		is_published BOOLEAN NOT NULL DEFAULT false,
		view_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
	CREATE INDEX IF NOT EXISTS idx_articles_audience ON articles(audience) WHERE is_published = true;
	CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category) WHERE is_published = true;
	`
	_, err := s.db.Exec(ctx, ddl)
	return err
}

// slugify converts a title into a URL-safe slug.
func slugify(title string) string {
	s := strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// seedArticles inserts initial educational content using INSERT ON CONFLICT DO NOTHING.
func (s *ContentService) seedArticles(ctx context.Context) error {
	type seed struct {
		Slug     string
		Title    string
		Summary  string
		Body     string
		Category string
		Audience string
		Tags     []string
	}

	seeds := []seed{
		{
			Slug:     "how-to-grow-your-service-business-on-seva",
			Title:    "How to grow your service business on Seva",
			Summary:  "Practical strategies to attract more customers, increase your earnings, and build a sustainable service business on the Seva platform.",
			Body:     "## Build a Strong Profile\n\nYour profile is the first thing customers see. Add a professional photo, write a clear bio highlighting your experience, and list all your skills and certifications.\n\n## Respond Quickly\n\nProviders who respond within 15 minutes receive 3x more bookings. Enable push notifications and set up your schedule accurately so you only receive jobs when you are available.\n\n## Ask for Reviews\n\nAfter completing a job, politely ask satisfied customers to leave a review. Higher ratings lead to better visibility in search results and the matching algorithm.\n\n## Set Competitive Pricing\n\nResearch what other providers in your area charge for similar services. Price competitively, but do not undervalue your expertise. You can adjust rates as your reputation grows.\n\n## Use the Analytics Dashboard\n\nVisit your Provider Dashboard regularly to track earnings, response rates, and demand patterns. Use these insights to optimize your availability and service areas.\n\n## Invest in Skills\n\nKeep learning. Seva rewards providers who complete skill certifications with badges that boost trust scores.",
			Category: "provider_guide",
			Audience: "provider",
			Tags:     []string{"business", "growth", "tips", "earnings"},
		},
		{
			Slug:     "5-tips-for-getting-more-5-star-reviews",
			Title:    "5 tips for getting more 5-star reviews",
			Summary:  "Learn how top-rated providers consistently earn excellent reviews and build a stellar reputation on Seva.",
			Body:     "## 1. Communicate Clearly\n\nBefore starting a job, discuss the scope of work, timeline, and pricing with the customer. Clear expectations lead to happy outcomes.\n\n## 2. Arrive on Time\n\nPunctuality is one of the most valued traits. If you are running late, notify the customer immediately through the Seva messaging system.\n\n## 3. Do Quality Work\n\nTake pride in your craft. Clean up after yourself, test your work, and ensure everything is functioning properly before leaving.\n\n## 4. Be Professional\n\nDress appropriately, be courteous, and treat the customer's home with respect. Small gestures like wearing shoe covers make a big difference.\n\n## 5. Follow Up\n\nAfter the job, send a brief message thanking the customer and asking if everything is working well. This shows you care and often prompts a positive review.",
			Category: "provider_guide",
			Audience: "provider",
			Tags:     []string{"reviews", "reputation", "tips", "quality"},
		},
		{
			Slug:     "understanding-your-earnings-dashboard",
			Title:    "Understanding your earnings dashboard",
			Summary:  "A detailed walkthrough of the Provider Earnings Dashboard, including how to read charts, track payouts, and understand fee structures.",
			Body:     "## Overview\n\nYour earnings dashboard gives you a real-time view of your income on Seva. Here is how to use it effectively.\n\n## Earnings History\n\nThe earnings chart shows daily, weekly, or monthly income. Toggle between views to spot trends and plan your schedule around high-demand periods.\n\n## Payout Schedule\n\nSeva processes payouts on a weekly basis. Funds from completed jobs are held in escrow until the customer confirms satisfaction, then released to your linked bank account.\n\n## Fee Structure\n\nSeva charges a small platform fee on each completed transaction. This fee covers payment processing, customer support, and platform maintenance. The exact percentage depends on your subscription tier.\n\n## Tax Information\n\nYour dashboard includes a downloadable earnings report that you can use for tax filing. See our separate guide on tax filing for freelance service providers.",
			Category: "provider_guide",
			Audience: "provider",
			Tags:     []string{"earnings", "dashboard", "payments", "finance"},
		},
		{
			Slug:     "tax-filing-guide-for-freelance-service-providers",
			Title:    "Tax filing guide for freelance service providers",
			Summary:  "Everything you need to know about filing taxes as a freelance service provider in India, including GST, TDS, and quarterly advance tax.",
			Body:     "## Why Tax Filing Matters\n\nAs a freelance service provider, you are responsible for filing your own taxes. Proper tax compliance protects you from penalties and builds financial credibility.\n\n## Income Tax Basics\n\nAll income earned through Seva is taxable. You can download your annual earnings report from the Provider Dashboard under Settings > Tax Reports.\n\n## GST Registration\n\nIf your annual turnover exceeds Rs. 20 lakh (Rs. 10 lakh for special category states), you must register for GST. Even below this threshold, voluntary registration can help you claim input tax credits.\n\n## TDS (Tax Deducted at Source)\n\nSeva may deduct TDS on payments above certain thresholds as required by law. You will receive a TDS certificate that you can use to claim credit when filing returns.\n\n## Quarterly Advance Tax\n\nFreelancers with tax liability exceeding Rs. 10,000 in a year must pay advance tax in quarterly installments (June 15, September 15, December 15, March 15).\n\n## Deductible Expenses\n\nYou can deduct business expenses such as tools, travel, vehicle maintenance, phone bills, and professional development courses. Keep receipts and maintain proper records.\n\n## Recommended Steps\n\n1. Register on the Income Tax e-filing portal\n2. Maintain a separate bank account for business income\n3. Track all business expenses monthly\n4. File GST returns (if applicable) by the 20th of each month\n5. Consult a chartered accountant for personalized advice",
			Category: "provider_guide",
			Audience: "provider",
			Tags:     []string{"tax", "finance", "gst", "legal", "compliance"},
		},
		{
			Slug:     "how-to-prepare-your-home-for-a-plumber-visit",
			Title:    "How to prepare your home for a plumber visit",
			Summary:  "Simple steps to prepare before a plumber arrives so the job goes smoothly, saving you time and money.",
			Body:     "## Clear the Work Area\n\nRemove personal items, cleaning supplies, and anything stored under sinks or around the work area. This gives the plumber clear access and prevents damage to your belongings.\n\n## Know Your Water Shutoff\n\nLocate your main water shutoff valve and know how to turn it off. In an emergency, this knowledge is invaluable. Show the plumber where it is when they arrive.\n\n## Document the Problem\n\nTake photos or videos of the issue before the plumber arrives. Note when the problem started and whether it is getting worse. This helps the plumber diagnose the issue faster.\n\n## Secure Pets\n\nKeep pets in a separate room during the visit. This is safer for everyone and prevents the plumber from being distracted.\n\n## Ask Questions\n\nDo not hesitate to ask the plumber to explain what they are doing and why. Understanding the repair helps you maintain your plumbing better.\n\n## Get a Written Estimate\n\nBefore work begins, ask for a written estimate through the Seva platform. This protects both you and the provider.",
			Category: "customer_tip",
			Audience: "customer",
			Tags:     []string{"plumbing", "preparation", "home", "tips"},
		},
		{
			Slug:     "what-should-an-electrician-charge-a-pricing-guide",
			Title:    "What should an electrician charge? A pricing guide",
			Summary:  "Understand typical electrician rates in India, what factors affect pricing, and how to ensure you are getting a fair deal.",
			Body:     "## Typical Rates\n\nElectrician rates in India vary by city and experience. Here are general ranges:\n\n- **Minor repairs** (switch replacement, fan installation): Rs. 200-500\n- **Medium jobs** (wiring repair, MCB installation): Rs. 500-1500\n- **Major work** (full house wiring, panel upgrade): Rs. 2000-10000+\n\n## Factors That Affect Pricing\n\n### Experience & Certification\nCertified electricians with more experience typically charge higher rates but deliver higher quality work with fewer callbacks.\n\n### Location\nMetro city rates are 20-40% higher than tier-2 cities. Rates also vary within a city based on the area.\n\n### Urgency\nEmergency or after-hours calls may carry a premium of 25-50% over standard rates.\n\n### Materials\nSome electricians include materials in their quote while others charge separately. Always clarify upfront.\n\n## How to Compare Quotes\n\nOn Seva, you can request multiple quotes for the same job. Compare not just the price but also the provider's rating, reviews, and trust score. The cheapest option is not always the best value.\n\n## Red Flags\n\n- Refusing to provide a written estimate\n- Demanding full payment upfront\n- No reviews or very new account\n- Significantly below-market rates\n\n## Seva's Protection\n\nAll payments on Seva go through escrow, meaning your money is held safely until you confirm the job is completed satisfactorily.",
			Category: "customer_tip",
			Audience: "customer",
			Tags:     []string{"electrician", "pricing", "cost", "guide"},
		},
		{
			Slug:     "seasonal-home-maintenance-checklist",
			Title:    "Seasonal home maintenance checklist",
			Summary:  "A comprehensive checklist of home maintenance tasks organized by season to keep your home in top condition year-round.",
			Body:     "## Pre-Monsoon (May-June)\n\n- [ ] Clean and inspect gutters and downspouts\n- [ ] Check roof for loose tiles or damaged waterproofing\n- [ ] Seal windows and doors against water intrusion\n- [ ] Service air conditioning units\n- [ ] Clear drains and check for blockages\n- [ ] Inspect and repair exterior paint or wall coatings\n- [ ] Trim trees near the house to prevent storm damage\n\n## Monsoon (July-September)\n\n- [ ] Monitor for leaks and water seepage\n- [ ] Check electrical systems for moisture damage\n- [ ] Ensure sump pumps are working (if applicable)\n- [ ] Keep an eye on walls for dampness or mold\n- [ ] Clean water tanks and filters regularly\n\n## Post-Monsoon (October-November)\n\n- [ ] Inspect roof and walls for monsoon damage\n- [ ] Repaint or touch up exterior walls if needed\n- [ ] Deep clean the entire house\n- [ ] Check for pest infestations (termites are common post-monsoon)\n- [ ] Service water heaters before winter\n\n## Winter (December-February)\n\n- [ ] Check water heater and geyser operation\n- [ ] Inspect gas connections and heaters\n- [ ] Clean chimney and exhaust fans\n- [ ] Check for pipe insulation in cold regions\n\n## Summer (March-April)\n\n- [ ] Service air conditioners and coolers\n- [ ] Check electrical load capacity\n- [ ] Paint interior rooms if needed\n- [ ] Inspect plumbing for leaks\n- [ ] Clean water storage tanks\n\n## Pro Tip\n\nBook recurring maintenance services on Seva to never miss a seasonal task. Our providers can set up quarterly visits to handle routine maintenance automatically.",
			Category: "maintenance",
			Audience: "customer",
			Tags:     []string{"maintenance", "seasonal", "checklist", "home"},
		},
		{
			Slug:     "how-to-choose-the-right-service-provider",
			Title:    "How to choose the right service provider",
			Summary:  "A step-by-step guide to finding and selecting the best service provider for your needs on Seva.",
			Body:     "## Define Your Needs\n\nBefore searching, clearly define what you need done. The more specific your job description, the more accurate the quotes you will receive.\n\n## Check Ratings and Reviews\n\nLook for providers with consistently high ratings (4.0+) and read recent reviews. Pay attention to reviews for jobs similar to yours.\n\n## Verify Credentials\n\nLook for the verified badge on provider profiles. Verified providers have completed Seva's KYC process, which includes identity verification and background checks.\n\n## Compare Multiple Quotes\n\nPost your job and wait for at least 3 quotes before deciding. Compare not just price, but also the provider's proposed approach and timeline.\n\n## Check the Trust Score\n\nSeva's Trust Score is a comprehensive metric that considers rating, completion rate, response time, and more. A score above 4.0 indicates a highly reliable provider.\n\n## Look at Response Time\n\nProviders who respond quickly tend to be more professional and engaged. The platform shows each provider's average response time.\n\n## Start with a Small Job\n\nIf you need ongoing services, start with a small job to evaluate the provider before committing to larger projects.\n\n## Use Escrow Protection\n\nAlways use Seva's built-in payment system. Escrow protection ensures your money is safe until the job meets your expectations.",
			Category: "customer_tip",
			Audience: "customer",
			Tags:     []string{"choosing", "hiring", "tips", "guide"},
		},
		{
			Slug:     "understanding-sevas-escrow-payment-protection",
			Title:    "Understanding Seva's escrow payment protection",
			Summary:  "How Seva's escrow system works to protect both customers and providers during every transaction.",
			Body:     "## What Is Escrow?\n\nEscrow is a financial arrangement where a third party (Seva) holds payment on behalf of the customer until the agreed-upon service is completed satisfactorily.\n\n## How It Works\n\n1. **Customer pays**: When you accept a quote, the payment amount is transferred to Seva's escrow account.\n2. **Provider works**: The provider completes the job knowing the funds are secured.\n3. **Customer confirms**: After the job is done, you confirm satisfaction through the app.\n4. **Provider receives payment**: Once confirmed, the funds are released to the provider's bank account.\n\n## Protection for Customers\n\n- Your money is held safely until you are satisfied\n- If the provider does not show up, you get a full refund\n- If the work is unsatisfactory, you can open a dispute\n\n## Protection for Providers\n\n- You know the customer has the funds available\n- No risk of non-payment after completing work\n- Disputes are mediated fairly by Seva's team\n\n## Dispute Process\n\nIf either party is unhappy, they can open a dispute within 48 hours. Seva's mediation team reviews evidence from both sides and makes a fair resolution.\n\n## Payout Timeline\n\n- Standard: Funds released within 24 hours of customer confirmation\n- Disputes: Resolution typically within 3-5 business days",
			Category: "customer_tip",
			Audience: "customer",
			Tags:     []string{"escrow", "payment", "protection", "security"},
		},
		{
			Slug:     "how-sevas-trust-score-works",
			Title:    "How Seva's trust score works",
			Summary:  "A transparent explanation of how Seva calculates trust scores and what providers can do to improve theirs.",
			Body:     "## What Is the Trust Score?\n\nThe trust score is a comprehensive metric (0-5.0) that reflects a provider's reliability, quality, and professionalism on the Seva platform.\n\n## Score Components\n\n### Average Rating (40%)\nYour customer ratings form the largest part of your trust score. Consistently delivering quality work is the best way to improve this.\n\n### Completion Rate (25%)\nThe percentage of accepted jobs that you successfully complete. Cancellations and no-shows significantly impact this metric.\n\n### Response Time (15%)\nHow quickly you respond to job requests and messages. Providers who respond within 15 minutes score highest in this category.\n\n### Volume Bonus (10%)\nMore completed jobs contribute positively to your score, rewarding experienced providers.\n\n### Recency (10%)\nRecent reviews carry more weight than older ones. This ensures the score reflects your current performance.\n\n## Bonus Modifiers\n\n- **Verified Provider**: +0.10 for completing KYC verification\n- **Complete Profile**: +0.05 for having a detailed bio and profile photo\n- **Bank Account Linked**: +0.05 for having a verified bank account\n\n## Provider Levels\n\nYour trust score determines your provider level:\n\n| Level | Trust Score | Jobs Required |\n|-------|------------|---------------|\n| New | < 2.0 | < 3 |\n| Active | >= 2.0 | >= 3 |\n| Trusted | >= 3.5 | >= 20 |\n| Expert | >= 4.0 | >= 50 |\n| Local Champion | >= 4.5 | >= 100 |\n\nHigher levels unlock benefits like priority matching, featured listings, and reduced platform fees.",
			Category: "provider_guide",
			Audience: "both",
			Tags:     []string{"trust-score", "reputation", "levels", "quality"},
		},
		{
			Slug:     "getting-started-with-recurring-service-plans",
			Title:    "Getting started with recurring service plans",
			Summary:  "How to set up and manage recurring service plans for regular maintenance and repeat services on Seva.",
			Body:     "## What Are Recurring Plans?\n\nRecurring service plans let you schedule regular services (cleaning, maintenance, gardening, etc.) with your preferred provider. Set it once and the platform handles scheduling and payments automatically.\n\n## How to Set Up\n\n1. Find a provider you trust or complete your first job with them\n2. Go to your job history and select 'Set up recurring'\n3. Choose your frequency: daily, weekly, biweekly, monthly, or quarterly\n4. Set your preferred day and time\n5. Confirm the recurring rate with the provider\n\n## Benefits for Customers\n\n- **Convenience**: No need to re-book every time\n- **Priority scheduling**: Recurring customers get priority slots\n- **Consistent quality**: Build a relationship with a provider who knows your home\n- **Potential savings**: Many providers offer discounted rates for recurring clients\n\n## Benefits for Providers\n\n- **Steady income**: Predictable revenue stream\n- **Efficient scheduling**: Plan your routes and days in advance\n- **Customer retention**: Build long-term client relationships\n- **Reduced marketing**: Spend less time finding new customers\n\n## Managing Your Plans\n\nYou can pause, resume, or cancel recurring plans anytime from the Recurring section in your dashboard. Providers are notified automatically of any changes.\n\n## Payment\n\nPayments are processed automatically before each scheduled service. You will receive a notification and can review the charge before it goes through.",
			Category: "provider_guide",
			Audience: "both",
			Tags:     []string{"recurring", "plans", "scheduling", "maintenance"},
		},
		{
			Slug:     "safety-tips-when-booking-home-services",
			Title:    "Safety tips when booking home services",
			Summary:  "Essential safety guidelines for both customers and providers when booking and performing home services through Seva.",
			Body:     "## For Customers\n\n### Before the Visit\n- **Share your booking** with a family member or friend\n- **Use Seva's live tracking** feature when available\n- **Add emergency contacts** in the Safety section of your profile\n- **Verify the provider** by checking their profile, reviews, and verification status\n\n### During the Visit\n- **Keep the door open** or have someone else at home\n- **Do not share personal information** beyond what is needed for the job\n- **Monitor the work** and ask questions if anything seems off\n- **Use the in-app SOS button** if you feel unsafe at any time\n\n### After the Visit\n- **Check the work** thoroughly before confirming completion\n- **Leave an honest review** to help other customers\n- **Report any concerns** immediately through the app\n\n## For Providers\n\n### Before the Visit\n- **Verify the job details** and customer address\n- **Share your schedule** with a trusted person\n- **Use GPS navigation** through the app for route tracking\n\n### During the Visit\n- **Be professional** and respectful at all times\n- **Communicate clearly** about what you will be doing\n- **Document the work** with photos (before and after)\n\n### After the Visit\n- **Clean up** your work area\n- **Report any safety concerns** about the location\n\n## Seva's Safety Features\n\n- **SOS Button**: One-tap emergency alert to Seva's safety team and your emergency contacts\n- **Live Location Sharing**: Real-time tracking during active jobs\n- **OTP Verification**: Verify provider identity at the door\n- **Background Checks**: All providers undergo identity and background verification\n- **In-App Communication**: All messages are logged for safety",
			Category: "customer_tip",
			Audience: "both",
			Tags:     []string{"safety", "security", "home", "tips"},
		},
	}

	for _, sd := range seeds {
		_, err := s.db.Exec(ctx, `
			INSERT INTO articles (slug, title, summary, body, category, audience, tags, language, author_name, is_published)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'en', 'Seva Team', true)
			ON CONFLICT (slug) DO NOTHING
		`, sd.Slug, sd.Title, sd.Summary, sd.Body, sd.Category, sd.Audience, sd.Tags)
		if err != nil {
			log.Warn().Err(err).Str("slug", sd.Slug).Msg("content: failed to seed article")
		}
	}

	log.Info().Int("count", len(seeds)).Msg("content: seed articles processed")
	return nil
}

// ListArticles returns published articles filtered by audience, category, and
// language. Returns the articles and total count for pagination.
func (s *ContentService) ListArticles(ctx context.Context, audience, category, lang string, limit, offset int) ([]Article, int, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build dynamic WHERE clause.
	conditions := []string{"is_published = true"}
	args := []interface{}{}
	argIdx := 1

	if audience != "" && audience != "all" {
		conditions = append(conditions, fmt.Sprintf("(audience = $%d OR audience = 'both')", argIdx))
		args = append(args, audience)
		argIdx++
	}

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}

	if lang != "" {
		conditions = append(conditions, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, lang)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count total.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM articles WHERE %s", where)
	var total int
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count articles: %w", err)
	}

	// Fetch page.
	query := fmt.Sprintf(`
		SELECT id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
		FROM articles
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(
			&a.ID, &a.Slug, &a.Title, &a.Summary, &a.Body,
			&a.Category, &a.Audience, &a.Tags, &a.Language,
			&a.AuthorName, &a.IsPublished, &a.ViewCount,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, a)
	}

	if articles == nil {
		articles = []Article{}
	}

	return articles, total, nil
}

// GetArticle returns a published article by slug and increments its view count.
func (s *ContentService) GetArticle(ctx context.Context, slug string) (*Article, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	var a Article
	err := s.db.QueryRow(ctx, `
		UPDATE articles SET view_count = view_count + 1, updated_at = NOW()
		WHERE slug = $1 AND is_published = true
		RETURNING id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
	`, slug).Scan(
		&a.ID, &a.Slug, &a.Title, &a.Summary, &a.Body,
		&a.Category, &a.Audience, &a.Tags, &a.Language,
		&a.AuthorName, &a.IsPublished, &a.ViewCount,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get article by slug: %w", err)
	}

	return &a, nil
}

// GetArticleByID returns an article by UUID (includes unpublished for admin use).
func (s *ContentService) GetArticleByID(ctx context.Context, id uuid.UUID) (*Article, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	var a Article
	err := s.db.QueryRow(ctx, `
		SELECT id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
		FROM articles WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Slug, &a.Title, &a.Summary, &a.Body,
		&a.Category, &a.Audience, &a.Tags, &a.Language,
		&a.AuthorName, &a.IsPublished, &a.ViewCount,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get article by id: %w", err)
	}

	return &a, nil
}

// CreateArticle inserts a new article. Slug is auto-generated from the title if
// not provided.
func (s *ContentService) CreateArticle(ctx context.Context, article *Article) error {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	if article.ID == uuid.Nil {
		article.ID = uuid.New()
	}
	if article.Slug == "" {
		article.Slug = slugify(article.Title)
	}
	if article.Language == "" {
		article.Language = "en"
	}
	if article.AuthorName == "" {
		article.AuthorName = "Seva Team"
	}
	if article.Tags == nil {
		article.Tags = []string{}
	}

	now := time.Now().UTC()
	article.CreatedAt = now
	article.UpdatedAt = now

	_, err := s.db.Exec(ctx, `
		INSERT INTO articles (id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 0, $12, $13)
	`,
		article.ID, article.Slug, article.Title, article.Summary, article.Body,
		article.Category, article.Audience, article.Tags, article.Language,
		article.AuthorName, article.IsPublished,
		article.CreatedAt, article.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create article: %w", err)
	}

	log.Info().Str("slug", article.Slug).Str("id", article.ID.String()).Msg("content: article created")
	return nil
}

// UpdateArticle updates an existing article by ID with the given fields.
func (s *ContentService) UpdateArticle(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	allowedFields := map[string]bool{
		"title": true, "summary": true, "body": true, "category": true,
		"audience": true, "tags": true, "language": true, "author_name": true,
		"is_published": true, "slug": true,
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	for field, value := range updates {
		if !allowedFields[field] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, value)
		argIdx++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	// Always update the timestamp.
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now().UTC())
	argIdx++

	args = append(args, id)
	query := fmt.Sprintf("UPDATE articles SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)

	result, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	log.Info().Str("id", id.String()).Msg("content: article updated")
	return nil
}

// DeleteArticle removes an article by ID.
func (s *ContentService) DeleteArticle(ctx context.Context, id uuid.UUID) error {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	result, err := s.db.Exec(ctx, "DELETE FROM articles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("article not found")
	}

	log.Info().Str("id", id.String()).Msg("content: article deleted")
	return nil
}

// GetPopular returns the most-viewed published articles for a given audience.
func (s *ContentService) GetPopular(ctx context.Context, audience string, limit int) ([]Article, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	if limit <= 0 {
		limit = 10
	}

	var rows pgx.Rows
	var err error

	if audience != "" && audience != "all" {
		rows, err = s.db.Query(ctx, `
			SELECT id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
			FROM articles
			WHERE is_published = true AND (audience = $1 OR audience = 'both')
			ORDER BY view_count DESC
			LIMIT $2
		`, audience, limit)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
			FROM articles
			WHERE is_published = true
			ORDER BY view_count DESC
			LIMIT $1
		`, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("get popular articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(
			&a.ID, &a.Slug, &a.Title, &a.Summary, &a.Body,
			&a.Category, &a.Audience, &a.Tags, &a.Language,
			&a.AuthorName, &a.IsPublished, &a.ViewCount,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, a)
	}

	if articles == nil {
		articles = []Article{}
	}

	return articles, nil
}

// GetRelated returns articles that share the same category or overlapping tags
// with the given article, excluding the article itself.
func (s *ContentService) GetRelated(ctx context.Context, articleID uuid.UUID, limit int) ([]Article, error) {
	if err := s.ensureTableExists(ctx); err != nil {
		log.Warn().Err(err).Msg("content: failed to ensure table exists")
	}

	if limit <= 0 {
		limit = 5
	}

	// First fetch the source article to know its category and tags.
	source, err := s.GetArticleByID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("get source article: %w", err)
	}
	if source == nil {
		return []Article{}, nil
	}

	// Find related by matching category or overlapping tags, ordered by relevance.
	rows, err := s.db.Query(ctx, `
		SELECT id, slug, title, summary, body, category, audience, tags, language, author_name, is_published, view_count, created_at, updated_at
		FROM articles
		WHERE is_published = true
		  AND id != $1
		  AND (category = $2 OR tags && $3)
		ORDER BY
			CASE WHEN category = $2 THEN 0 ELSE 1 END,
			array_length(ARRAY(SELECT unnest(tags) INTERSECT SELECT unnest($3::text[])), 1) DESC NULLS LAST,
			view_count DESC
		LIMIT $4
	`, articleID, source.Category, source.Tags, limit)
	if err != nil {
		return nil, fmt.Errorf("get related articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(
			&a.ID, &a.Slug, &a.Title, &a.Summary, &a.Body,
			&a.Category, &a.Audience, &a.Tags, &a.Language,
			&a.AuthorName, &a.IsPublished, &a.ViewCount,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, a)
	}

	if articles == nil {
		articles = []Article{}
	}

	return articles, nil
}
