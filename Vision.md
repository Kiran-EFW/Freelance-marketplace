# Seva — Service Marketplace Platform

## 1. Core Concept

A platform that aggregates and matches service providers — electricians, plumbers, painters, lawyers, chartered accountants, and other utility/service professionals — with customers who need them.

**What makes this different:**

- **Inclusive by design** — serves providers who may not own a smartphone, operating via SMS and IVR alongside a full mobile/web app
- **Postcode-based matching** — services are local; the platform matches based on proximity, skill, availability, and trust
- **Both companies and freelancers** — a single platform that handles registered businesses and informal solo providers
- **Multi-jurisdiction** — architected from day one to expand across regions, languages, currencies, and regulatory environments

---

## 2. Provider Types

The platform serves two fundamentally different provider profiles:

| Dimension | Informal Freelancer | Registered Company |
|---|---|---|
| Discovery | Postcode + skill match | Brand search + category |
| Trust signal | Peer reviews, ID verification | License, GST/VAT, portfolio |
| Pricing | Negotiable, visit-based | Fixed quotes, packages |
| Communication | SMS / phone call | App / chat |
| Onboarding | Field agent or self via SMS | Self-serve via app/web |
| Interface | SMS/IVR or basic app | Full app with dashboard |

### Provider onboarding channels

- **Smartphone app** — self-serve registration
- **Agent-assisted** — field agents register providers in person (modeled on JioMart's kirana onboarding)
- **SMS/USSD flow** — basic registration via text messages
- **Community referral** — customers onboard providers they already know (highest incentive reward)

---

## 3. Incentive and Gamification System

No blockchain. A points-based system tracked in a standard database with audit logging.

### Earning points

**Providers earn by:**
- Completing jobs — 10 points
- Getting 5-star reviews — 5 points
- Referring another provider — 20 points
- Responding within 10 minutes — 3 points
- Early adopter bonus (first 50 jobs) — 2x multiplier

**Customers earn by:**
- Posting a job — 2 points
- Leaving a review — 3 points
- Referring a friend — 15 points
- Onboarding a local provider — 25 points (highest reward, solves cold start)

### Spending points

**Customers spend on:**
- Priority matching — 10 points
- Verified-only provider filter — 5 points
- Discount on service fee — varies

**Providers spend on:**
- Boosted listing in their postcode — 15 points
- "Trusted" badge — 50 points
- Withdraw as real money — at defined exchange rate

### Provider level system

| Level | Name | Requirement | Perk |
|---|---|---|---|
| 1 | New | Signed up | Listed in results |
| 2 | Active | 5 jobs completed | Priority in search |
| 3 | Trusted | 20 jobs + 4.5 avg rating | Badge + lower commission |
| 4 | Expert | 100 jobs + 4.7 avg rating | Featured listing + mentorship |
| 5 | Local Champion | Top provider in postcode | Exclusive leads + bonus points |

### Customer engagement

- **Streaks** — "Used the platform 3 months in a row" (small point bonus)
- **Neighborhood leaderboard** — "You onboarded 4 providers in your area"
- **Review milestones** — "Your reviews helped 50 people find a provider"

### Design principles

- Only reward, never punish — a provider who takes fewer jobs shouldn't feel penalized
- Keep it simple — a provider checking SMS should never need to understand "XP systems"
- Gamification for basic phone users is communicated as a weekly SMS summary of points and level

---

## 4. Multi-Jurisdiction Architecture

### Per-jurisdiction variables

**Legal:**
- Business registration requirements for providers
- Consumer protection laws (refund rules, liability)
- Data residency (where user data must be stored)
- Tax collection obligations (GST, VAT, sales tax)
- Professional licensing requirements

**Operational:**
- Currency
- Payment methods people actually use
- Address/postcode format
- Phone number format
- SMS gateway provider

**Cultural:**
- Language(s)
- Service categories that exist locally
- Pricing norms (hourly vs. fixed vs. negotiated)
- Trust signals that matter locally

### Architecture approach — jurisdiction as config, not a code fork

```
┌──────────────────────────────────┐
│         Core Platform            │
│  (matching, messaging, ratings,  │
│   user management, job flow)     │
├──────────────────────────────────┤
│      Jurisdiction Config         │
│  ┌────────┐ ┌────────┐ ┌──────┐ │
│  │ India  │ │  UK    │ │ UAE  │ │
│  │--------│ │--------│ │------│ │
│  │ INR    │ │ GBP    │ │ AED  │ │
│  │ UPI    │ │ Stripe │ │ Tap  │ │
│  │ Aadhaar│ │ Gov.UK │ │ EID  │ │
│  │ hi,en  │ │ en     │ │ ar,en│ │
│  │ GST    │ │ VAT    │ │ VAT  │ │
│  └────────┘ └────────┘ └──────┘ │
└──────────────────────────────────┘
```

Each jurisdiction is a config bundle. Adding a new country = writing adapters + config. Core platform code doesn't change.

### Three separation layers

1. **CORE** (jurisdiction-agnostic) — user accounts, matching algorithm, job lifecycle, ratings, messaging, gamification, notifications
2. **ADAPTERS** (pluggable per jurisdiction) — payment gateway, SMS/IVR provider, identity verification, tax calculation, address/postcode resolution
3. **CONFIG** (data per jurisdiction) — supported languages, currency + formatting, service categories, legal copy, provider onboarding requirements

**Validation test:** If launching in country #2 requires changing core code, the architecture failed.

---

## 5. Language and Internationalization

Built in from day one — retrofitting i18n is painful.

| Concern | Approach |
|---|---|
| UI strings | Standard i18n library (i18next or equivalent) with key-based translations. Never hardcode text |
| Provider-generated content | Store in original language, offer machine translation on read |
| SMS templates | Per-language templates stored in jurisdiction config |
| Search | Transliteration support — user types "plumber" in Hindi script or Roman, both work |
| Right-to-left | CSS/layout supports RTL from day one (Arabic, Urdu) for Middle East expansion |
| Service category names | Translated and localized ("CA" in India = "accountant" elsewhere) |

### Rollout

- **Day one:** English + one local language (Hindi if India-first)
- **Scaling:** Add languages per jurisdiction as expansion happens
- **SMS:** Templates are short, easy to maintain in 10+ languages

---

## 6. Payment Gateways

### Unified payment adapter interface

```
Payment Adapter
  ├── chargeCustomer()
  ├── holdInEscrow()
  ├── releasePayout()
  ├── refund()
  └── getStatus()

Implementations:
  ├── India:   Razorpay (UPI, cards, wallets)
  ├── UK:      Stripe (cards, bank transfer)
  ├── UAE:     Tap Payments / PayTabs
  ├── Africa:  M-Pesa / Flutterwave
  ├── SEA:     GrabPay / GCash
  └── Fallback: Cash (mark as "pay on completion")
```

### Cash is a first-class payment method

In many markets, cash dominates for blue-collar services. The platform handles:

- **Cash jobs** — platform tracks the job, customer confirms completion, ratings still happen, no digital payment. Revenue comes from lead fees or subscriptions
- **Digital payment** — full escrow flow
- **Hybrid** — small platform fee charged digitally, service paid in cash

### Payout splits

```
Customer pays ₹1,000
  ├── Provider gets ₹900 (via bank transfer / UPI)
  ├── Platform keeps ₹80 (commission)
  └── Tax collected ₹20 (remitted to government)
```

Each jurisdiction has different rules on tax collection obligations, commission ceilings, and payout timelines.

### Jurisdiction-specific pricing config

```
jurisdiction_config:
  india:
    currency: INR
    commission_rate: {low: 0.03, mid: 0.05, high: 0.08}
    lead_fee_range: {min: 15, max: 500}
    subscription_pro: 299
    subscription_biz: 999
    urgent_fee: 49
    cash_jobs_allowed: true

  uk:
    currency: GBP
    commission_rate: {low: 0.03, mid: 0.05, high: 0.07}
    lead_fee_range: {min: 2, max: 50}
    subscription_pro: 14.99
    subscription_biz: 49.99
    urgent_fee: 4.99
    cash_jobs_allowed: false
```

Pricing reflects local purchasing power. Never just currency-convert.

---

## 7. Dispute Resolution

### Severity tiers

**Low — automated, no human needed:**
- Provider late by 15 minutes
- Minor price discrepancy
- Customer cancels last minute

**Medium — basic mediation:**
- "Work quality was poor"
- Provider didn't finish the job
- Customer refuses to pay after completion
- Price charged differs from quote

**High — needs investigation:**
- Property damage claim
- Provider no-show after customer waited hours
- Fraudulent provider / fake profile
- Harassment or safety issue

**Critical — legal / platform risk:**
- Theft
- Physical harm
- Regulatory violation

### Resolution flows

#### Tier 1: Automated

**Provider late:**
- Provider hasn't arrived 15 min after scheduled time
- Auto-SMS to customer: "Provider is running late. Reply 1=Wait 2=Cancel (no charge)"
- Auto-SMS to provider: "You're marked late. This affects your reliability score."
- 30+ min late and customer cancels: full refund (if prepaid), provider gets a late strike
- 3 strikes in a month = temporary search demotion

**Customer cancels last minute:**
- Cancel < 1 hour before: provider gets cancellation fee (small fixed amount), customer warned
- Cancel with 24h+ notice: no penalty

#### Tier 2: Structured mediation

**Work quality complaint flow:**
1. Customer rates job < 3 stars, selects what went wrong
2. Provider notified, can: accept fault / dispute / offer to redo
3. Resolution paths:
   - Provider offers redo, customer accepts — resolved
   - Provider accepts fault — partial/full refund
   - Both disagree — escalate to human mediator
   - No provider response in 48h — ruled in customer favor

**Customer won't pay (cash jobs):**
- Provider marks job complete, customer confirms or doesn't respond
- No response in 24h: system follows up via SMS/call
- Repeat non-payer: future bookings require prepayment, 3rd offense = suspension

**Overcharge vs. quote:**
- Digital payment: only quoted amount released from escrow, provider must justify additional charges
- Cash payment: customer reports overcharge, enters mediation, affects provider trust score

#### Tier 3: Human mediation

- Both parties submit evidence (photos, messages, receipts)
- Mediator reviews (platform team or trained local partners)
- Resolution: full refund, partial refund, provider compensated, warnings, or suspension
- SLA: resolved within 72 hours
- Mediators must speak local language — local partners/BPO preferred over centralization

#### Tier 4: Critical — safety and legal

- Account suspended pending investigation
- Local authorities notified if criminal
- Platform provides records if legally required
- Insurance claim triggered if coverage exists

### Escrow as dispute prevention

```
Customer books → money held by platform
  → Provider does work
    → Customer confirms completion
      → Money released to provider (minus commission)

During dispute: money stays held, resolution process runs,
money released to the party the ruling favors.
Auto-release after 7 days if customer goes silent.
```

For cash jobs, tools are limited to reputation consequences, account restrictions, and requiring future prepayment.

### SMS dispute flow for basic phone users

```
SMS: "Job #4521 done? Reply 1=Yes 2=Problem"

If provider replies 2:
"What happened? Reply:
 1=Customer not home
 2=Customer won't pay
 3=Job was different than described
 4=Other (call us)"

Option 4 → IVR connects to human support in local language
```

Maximum 3 SMS exchanges before routing to a human.

---

## 8. Monetization

### Revenue streams

#### 1. Commission on digital transactions

Split between both sides to keep it light:

```
Provider commission: 5%
Customer service fee: 3%
Total platform take: 8%
```

Commission scaling by job value:
- ₹0–2,000: 8% total
- ₹2,000–10,000: 5%
- ₹10,000+: 3%
- Trusted providers (Level 3+): additional 1% discount

#### 2. Lead fees (for cash-heavy markets)

Provider pays a small fee to accept a lead, keeps 100% of job payment.

- Plumber/electrician lead: ₹15–30
- Painter (bigger job): ₹50–100
- Lawyer/CA consultation: ₹100–300
- Home renovation: ₹200–500

Lead fee deducted from a prepaid wallet topped up via UPI, bank transfer, or local agent.

#### 3. Provider subscriptions

**Free tier (default):** Listed in search, up to 5 leads/month, standard profile.

**Pro tier (₹299/month):** Unlimited leads, priority search, "Pro" badge, analytics dashboard, lower commission (drop by 2%).

**Business tier (₹999/month):** Everything in Pro + multiple team member profiles, branded company page, quote templates, invoice generation, bulk job management.

Free tier must be genuinely usable. Subscription upsells only work after a provider is already earning on the platform.

#### 4. Customer-side premium features

- **Urgent booking (₹49):** Get a provider within 1 hour
- **Verified-only filter (₹29):** Show only ID-verified, background-checked providers
- **Job insurance (₹99):** If work is faulty within 30 days, platform sends another provider free
- **Concierge service (₹199):** Platform finds, vets, and schedules — customer just approves

#### 5. Promoted listings

- Top of search in postcode: ₹50/day
- Featured profile on category page: ₹500/month
- Sponsored badge in results: ₹200/month

Rules: clearly marked as "Sponsored," only available to 4+ star providers, max 2 sponsored results per search page.

#### 6. Financial services (phase 2)

**For providers:**
- Microloans for tools/equipment (partner with NBFC/fintech, earn referral fee)
- Liability insurance (partner with insurer, earn commission)
- Interest-bearing wallet for earnings (partner with bank, earn float)

**For customers:**
- BNPL — "Pay in 3 installments" for jobs over ₹10,000 (BNPL partner, earn referral fee)
- Home maintenance plan — ₹499/month for 2 free service calls + 10% off all bookings (recurring revenue, locks in customer)

### Revenue mix by maturity

**Year 1 (proving the model):** Lead fees 50%, Commission 30%, Urgent booking 20%

**Year 2 (scaling):** Commission 35%, Subscriptions 25%, Lead fees 20%, Customer premiums 15%, Promoted listings 5%

**Year 3+ (platform economics):** Commission 25%, Financial services 25%, Subscriptions 20%, Promoted listings 15%, Customer premiums 10%, Lead fees 5%

### What to avoid

- Don't charge providers to sign up — kills supply-side growth
- Don't take commission on cash jobs — can't enforce it, trying looks bad
- Don't make the free tier useless — if free providers get zero leads, they leave and tell others
- Don't hide fees — both sides see exactly what they pay before confirming

---

## 9. Disintermediation and Retention

### The reality

~40-60% of customers will save a provider's number and contact them directly after the first job. This is unavoidable and attempting to prevent it with punitive measures (hiding phone numbers, threatening bans) accelerates churn.

### Strategy: make staying feel better than leaving

**What direct contact lacks that the platform provides:**
- Payment protection / escrow
- Dispute resolution
- Scheduled, tracked, and reminded bookings
- Full job history and invoices
- Automated reminders for recurring service
- Provider replacement guarantee if someone cancels
- Job insurance / warranty

### Recurring work — the key retention lever

#### Maintenance plans

Customers set up recurring jobs:
- Home cleaning on the 1st and 15th, auto-assigned to their preferred provider
- AC servicing every 3 months, auto-pay enabled
- Annual plumbing inspection with 7-day advance reminder

**Customer value:** "Set it and forget it" — no calling, no remembering, no negotiating.

**Provider value:** Guaranteed recurring income. Platform becomes their booking system.

**Commission on recurring work is lower** (2-3% instead of 5-8%) — volume and predictability compensate.

#### Loyalty pricing

```
1st booking with a provider:    8% total fee
2nd booking (same provider):    6%
3rd booking:                    4%
4th+ booking:                   3%
Recurring plan:                 2%
```

The gap between on-platform and off-platform cost becomes so small it's not worth the hassle of going direct.

#### Service history as a product

The platform maintains a home maintenance log:
- Every job done, by whom, when, and how much
- Warranty expiry dates
- Upcoming service reminders
- Transferable to new homeowner if they sell

This has standalone value that going direct doesn't provide.

#### Provider replacement guarantee

"Book through us and if your provider cancels, we'll find a replacement within 2 hours. Book directly? You're on your own."

#### Financial tools for providers

- Instant settlement (vs. waiting for cash from customer)
- Income advances against booked jobs
- Earnings tracking and tax summaries
- Micro-loans based on platform history

If the platform is where a provider's income is stable, tracked, and accessible, leaving costs them real money.

### The honest math

Out of 100 customers:
- 40 leave after first job — earned one-time commission, some return for a different service category
- 35 stay for a few jobs, then drift — earned 3-5 commissions, recurring reminders may pull them back
- 25 become regular platform users — recurring revenue, subscriptions, financial services; worth more than the other 75 combined

The business model works by making the retained 25% extremely valuable, not by preventing leakage.

---

## 10. Technology Architecture

### High-level system design

```
                    ┌──────────────┐
                    │   CDN/Edge   │
                    │  (Cloudflare)│
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼────┐ ┌────▼─────┐ ┌────▼─────┐
        │ Web/PWA  │ │ Mobile   │ │ SMS/IVR  │
        │(Svelte-  │ │(Flutter) │ │ Gateway  │
        │  Kit)    │ │          │ │          │
        └─────┬────┘ └────┬─────┘ └────┬─────┘
              │            │            │
              └────────────┼────────────┘
                           │
                    ┌──────▼───────┐
                    │  API Gateway │
                    │  Rate limit  │
                    │  Auth/Routing│
                    └──────┬───────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
┌──────▼──────┐    ┌───────▼──────┐    ┌───────▼──────┐
│ User Layer  │    │  Job Layer   │    │ Payment Layer│
│ ┌─────────┐ │    │ ┌──────────┐ │    │ ┌──────────┐ │
│ │ User    │ │    │ │ Job      │ │    │ │ Payment  │ │
│ │ Service │ │    │ │ Service  │ │    │ │ Service  │ │
│ ├─────────┤ │    │ ├──────────┤ │    │ ├──────────┤ │
│ │ Auth    │ │    │ │ Matching │ │    │ │ Escrow   │ │
│ │ Service │ │    │ │ Service  │ │    │ │ Service  │ │
│ ├─────────┤ │    │ ├──────────┤ │    │ ├──────────┤ │
│ │ Profile │ │    │ │ Review   │ │    │ │ Payout   │ │
│ │ Service │ │    │ │ Service  │ │    │ │ Service  │ │
│ └─────────┘ │    │ ├──────────┤ │    │ ├──────────┤ │
│             │    │ │ Dispute  │ │    │ │ Wallet   │ │
│             │    │ │ Service  │ │    │ │ Service  │ │
│             │    │ └──────────┘ │    │ └──────────┘ │
└──────┬──────┘    └───────┬──────┘    └───────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
       ┌──────▼────┐ ┌────▼─────┐ ┌────▼──────┐
       │PostgreSQL │ │  Redis   │ │Meilisearch│
       │(primary)  │ │(cache,   │ │(search +  │
       │+ PostGIS  │ │ queues,  │ │ fuzzy     │
       │           │ │ sessions)│ │ matching) │
       └───────────┘ └──────────┘ └───────────┘
```

### Core services

**1. User Service** — registration (customer/provider/company), profile management, KYC/ID verification, jurisdiction-aware fields, trust score calculation.

**2. Auth Service** — OTP-based login via SMS (works on any phone), social login (Google, Apple for smartphone users), session management, role-based access (customer, provider, admin, mediator).

**3. Matching Service** — the brain of the platform. Postcode/geolocation search, category + skill filtering, availability check, trust score weighting, price range matching, language matching, response time prediction, smart ranking.

**4. Job Service** — job posting + categorization, lifecycle management (posted → matched → accepted → in-progress → completed → reviewed), scheduling (one-time + recurring), quote management, job history + home profile.

**5. Communication Service** — in-app chat (smartphone), SMS relay (basic phone: customer sends message in app → provider receives SMS → provider replies → customer sees in app), IVR call routing, push notifications, transactional email.

**6. Payment Service** — gateway adapter, escrow hold + release, commission calculation, tax computation per jurisdiction, invoice generation, refund processing, cash job tracking.

**7. Payout Service** — provider bank/wallet payouts, settlement scheduling (instant/daily/weekly), reconciliation, provider wallet for lead fee prepayment.

**8. Review & Rating Service** — post-job ratings (bidirectional), review moderation, trust score aggregation, review translation.

**9. Dispute Service** — automated resolution (tier 1), mediation workflow (tier 2), evidence collection, escalation routing, resolution tracking.

**10. Notification Service** — push notification orchestration, SMS gateway abstraction, email templates, IVR triggers, per-user preferences, jurisdiction-aware templates.

**11. Gamification Service** — points ledger, level progression engine, achievement tracking, leaderboard computation, reward disbursement.

**12. Admin Service** — ops dashboard (metrics, disputes, KYC queue), provider approval workflow, content moderation, jurisdiction management, feature flags per market.

### Matching algorithm

```
matchProviders(jobRequest):

  1. FILTER (hard constraints — must pass all)
     - Postcode within service radius
     - Skill matches job category
     - Provider is active and not suspended
     - Provider available at requested time
     - Provider serves this job type

  2. SCORE (soft ranking — weighted)
     - Distance (closer = higher)           25%
     - Trust score                          25%
     - Response time history                15%
     - Price competitiveness                15%
     - Completion rate                      10%
     - Language match with customer          5%
     - Platform loyalty                      5%

  3. BOOST (contextual)
     - "Urgent" → boost fast responders
     - "Budget" → boost lower price
     - "Quality" → boost high-rated
     - Subscription provider → small boost
     - Promoted listing → top 2 slots (marked sponsored)

  4. RETURN top 5-10 results

  For SMS/basic phone providers:
     Send job to top 3 via SMS, first to accept gets it.
     No response in 30 min → expand radius and retry.
```

### Key database entities

**users** — id, type (customer/provider/company), phone, jurisdiction_id, preferred_language, device_type, created_at.

**provider_profiles** — user_id, skills[], service_radius_km, postcode, lat/lng, trust_score, level, verification_status, subscription_tier, availability_schedule, bank_account/wallet_id.

**jobs** — id, customer_id, provider_id, category_id, postcode, lat/lng, status, scheduled_at, quoted_price, final_price, payment_method, is_recurring, recurrence_rule, jurisdiction_id.

**transactions** — job_id, amount, currency, commission_amount, tax_amount, provider_payout_amount, payment_gateway_ref, escrow_status, settled_at.

**reviews** — job_id, reviewer_id, reviewee_id, rating, comment, language, moderation_status.

**disputes** — job_id, raised_by, type, severity_tier, status, resolution, resolved_by, evidence[].

### Tech stack

> Full rationale and library choices in **Section 24 — Technology Foundation**.

**Frontend (Web + Admin):** SvelteKit — SSR for SEO, PWA for offline, single codebase for customer, provider, and admin interfaces. Paraglide (Inlang) for compile-time i18n.

**Frontend (Mobile):** Flutter/Dart — customer app (~8 MB) and provider app (~5 MB) as separate targets sharing a core business-logic package. Offline-first with local SQLite + background sync.

**Backend:** Go with Fiber — single compiled binary, ~5 MB RAM per instance. sqlc for type-safe SQL, Asynq (Redis-backed) for background jobs. OpenAPI 3.1 as single source of truth generating Go server stubs, SvelteKit TypeScript client, and Flutter Dart client.

**Data:** PostgreSQL with PostGIS (primary + geo queries), Redis (cache, queues via Asynq, sessions), Meilisearch (provider search, autocomplete, fuzzy matching, typo-tolerant), S3-compatible storage (files, photos).

**Infrastructure:** Docker + Kubernetes, GitHub Actions (CI/CD), Prometheus + Grafana (monitoring), Sentry (error tracking), structured JSON logging.

**SMS:** Twilio (global) + local providers (MSG91 for India, etc.) via adapter pattern.

### Three interface tiers

**Customer (smartphone app / web):** Full experience — search, book, pay, review.

**Provider — smartphone (app):** Accept jobs, view earnings, track level, manage schedule.

**Provider — basic phone (SMS/IVR):**
```
SMS: "New job: Fix tap at 4pm, 2km away. Reply 1=Accept 2=Decline"
Earnings and level sent weekly via SMS.
```

### Offline and low-connectivity handling

**Basic phone providers:** Everything works via SMS — no connectivity issue.

**Smartphone providers (poor connection):** Job notifications cached and delivered when online. Accept/decline queued offline, synced on reconnect. Grace period for response time scoring.

**Customers (poor connection):** Last search results cached. Booking confirmation queued. Job status viewable offline (last known state). Critical actions queued and synced.

---

## 11. Provider Verification / KYC

**Tier 1 — Basic (required):**
- Phone number verified via OTP
- Government ID uploaded (Aadhaar, passport, etc.)
- Selfie match with ID (automated via API)

**Tier 2 — Verified (optional, earns badge):**
- Address proof
- Skill certificate (if applicable)
- Background check (criminal records API)
- Reference from existing verified provider

**Tier 3 — Licensed (for regulated professions):**
- Professional license (lawyers, CAs, electricians in jurisdictions requiring it)
- Business registration / GST
- Insurance proof

**For basic phone providers:** Field agent does verification in person — takes photo of ID, fills form on their device. Provider signs consent via thumbprint or OTP.

---

## 12. Fraud Detection

**Provider fraud:**
- Fake profiles (same phone, multiple accounts)
- Fake reviews (ring of providers rating each other)
- Accepting jobs and not showing up (lead fee scam)
- Bait-and-switch pricing
- Detection: velocity checks, review graph analysis, GPS verification at job site

**Customer fraud:**
- Fake complaints for free service
- Chargeback abuse
- Multiple accounts for referral bonus farming
- Detection: device fingerprinting, behavior patterns, refund rate monitoring

**Platform-level:**
- SMS OTP abuse (bots creating accounts)
- Scraping provider data
- Rate limiting + CAPTCHA at registration

---

## 13. B2B Angle

Beyond individual customers, the platform serves business clients with higher-value, lower-churn contracts:

**Housing societies:** "We need a plumber on retainer for 200 flats" — bulk contract, monthly subscription.

**Property management companies:** "Manage maintenance for 50 rental properties" — dashboard, SLA tracking, automated scheduling.

**Offices / co-working spaces:** Regular cleaning, AC maintenance, electrical — multi-category recurring contracts.

**Real estate developers:** Warranty service for new homeowners — developer pays, homeowner books through platform.

Separate onboarding track and dashboard for B2B clients.

---

## 14. Cold Start Strategy

### Go hyperlocal

**Phase 1 — one neighborhood / postcode:**
- Physically onboard 20-30 providers (door to door)
- Give them free subscription for 6 months
- Run local ads (WhatsApp groups, community boards, flyers)
- Target: 100 jobs in first month in that postcode

**Phase 2 — adjacent postcodes:**
- Use existing provider referrals (point incentive)
- Use customer referrals
- Expand radius gradually

**Phase 3 — city-wide:**
- Digital marketing kicks in
- SEO for "plumber near me in [city]"
- Partnerships with housing societies

**Do not** launch in 10 cities simultaneously. Dominate one neighborhood first.

---

## 15. SEO and Organic Discovery

Auto-generated landing pages:

```
/electrician/bangalore/koramangala
/plumber/mumbai/andheri-west
/lawyer/delhi/connaught-place
```

Each page includes:
- Top-rated providers in that postcode
- Average pricing
- Recent reviews
- "Book now" CTA
- FAQ schema markup for Google rich results

This drives organic traffic at scale. Build it into the platform from day one.

---

## 16. Provider Scheduling and Availability

**Smartphone providers:**
- Weekly schedule (Mon-Fri 9am-6pm)
- Block specific dates (holidays, personal)
- Real-time toggle (online/offline)
- Auto-offline after accepting max jobs per day

**Basic phone providers:**
```
SMS: "Reply 1 to go offline today, 2 to stay available"
Weekly: "Set your hours for next week. Reply like: MON-FRI 9-6"
```

---

## 17. Analytics for Providers

Providers who see their data stay on the platform longer.

**Weekly SMS to basic phone providers:**
```
"Your week: 8 jobs, ₹12,400 earned
 Rating: 4.7
 Rank: #3 electrician in your area"
```

**Smartphone app dashboard:**
- Earnings graph (daily/weekly/monthly)
- Demand heatmap for their area
- Peak hours analysis
- Customer retention rate
- Comparison with category average (anonymized)

---

## 18. Safety Features

**For customers:**
- Provider live tracking en route
- Share job details with emergency contact
- In-app SOS button (alerts platform + local emergency number)
- Provider digital ID card (verifiable)
- OTP-based job start (provider shares OTP at door, customer confirms identity)

**For providers:**
- Customer rating visible before accepting a job
- Decline without penalty (within reason)
- Route shared with emergency contact
- Harassment reporting

---

## 19. Content and Education

**For providers (retention + upskilling):**
- "How to grow your business on the platform"
- Skill certification courses (partner with training providers)
- Tax filing guidance for freelancers
- Financial literacy content

**For customers (SEO + trust):**
- "How to prepare for a plumber visit"
- "What should an electrician charge for X?"
- Maintenance tips / seasonal reminders
- Cost estimator tools

---

## 20. Partnerships and Integrations

**Supply partnerships:**
- Hardware stores — providers buy materials, charge to customer via platform
- Training institutes — certified providers get boosted listing
- Insurance companies — liability coverage sold through platform

**Demand partnerships:**
- Real estate portals — "just moved? find services" integration
- Banking apps — "home loan approved? book a painter"
- Smart home devices — "AC filter alert → book servicing"
- Housing society apps — embedded booking widget

**Infrastructure partnerships:**
- Telecom providers — zero-rated SMS for platform messages
- Digital payment wallets — co-branded offers
- Map providers — accurate postcode/address data

---

## 21. Trust Score System

### Provider trust score (internal)

| Factor | Weight |
|---|---|
| On-time rate | 25% |
| Completion rate | 25% |
| Customer ratings | 30% |
| Dispute rate | 15% |
| Response time | 5% |

**Effects:**
- Score > 90: "Trusted" badge, lower commission
- Score 70-90: Normal listing
- Score 50-70: Deprioritized in search
- Score < 50: Suspended, must re-verify

### Customer trust score (simpler)

- Payment reliability
- Cancellation rate
- Fair rating history

Bad customer score is visible to providers, who can decline jobs accordingly. Both sides are accountable.

---

## 22. Complete Service Catalog

### Category activation logic

The full catalog is built into the platform, but categories are only shown to users when providers exist nearby:

```
IF providers_in_radius >= 3:
  → Show category normally

IF providers_in_radius == 1-2:
  → Show with note: "Limited availability"

IF providers_in_radius == 0:
  → Hide from browse
  → Allow "Request this service"
    (customer registers interest → demand signal
     for where to onboard providers next)

IF "Request this service" hits 20+ requests in a postcode:
  → Priority onboarding target for that area
```

Customers never see an empty category. The platform looks full everywhere because it only shows what's actually available.

### Category hierarchy

```
TOP LEVEL (what users see):
  ├── Home Repair & Maintenance
  ├── Cleaning
  ├── Beauty & Wellness
  ├── Professional Services
  ├── Vehicle Services
  ├── Education & Tutoring
  ├── Care Services
  ├── Events & Occasions
  ├── Moving & Logistics
  ├── Tech & Digital
  ├── Crop & Land Services        ← jurisdiction-specific
  └── Construction & Civil
```

### 22.1 Home Repair & Maintenance

- **Plumbing** — tap/faucet, pipe leak, toilet, water heater/geyser, drainage/blockage, water purifier, bathroom fitting
- **Electrical** — wiring/rewiring, switch/socket, fan install/repair, inverter/UPS, earthing, MCB/fuse, lighting
- **Carpentry** — furniture repair, door/window fix, custom furniture, cabinet/shelf, wood polishing
- **Painting** — interior, exterior, texture/decorative, wood/metal, waterproof coating
- **Appliance repair** — AC, washing machine, refrigerator, microwave/oven, TV, chimney, dishwasher
- **Masonry & civil** — wall repair/plastering, tile work, flooring, waterproofing, false ceiling, demolition
- **Roofing** — repair, waterproofing, installation (sheet/tile)
- **Welding & fabrication** — gate/grill, railing, structural welding
- **Glass work** — window glass, mirror, glass partition
- **Locksmith** — lock repair/replacement, key duplication, digital lock
- **Pest control** — general (cockroach, ant), termite, mosquito, bed bug, rodent, snake removal

### 22.2 Cleaning

Home cleaning (regular/deep/move-in-out), kitchen deep cleaning, bathroom deep cleaning, sofa/upholstery, carpet, water tank, overhead tank/sump, window cleaning, post-construction cleanup, office cleaning, industrial cleaning.

### 22.3 Beauty & Wellness

- **Salon at home (women)** — haircut/styling, facial, waxing, manicure/pedicure, threading, bleach/detan
- **Salon at home (men)** — haircut, shave/beard trim, facial
- **Wellness** — massage/spa at home, mehendi, makeup artist, personal trainer, yoga instructor, dietitian/nutritionist

### 22.4 Professional Services

- **Legal** — lawyer (civil, criminal, family, property, labour), notary, legal document drafting
- **Finance** — chartered accountant, tax filing (individual/business), GST registration/filing, audit, company registration, financial advisor
- **Architecture & design** — architect, interior designer, vastu consultant, structural engineer
- **Documentation** — document writer/typist, translation/interpreter, stamp paper/affidavit, passport/visa assistance
- **Consulting** — business consultant, HR/recruitment, real estate agent

### 22.5 Vehicle Services

Car mechanic (general service, engine, AC, body/denting, electrical), two-wheeler mechanic, car wash/detailing (doorstep), tyre puncture/replacement, battery jumpstart/replacement, towing, vehicle inspection, driving instructor, driver on demand, vehicle registration/RTO work.

### 22.6 Education & Tutoring

Academic tutor (primary, high school), exam preparation, music teacher (vocal, guitar, piano, drums, classical), dance teacher, art/drawing, language tutor (English, Hindi, regional, foreign), computer/IT tutor, sports coaching (cricket, swimming, tennis, martial arts), special needs educator.

### 22.7 Care Services

Elderly care/attendant, baby sitter/nanny, home nurse, physiotherapist, occupational therapist, speech therapist, mental health counselor, ambulance, pet grooming, pet sitting/walking, veterinary (home visit), pet training.

### 22.8 Events & Occasions

Catering (small party, large event, specific cuisine, live cook), decoration (birthday, wedding, festival, balloon/flower), tent/pandal/canopy, sound/DJ/music, photography (event, portrait, product, drone), videography, anchor/emcee, priest/pandit/religious services, wedding planning.

### 22.9 Moving & Logistics

Packers and movers (local, intercity, office), furniture assembly/disassembly, courier (local same-day), junk removal/disposal, scrap pickup, storage (short-term).

### 22.10 Tech & Digital

Computer/laptop repair, mobile phone repair, CCTV installation, security system setup, WiFi/network setup, printer setup/repair, smart home setup, data recovery, website help (small business), social media management, graphic design.

### 22.11 Crop & Land Services (see Section 23 for full framework)

Jurisdiction-specific crop maintenance, harvesting, land management, and agricultural support services. Fully detailed in the next section.

### 22.12 Construction & Civil

General contractor (new build), civil engineer, surveyor/land measurement, soil testing (construction), borehole/foundation, scaffolding, ready-mix concrete, solar panel installation, rainwater harvesting.

### Category metadata schema

Every category carries structured metadata:

```
category:
  id: "plumbing_tap_repair"
  parent: "plumbing"
  top_level: "home_repair"

  display:
    names:
      en: "Tap / Faucet Repair"
      hi: "नल की मरम्मत"
      ml: "ടാപ്പ് റിപ്പയർ"
      # added per jurisdiction

  matching:
    urgency_level: "high"
    typical_duration: "30-60min"
    requires_site_visit: true
    multiple_providers_needed: false
    service_model: "on_demand"

  pricing:
    model: "fixed_or_quote"
    typical_range:
      INR: [200, 800]
      GBP: [40, 120]
      AED: [50, 200]

  provider_requirements:
    min_verification: "tier_1"
    license_required: false
    tools_required: true

  jurisdiction_availability:
    india_urban: true
    india_rural: true
    uk: true
    uae: true

  sms_template:
    en: "New job: {category} at {time}, {distance}km.
         ~{currency}{price}. Reply 1=Accept 2=Decline"
```

### What the platform does NOT cover

These need fundamentally different platform models:

- **Daily wage farm labour** — labour marketplace, bulk hiring, crew management
- **Heavy machinery rental** — equipment rental marketplace with availability calendars
- **Cold storage / warehousing** — facility-based, not provider-visits-customer
- **Freight / transport** — logistics platform
- **Full-time domestic employment** (live-in maid, full-time cook) — staffing platform, not on-demand
- **Medical diagnosis/treatment** — heavy regulation, liability; physiotherapy and counseling are fine

---

## 23. Crop & Land Services — Generalized Framework

### The universal problem

Across every geography, households and smallholdings have crops, trees, and land that need regular maintenance. The workers who do this work are:

- Informal, operating on word of mouth
- Often don't have smartphones
- Have fixed routes or seasonal availability
- Aging workforce with few replacements entering the trade
- Invisible to any digital platform

Finding them is the same pain everywhere — ask neighbors, call broken numbers, wait weeks. The platform solves this by being the first place that maps this invisible workforce to the households that need them.

### How it works: jurisdiction-specific crop catalogs

Every jurisdiction defines its own crop catalog. Crops, the work they require, the workers who do it, seasonal patterns, and pricing — all are config, not code.

```
FRAMEWORK:

  jurisdiction:
    region_id: "kerala" | "tamil_nadu" | "uk_rural" | ...

    crops:
      - crop_id
      - local_names (multi-language)
      - prevalence (how common in this region)
      - services required (list of work types)
      - service_model (route_based | seasonal | on_demand | contract)
      - seasonal_calendar (when each service is needed)
      - worker_common_name (local term for the worker)
      - typical_pricing_model (per_tree | per_acre | per_day | per_visit)
      - frequency (how often the service is needed)
```

### Global crop-work catalog

The platform maintains a master catalog of crop types and the work they generate. Each jurisdiction activates the crops relevant to it and adds local names, pricing, and seasonal data.

```
TREE CROPS (perennial — recurring maintenance):
┌─────────────────┬──────────────────────────────────┬────────────────────┐
│ Crop            │ Work required                    │ Common regions     │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Coconut         │ Climbing & plucking              │ Kerala, TN, KA,    │
│                 │ De-husking                       │ Goa, coastal India,│
│                 │ Frond/leaf removal               │ Sri Lanka, SEA,    │
│                 │ Tree health treatment            │ Pacific Islands    │
│                 │ Pest treatment                   │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Arecanut        │ Climbing & plucking              │ Kerala, Karnataka, │
│                 │ Processing/drying                │ Assam, Meghalaya   │
│                 │ Tree maintenance                 │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Rubber          │ Tapping                          │ Kerala, Tripura,   │
│                 │ Sheet making                     │ TN hills, SEA,     │
│                 │ Rain guarding                    │ West Africa        │
│                 │ Tree treatment                   │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Coffee          │ Pruning                          │ Karnataka, Kerala, │
│                 │ Harvesting (picking)             │ TN, Ethiopia,      │
│                 │ Processing (pulping, drying)     │ Colombia, Vietnam  │
│                 │ Shade tree management            │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Tea             │ Pruning                          │ Assam, Darjeeling, │
│                 │ Plucking                         │ Kerala, Sri Lanka, │
│                 │ Bush maintenance                 │ Kenya, China       │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Mango           │ Pruning                          │ Nearly all tropical│
│                 │ Pest spraying                    │ and subtropical     │
│                 │ Harvesting                       │ regions            │
│                 │ Grafting                         │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Cashew          │ Harvesting                       │ Goa, Kerala, KA,   │
│                 │ Processing                       │ Maharashtra, TN,   │
│                 │ Pruning                          │ West Africa        │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Jackfruit       │ Harvesting                       │ Kerala, Karnataka, │
│                 │ Pruning                          │ SEA                │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Citrus          │ Pruning, pest control,           │ Maharashtra, NE    │
│ (orange, lemon) │ harvesting, grafting             │ India, Mediterranean│
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Date palm       │ Pollination, bunch thinning,     │ Middle East, North │
│                 │ harvesting, frond trimming       │ Africa, Rajasthan  │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Oil palm        │ Harvesting (bunch cutting),      │ Malaysia, Indonesia│
│                 │ pruning, pest control            │ West Africa, NE Ind│
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Olive           │ Pruning, harvesting,             │ Mediterranean,     │
│                 │ pest control                     │ Middle East        │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Apple / pear    │ Pruning, thinning, harvesting,   │ Kashmir, HP,       │
│                 │ pest spraying, grafting          │ Uttarakhand, EU, US│
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Pomegranate     │ Pruning, pest control,           │ Maharashtra, KA,   │
│                 │ harvesting                       │ Rajasthan, Iran    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Banana          │ Planting, de-suckering,          │ Nearly all tropical│
│                 │ bunch covering, harvesting,      │ regions            │
│                 │ post-harvest cleanup             │                    │
└─────────────────┴──────────────────────────────────┴────────────────────┘

SPICE CROPS (often grown alongside tree crops):
┌─────────────────┬──────────────────────────────────┬────────────────────┐
│ Pepper          │ Vine training, harvesting,       │ Kerala, Karnataka, │
│                 │ drying, disease treatment        │ Vietnam, Brazil    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Cardamom        │ Harvesting, curing/drying,       │ Kerala, Karnataka, │
│                 │ shade management, pest control   │ Guatemala          │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Clove           │ Harvesting, drying               │ Kerala, TN,        │
│                 │                                  │ Indonesia, Zanzibar│
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Nutmeg          │ Harvesting, mace separation,     │ Kerala, Karnataka, │
│                 │ drying                           │ Indonesia          │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Vanilla         │ Hand pollination, harvesting,    │ Kerala, Karnataka, │
│                 │ curing                           │ Madagascar         │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Turmeric /      │ Planting, harvesting, drying,    │ TN, AP, Karnataka, │
│ Ginger          │ processing                       │ NE India           │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Cinnamon        │ Bark stripping, pruning, drying  │ Kerala, Sri Lanka  │
└─────────────────┴──────────────────────────────────┴────────────────────┘

FIELD / ROW CROPS (seasonal — planted and harvested):
┌─────────────────┬──────────────────────────────────┬────────────────────┐
│ Paddy / rice    │ Land prep, transplanting,        │ All of Asia, parts │
│                 │ weeding, harvesting, threshing   │ of Africa, Americas│
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Wheat           │ Sowing, irrigation, harvesting   │ Punjab, Haryana,   │
│                 │                                  │ MP, UP, EU, US     │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Sugarcane       │ Planting, weeding, harvesting,   │ Maharashtra, UP,   │
│                 │ transport to mill                │ Karnataka, Brazil  │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Cotton          │ Sowing, pest spraying, picking   │ Gujarat, MH, AP,   │
│                 │                                  │ Rajasthan, US      │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Vegetables      │ Land prep, sowing/transplanting, │ Everywhere         │
│ (general)       │ weeding, pest control, harvest   │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Groundnut       │ Sowing, weeding, harvesting,     │ Gujarat, AP, TN,   │
│                 │ drying                           │ Rajasthan          │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Tapioca/cassava │ Planting, weeding, harvesting    │ Kerala, TN, NE,    │
│                 │                                  │ Africa, SEA        │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Tobacco         │ Transplanting, curing,           │ AP, Gujarat, KA    │
│                 │ harvesting                       │                    │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Grapes          │ Pruning, training, pest control, │ Maharashtra, KA,   │
│                 │ harvesting                       │ Mediterranean, US  │
└─────────────────┴──────────────────────────────────┴────────────────────┘

PLANTATION / COMMERCIAL:
┌─────────────────┬──────────────────────────────────┬────────────────────┐
│ Cocoa           │ Harvesting, fermenting, drying   │ Kerala, KA, W.     │
│                 │                                  │ Africa, S. America │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Betel leaf/nut  │ Vine maintenance, harvesting     │ Kerala, WB, Bihar, │
│                 │                                  │ Assam              │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Bamboo          │ Harvesting, management           │ NE India, Kerala,  │
│                 │                                  │ SEA                │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Sapota /        │ Pruning, harvesting,             │ Gujarat, MH, KA,   │
│ Chikoo          │ pest control                     │ SEA                │
├─────────────────┼──────────────────────────────────┼────────────────────┤
│ Guava           │ Pruning, pest control,           │ UP, Bihar, MH,     │
│                 │ harvesting                       │ tropical regions   │
└─────────────────┴──────────────────────────────────┴────────────────────┘

LAND & PROPERTY MAINTENANCE (not crop-specific):
  ├── Compound/boundary wall maintenance
  ├── Grass cutting / brush clearing
  ├── Weed removal
  ├── Landscaping / garden maintenance
  ├── Tree cutting (non-crop — safety, clearance)
  ├── Fencing (wire, wooden, compound)
  ├── Drainage / bund maintenance
  ├── Well cleaning / maintenance
  ├── Pond cleaning / de-silting
  └── Land leveling / earthwork

SUPPORT SERVICES (crop-agnostic):
  ├── Bore well drilling & repair
  ├── Irrigation setup & repair (drip, sprinkler, canal)
  ├── Crop spraying (pesticide, fungicide, fertilizer)
  ├── Soil testing
  ├── Agricultural consulting / agronomy
  ├── Organic farming consultation
  ├── Greenhouse / polyhouse setup & maintenance
  ├── Farm pond construction
  ├── Composting / vermicomposting setup
  ├── Veterinary services (farm animals)
  ├── Animal husbandry consulting
  └── Rainwater harvesting (agricultural)
```

### Service models for crop work

Not all crop work is on-demand. The platform supports four distinct service models:

```
1. ROUTE-BASED
   Used for: Coconut plucking, arecanut harvesting

   How it works:
   - Worker has a fixed circuit of properties
   - Visits each every N days (e.g., 45-60 for coconut)
   - Platform manages the route, fills gaps with new customers
   - Customers get notified of approximate visit date
   - Worker sees optimized route (map on app, list on SMS)

   Platform features needed:
   - Route planning and optimization
   - Approximate schedule notifications to customers
   - "Add me to route" request from new customers
   - Route gap detection ("3 houses near your Tuesday
     stop are requesting service")
   - Seasonal interval adjustment (monsoon = longer gaps)

2. SEASONAL CONTRACT
   Used for: Rubber tapping, paddy transplanting

   How it works:
   - Worker assigned to a property for an entire season
   - Daily or alternate-day work for weeks/months
   - Platform facilitates matching at season start
   - Tracks attendance, output, and payment

   Platform features needed:
   - Contract creation (start date, end date, terms)
   - Attendance tracking (worker checks in daily)
   - Output logging (kg of rubber, area transplanted)
   - Payment scheduling (weekly/monthly settlement)
   - Season-start matching (reminders sent to both
     parties 30 days before season)

3. ON-DEMAND
   Used for: Tree cutting, bore well repair, pest spraying

   How it works:
   - Standard marketplace flow
   - Customer posts need → matched with provider → job done
   - Same as urban services

   No special features needed beyond standard job flow.

4. PERIODIC / ANNUAL
   Used for: Pepper harvesting, mango picking, pruning

   How it works:
   - Needed once or a few times per year
   - Platform sends reminders when season approaches
   - Customer confirms need → matched with available workers
   - Often needs a crew (multiple workers), not individual

   Platform features needed:
   - Seasonal calendar per crop per jurisdiction
   - Pre-season reminders to customers
   - Crew booking (request 3-5 workers for a job)
   - Crew leader as single point of contact
```

### Seasonal calendar system

Every crop has a regional calendar. The platform uses this to proactively remind customers and prepare supply:

```
SEASONAL CALENDAR CONFIG:

jurisdiction: "kerala"
crop: "pepper"
calendar:
  jan:  { activity: "harvesting", demand: "high" }
  feb:  { activity: "harvesting/drying", demand: "high" }
  mar:  { activity: "drying/processing", demand: "medium" }
  apr:  { activity: "off-season", demand: "low" }
  may:  { activity: "vine training", demand: "medium" }
  jun:  { activity: "monsoon care", demand: "low" }
  jul:  { activity: "fertilizing", demand: "medium" }
  aug:  { activity: "pest watch", demand: "medium" }
  sep:  { activity: "pest treatment", demand: "medium" }
  oct:  { activity: "pre-harvest prep", demand: "medium" }
  nov:  { activity: "early harvest begins", demand: "high" }
  dec:  { activity: "peak harvest", demand: "very_high" }

reminders:
  - trigger: "30 days before harvest season"
    to: customer
    message: "Pepper harvest season approaching.
              Book workers now to avoid the rush."
  - trigger: "season start"
    to: provider
    message: "Pepper harvesting season is here.
              You have 5 pending requests in your area."
```

This calendar is **config per jurisdiction per crop**. Adding a new region means filling in its seasonal calendar — the platform logic doesn't change.

### Worker discovery framework

For any jurisdiction, the platform finds crop/land workers through these channels:

```
1. CUSTOMER-LED ONBOARDING (most effective)
   - Customer searches for a service, no results nearby
   - Platform: "Know someone who does this?
     Register them and earn 25 points"
   - Customer provides worker's phone number
   - Platform sends SMS to worker in local language:
     "[Customer name] registered you on [Platform].
      Get more work near you. No app needed.
      Reply 1=Join 2=More info 3=No"
   - Worker joins with existing customer as first client
   - Lowest friction — trust comes from existing relationship

2. LOCAL GOVERNMENT PARTNERSHIPS
   - India: Panchayat / ward offices, Krishi Bhavan
   - UK: Parish councils, local authority
   - Any jurisdiction: the local administrative body that
     knows its workers
   - They have lists, registries, or informal knowledge
   - Field agent visits, gets introductions, registers workers
   - Incentive: digital directory of services for their area

3. COOPERATIVES AND AGRICULTURAL BODIES
   - Commodity boards (Coconut Development Board,
     Rubber Board, Coffee Board, Spices Board, etc.)
   - Farmer Producer Organizations (FPOs)
   - Self-help groups (Kudumbashree in Kerala,
     equivalent in other states/countries)
   - These have member rolls and community trust

4. TRAINING PROGRAM PARTNERSHIPS
   - Coconut climbing training programs (CDB, FoCT)
   - Rubber tapping training (Rubber Board)
   - Agricultural extension programs
   - Graduates need customers immediately
   - Auto-register every graduate on the platform
   - Pipeline of new supply in trades where supply is shrinking

5. SUPPLY CHAIN ENDPOINTS
   - Coconut oil mills → know who brings coconuts
   - Rubber dealers → know local tappers
   - Spice traders → know harvesters
   - Agricultural equipment shops → know workers
   - Fertilizer/pesticide dealers → know sprayers
   - Each is a referral node — small bonus per
     active worker onboarded

6. COMMUNITY NETWORKS
   - Places of worship (church, temple, mosque)
   - Community gathering points
   - Local shops and markets
   - QR code posters + announcements
   - WhatsApp groups (single message from trusted
     community leader)

GENERALIZED PROCESS FOR ANY NEW JURISDICTION:

  Step 1: IDENTIFY
    Talk to 20 households in the area.
    "What crop/land work do you struggle to
     find people for?"
    Map the top 5 services.

  Step 2: MAP CHANNELS
    Where do workers and customers currently
    find each other?
    List all local word-of-mouth networks,
    government bodies, cooperatives, training
    programs, and supply chain endpoints.

  Step 3: PARTNER
    Pick 2-3 channels for supply acquisition:
    - One government (panchayat/council)
    - One community (cooperative/SHG/religious body)
    - One commercial (dealer/mill/shop)

  Step 4: ACTIVATE CUSTOMER ONBOARDING
    "Register your existing worker" feature.
    Highest point reward in the system.
    SMS-based worker onboarding (no app needed).

  Step 5: CONFIGURE
    Add crop/service categories with local language
    names, service models, seasonal calendars,
    pricing norms, and worker common names.
```

### Route management system

For route-based services (coconut plucking, arecanut harvesting, etc.), the platform provides a route management feature that doesn't exist in any urban service marketplace:

```
ROUTE MANAGEMENT:

  Worker's route:
  ├── List of properties on their circuit
  ├── Approximate visit frequency per property
  ├── Optimized order (minimize travel)
  ├── Estimated visit dates for each property
  └── Capacity indicator (how many more can they take)

  Customer's view:
  ├── "Your provider will visit around [date]"
  ├── Option to reschedule
  ├── Backup provider if regular is unavailable
  └── Payment history per visit

  Platform intelligence:
  ├── Detect route gaps (geographic holes where
  │   demand exists but no provider visits)
  ├── Suggest new customers to workers based on
  │   proximity to existing route
  ├── Seasonal interval adjustment (monsoon =
  │   longer gaps, summer = shorter)
  ├── Automatic re-routing if a worker drops a
  │   property (assign to another worker's route)
  └── Demand heatmap for route-based services
       per postcode

  SMS flow for basic phone worker:

  Weekly SMS:
  "This week's visits:
   Mon: Anil (15 trees, Ollur)
   Wed: Priya (8 trees, Thrissur East)
   Thu: Suresh (20 trees, Kunnamkulam)

   New request near your Thu route:
   Vineeth (12 trees, 2km from Suresh).
   Reply 1=Add 2=Skip"
```

### Crew booking

Some crop work (harvesting, land clearing) needs multiple workers. The platform supports crew booking:

```
CREW BOOKING:

  Customer: "Need 4 workers for pepper harvesting,
             2 acres, approximately 3 days"

  Options:

  A) CREW LEADER MODEL
     - Platform matches with a crew leader
     - Crew leader brings their own team
     - Single point of contact for customer
     - Crew leader handles internal payment split
     - Platform charges lead fee or commission
       to crew leader only

  B) INDIVIDUAL ASSEMBLY
     - Platform assembles available workers individually
     - Customer sees each worker's profile/rating
     - Platform coordinates scheduling
     - Each worker tracked and paid separately
     - Higher platform overhead but more transparency

  Default to crew leader model where available —
  it mirrors how this work actually happens today.
```

### Pricing models for crop work

Different from urban services — crop work has unique pricing conventions:

```
PRICING MODELS:

  Per tree:
    Coconut plucking — ₹30-50 per tree
    Arecanut plucking — ₹20-40 per tree

  Per acre / per cent:
    Paddy transplanting — ₹X per acre
    Weeding — ₹X per cent
    Grass cutting — ₹X per cent

  Per day (daily wage):
    General agricultural labour
    Harvesting crews

  Per kg / per unit output:
    Rubber tapping — ₹X per kg of latex
    Coffee picking — ₹X per kg of berries

  Per visit (fixed):
    Pesticide spraying — ₹X per visit
    Soil testing — ₹X per test

  Seasonal contract:
    Rubber tapping for season — ₹X per month
    Regular compound maintenance — ₹X per month

  The platform supports all pricing models.
  Each crop category config specifies which
  model(s) apply.
```

### Jurisdiction config example — how it all comes together

```
jurisdiction:
  id: "kerala"
  languages: ["ml", "en"]
  currency: "INR"

  crop_catalog:
    coconut:
      names: { ml: "തെങ്ങ്", en: "Coconut" }
      prevalence: "very_high"
      services:
        - id: "coconut_plucking"
          names: { ml: "തേങ്ങയിടൽ", en: "Coconut plucking" }
          service_model: "route_based"
          frequency: "every 45-60 days"
          pricing_model: "per_tree"
          typical_price_range: [30, 50]
          worker_name: { ml: "തെന്നുപറ", en: "Coconut climber" }
          urgency: "low"
        - id: "coconut_dehusking"
          names: { ml: "തേങ്ങ പൊളിക്കൽ", en: "De-husking" }
          service_model: "on_demand"
          pricing_model: "per_unit"
        - id: "coconut_tree_treatment"
          names: { ml: "തെങ്ങ് ചികിത്സ", en: "Tree treatment" }
          service_model: "on_demand"
          pricing_model: "per_tree"
      calendar:
        jan: { plucking: true, demand: "normal" }
        # ... (year-round for coconut)

    rubber:
      names: { ml: "റബ്ബർ", en: "Rubber" }
      prevalence: "high"
      services:
        - id: "rubber_tapping"
          names: { ml: "റബ്ബർ വെട്ട്", en: "Rubber tapping" }
          service_model: "seasonal_contract"
          season: { start: "June", end: "January" }
          pricing_model: "per_month"
          worker_name: { ml: "റബ്ബർ വെട്ടുകാരൻ", en: "Rubber tapper" }
      calendar:
        jun: { tapping: true, demand: "high", season_start: true }
        # ...
        feb: { tapping: false, demand: "none", off_season: true }

    pepper:
      names: { ml: "കുരുമുളക്", en: "Pepper" }
      # ...

  # Another jurisdiction — completely different crops

jurisdiction:
  id: "maharashtra_western"
  languages: ["mr", "en", "hi"]
  currency: "INR"

  crop_catalog:
    sugarcane:
      names: { mr: "ऊस", hi: "गन्ना", en: "Sugarcane" }
      prevalence: "very_high"
      services:
        - id: "sugarcane_harvesting"
          service_model: "seasonal"
          pricing_model: "per_acre"
          crew_required: true
          min_crew_size: 4
      calendar:
        oct: { harvesting: true, demand: "very_high" }
        # ...

    mango:
      names: { mr: "आंबा", en: "Mango" }
      services:
        - id: "mango_harvesting"
          service_model: "periodic"
          season: { start: "April", end: "June" }
        - id: "mango_pruning"
          service_model: "on_demand"
          season: { start: "September", end: "October" }

    grapes:
      names: { mr: "द्राक्षे", en: "Grapes" }
      prevalence: "high"  # Nashik belt
      # ...

jurisdiction:
  id: "uk_rural"
  languages: ["en"]
  currency: "GBP"

  crop_catalog:
    # UK doesn't have coconut or rubber, but has:
    hedgerow:
      names: { en: "Hedgerow" }
      services:
        - id: "hedge_trimming"
          service_model: "periodic"
          frequency: "2-3 times per year"
          pricing_model: "per_meter"

    orchard:
      names: { en: "Orchard" }
      services:
        - id: "fruit_picking"
          service_model: "seasonal"
          crew_required: true
        - id: "tree_pruning"
          service_model: "periodic"

    garden_general:
      names: { en: "Garden" }
      services:
        - id: "lawn_mowing"
          service_model: "route_based"
          frequency: "weekly/fortnightly"
        - id: "garden_maintenance"
          service_model: "route_based"
```

### Why this matters as a differentiator

No existing platform does this:

- **Urban Company / Thumbtack** — urban services only, no crop work
- **JustDial / Sulekha** — listings, not route management or seasonal calendars
- **FarmEasy / Agri apps** — focused on inputs (seeds, fertilizer) or commodity trading, not service workers
- **Labour platforms** — handle daily wage workers, not skilled route-based service providers

The platform uniquely maps the invisible workforce of crop and land service workers to the households that need them, with route management, seasonal intelligence, and SMS-first access — features that emerge directly from understanding how this work actually happens.

---

## 24. Technology Foundation

### Design principles

1. **Monorepo** — all code in one repository. Shared configs, atomic changes across services. No version drift between API spec and clients.
2. **API-first** — OpenAPI 3.1 spec is the single source of truth. Go server handlers, SvelteKit TypeScript client, and Flutter Dart client are all generated from this one file. Change the spec, regenerate everything — backend and frontends never go out of sync.
3. **Config-driven** — jurisdictions, categories, crops, pricing, languages are all YAML config. Adding a new market is a config change, not a code change.
4. **Offline-aware** — the system handles unreliable connectivity gracefully for both providers and customers.
5. **AI-augmented** — AI is embedded at every layer: conversational booking, photo-to-job, voice IVR, smart matching, real-time translation, fraud detection, and content generation. But a human is always reachable.

### Final tech stack

```
┌───────────┬──────────────────────────────────────────┐
│ Layer     │ Technology                               │
├───────────┼──────────────────────────────────────────┤
│ Web       │ SvelteKit 2 + Tailwind CSS + Paraglide   │
│ Mobile    │ Flutter (Dart) — customer + provider apps │
│ Provider  │ PWA (SvelteKit) + SMS/IVR                │
│ Admin     │ SvelteKit (same codebase, role-gated)    │
│ Backend   │ Go (Fiber) + sqlc + Asynq                │
│ AI Layer  │ Claude API + Deepgram + Google TTS       │
│ Database  │ PostgreSQL 16 + PostGIS                  │
│ Cache     │ Redis 7 (sessions, queues, cache, pubsub)│
│ Search    │ Meilisearch (provider/category search)   │
│ SMS       │ Twilio + MSG91 + regional (adapter)      │
│ Payments  │ Razorpay + Stripe + regional (adapter)   │
│ Storage   │ Cloudflare R2 (S3-compatible)            │
│ CDN       │ Cloudflare (edge caching, WAF, DDoS)    │
│ API spec  │ OpenAPI 3.1 (generates all clients)      │
│ CI/CD     │ GitHub Actions                           │
│ Infra     │ Terraform + Docker + Fly.io or AWS ECS   │
│ Monitoring│ Grafana + Prometheus + Sentry             │
│ Logging   │ zerolog → Grafana Loki                   │
│ ML models │ Python (train) → ONNX (serve in Go)     │
└───────────┴──────────────────────────────────────────┘
```

### Why these choices

```
WHY SVELTEKIT (not Next.js, Nuxt, or Angular):
  ├── Smallest JS output — compiles away, no framework
  │   runtime shipped to browser. 3-5x less JavaScript
  │   than React. Pages load in <1s on 3G connections.
  ├── Simplest code — less code = fewer bugs.
  │   No useEffect footguns, no dependency arrays,
  │   no stale closures. Reactivity just works.
  ├── Built-in everything — form actions, page transitions,
  │   loading states, error boundaries, prefetching,
  │   service workers / PWA support.
  ├── SSR/SSG/ISR — full support for SEO landing pages.
  ├── Vite under the hood — instant HMR, fast builds.
  ├── PWA for providers — same codebase serves customer
  │   web + provider PWA + admin dashboard.
  └── AI generates simpler, more correct Svelte code
      because the framework has fewer patterns to misuse.

WHY FLUTTER (not React Native or native):
  ├── Compiles to native ARM — smooth 60fps even on
  │   budget ₹5,000 phones with 2GB RAM.
  ├── No JS bridge bottleneck — Flutter renders directly,
  │   React Native goes JS → Bridge → Native views.
  ├── App size: 5-8MB (Flutter) vs 25-40MB (React Native).
  │   Matters when users have 16GB total storage.
  ├── Best offline support — Hive/Isar embedded DB,
  │   built for mobile, zero configuration.
  ├── Pixel-identical on iOS and Android — no platform
  │   rendering quirks to debug.
  └── Hot reload is sub-second and stateful.

WHY GO (not NestJS/Node.js, Django, or Spring):
  ├── 20x more throughput per server than Node.js
  │   (~300K req/sec vs ~15K req/sec).
  ├── 10-15x less memory (30MB vs 300-500MB per instance).
  │   Translates directly to 60-80% lower infra costs.
  ├── True parallelism — goroutines handle CPU-intensive
  │   work (matching algorithm, geo calculations) and
  │   I/O concurrently. Node.js single-thread can't.
  ├── Compiles to single static binary — 15MB Docker image
  │   vs 400MB for NestJS. Starts in <100ms vs 3-5s.
  ├── Explicit code — no decorators, no DI magic, no
  │   hidden behavior. Every handler reads top to bottom.
  │   AI generates more predictable, verifiable Go code.
  └── Statically typed and compiled — type errors caught
      at compile time, not runtime in production.

WHY POSTGRESQL (not MongoDB, MySQL, or CockroachDB):
  ├── PostGIS — best geospatial query support
  │   (provider radius, postcode matching, route optimization).
  ├── JSONB columns — flexible jurisdiction config storage.
  ├── Row-level security — multi-tenant data isolation.
  ├── Proven at scale (Instagram, Uber, Stripe use it).
  └── Strong ACID guarantees (critical for payments/escrow).

WHY MEILISEARCH (not Elasticsearch):
  ├── Designed for end-user search (not log analytics).
  ├── Runs on 256MB RAM vs Elasticsearch's 1GB+ minimum.
  ├── Built-in: typo tolerance, geo search, facets,
  │   multi-language tokenization, synonyms.
  ├── <50ms search responses.
  ├── Rust-based — fast, low resource usage.
  └── Costs $5/month vs $50+/month for managed ES.

WHY SQLC (not an ORM like Prisma/GORM):
  ├── You write actual SQL queries.
  ├── sqlc generates type-safe Go functions from them.
  ├── No ORM magic, no surprise N+1 queries.
  ├── PostGIS queries work naturally (just SQL).
  └── AI writes SQL directly — cleaner than fighting
      an ORM's abstraction layer.
```

### Monorepo structure

```
service-marketplace/
│
├── web/                          # SvelteKit app
│   ├── src/
│   │   ├── routes/
│   │   │   ├── (customer)/       # Customer pages
│   │   │   │   ├── search/
│   │   │   │   ├── book/
│   │   │   │   ├── jobs/
│   │   │   │   └── profile/
│   │   │   ├── (provider)/       # Provider PWA pages
│   │   │   │   ├── dashboard/
│   │   │   │   ├── jobs/
│   │   │   │   ├── earnings/
│   │   │   │   └── route/
│   │   │   ├── (admin)/          # Admin dashboard
│   │   │   │   ├── users/
│   │   │   │   ├── disputes/
│   │   │   │   ├── kyc/
│   │   │   │   └── analytics/
│   │   │   └── [lang]/[service]/[city]/[area]/
│   │   │       └── +page.svelte  # SEO landing pages
│   │   │
│   │   ├── lib/
│   │   │   ├── api/              # Generated API client (from OpenAPI)
│   │   │   ├── components/       # UI components
│   │   │   ├── stores/           # Svelte stores (state)
│   │   │   ├── i18n/             # Paraglide messages
│   │   │   └── utils/
│   │   │
│   │   ├── service-worker.ts     # PWA offline support
│   │   └── app.html
│   │
│   ├── static/
│   ├── svelte.config.js
│   ├── tailwind.config.js
│   └── package.json
│
├── mobile/                       # Flutter apps
│   ├── packages/
│   │   ├── core/                 # Shared business logic
│   │   │   ├── lib/
│   │   │   │   ├── api/          # Generated API client (Dart, from OpenAPI)
│   │   │   │   ├── models/
│   │   │   │   ├── repositories/
│   │   │   │   ├── services/
│   │   │   │   └── offline/      # Offline sync engine (Hive/Isar)
│   │   │   └── pubspec.yaml
│   │   │
│   │   └── ui_kit/               # Shared widgets + theme
│   │       ├── lib/
│   │       │   ├── theme/
│   │       │   ├── widgets/
│   │       │   └── l10n/         # Flutter localization
│   │       └── pubspec.yaml
│   │
│   ├── apps/
│   │   ├── customer/             # Customer app (~8MB)
│   │   └── provider/             # Provider app (~5MB)
│   │
│   └── melos.yaml                # Flutter monorepo tool
│
├── backend/                      # Go API server
│   ├── cmd/
│   │   └── api/
│   │       └── main.go           # Entry point
│   │
│   ├── internal/
│   │   ├── config/               # Jurisdiction configs, env vars
│   │   │
│   │   ├── domain/               # Core business types (no dependencies)
│   │   │   ├── user.go
│   │   │   ├── job.go
│   │   │   ├── provider.go
│   │   │   ├── route.go
│   │   │   ├── payment.go
│   │   │   ├── review.go
│   │   │   ├── dispute.go
│   │   │   └── gamification.go
│   │   │
│   │   ├── service/              # Business logic
│   │   │   ├── matching/         # Matching algorithm
│   │   │   ├── routing/          # Route optimization
│   │   │   ├── payment/          # Payment + escrow logic
│   │   │   ├── job/              # Job lifecycle
│   │   │   ├── review/           # Rating + trust score
│   │   │   ├── dispute/          # Dispute resolution
│   │   │   ├── gamification/     # Points + levels
│   │   │   └── notification/     # Notification orchestration
│   │   │
│   │   ├── adapter/              # External integrations (pluggable)
│   │   │   ├── payment/
│   │   │   │   ├── razorpay.go
│   │   │   │   ├── stripe.go
│   │   │   │   └── cash.go
│   │   │   ├── sms/
│   │   │   │   ├── twilio.go
│   │   │   │   ├── msg91.go
│   │   │   │   └── africas_talking.go
│   │   │   ├── identity/
│   │   │   │   ├── aadhaar.go
│   │   │   │   └── govuk.go
│   │   │   ├── ai/
│   │   │   │   ├── claude.go          # LLM (conversation, content)
│   │   │   │   ├── deepgram.go        # Speech-to-text
│   │   │   │   ├── google_tts.go      # Text-to-speech
│   │   │   │   ├── google_vision.go   # OCR, document processing
│   │   │   │   └── google_translate.go# Translation
│   │   │   └── storage/
│   │   │       └── r2.go
│   │   │
│   │   ├── handler/              # HTTP handlers (API layer)
│   │   │   ├── user.go
│   │   │   ├── job.go
│   │   │   ├── provider.go
│   │   │   ├── search.go
│   │   │   ├── payment.go
│   │   │   ├── ai.go             # AI endpoints (chat, photo, voice)
│   │   │   └── admin.go
│   │   │
│   │   ├── repository/           # Database access (sqlc generated)
│   │   │   ├── postgres/
│   │   │   └── redis/
│   │   │
│   │   ├── middleware/           # Auth, rate limiting, logging
│   │   │
│   │   ├── worker/               # Background job processors (Asynq)
│   │   │   ├── sms_worker.go
│   │   │   ├── payout_worker.go
│   │   │   ├── reminder_worker.go
│   │   │   ├── score_worker.go
│   │   │   ├── ai_content_worker.go   # SEO page generation
│   │   │   ├── ai_index_worker.go     # Meilisearch indexing
│   │   │   └── ml_retrain_worker.go   # ML model retraining trigger
│   │   │
│   │   └── ml/                   # ML model serving
│   │       ├── matching_model.go      # ONNX runtime inference
│   │       ├── fraud_model.go
│   │       └── pricing_model.go
│   │
│   ├── pkg/                      # Shared utilities
│   │   ├── geo/                  # PostGIS helpers
│   │   ├── i18n/                 # Translation helpers
│   │   ├── validator/            # Input validation
│   │   └── logger/               # Structured logging (zerolog)
│   │
│   ├── queries/                  # sqlc SQL files
│   ├── migrations/               # SQL migrations (golang-migrate)
│   ├── sqlc.yaml
│   ├── Dockerfile
│   └── go.mod
│
├── ai/                           # AI/ML pipeline
│   ├── models/
│   │   ├── matching/             # Match quality prediction
│   │   │   ├── train.py
│   │   │   ├── evaluate.py
│   │   │   └── export_onnx.py
│   │   ├── fraud/                # Fraud detection
│   │   │   ├── train.py
│   │   │   └── export_onnx.py
│   │   └── pricing/              # Price prediction
│   │       ├── train.py
│   │       └── export_onnx.py
│   │
│   ├── pipelines/
│   │   ├── retrain_matching.py   # Weekly retrain pipeline
│   │   ├── retrain_fraud.py
│   │   └── generate_seo.py       # Batch SEO content generation
│   │
│   ├── prompts/                  # LLM prompt templates
│   │   ├── booking_agent.md      # Conversational booking prompt
│   │   ├── photo_analysis.md     # Photo-to-job prompt
│   │   ├── crop_diagnosis.md     # Crop disease identification prompt
│   │   ├── dispute_assist.md     # Dispute resolution suggestion prompt
│   │   ├── seo_content.md        # SEO page generation prompt
│   │   └── voice_agent.md        # Voice IVR agent prompt
│   │
│   ├── requirements.txt
│   └── Dockerfile
│
├── api/                          # API specification (source of truth)
│   └── openapi.yaml              # OpenAPI 3.1 spec
│                                 # → generates SvelteKit client (TypeScript)
│                                 # → generates Flutter client (Dart)
│                                 # → generates Go handler stubs
│
├── config/                       # Jurisdiction + category configs
│   ├── jurisdictions/
│   │   ├── india.yaml
│   │   ├── uk.yaml
│   │   └── uae.yaml
│   ├── categories/
│   │   ├── home_repair.yaml
│   │   ├── cleaning.yaml
│   │   └── crops/
│   │       ├── kerala.yaml
│   │       ├── tamil_nadu.yaml
│   │       └── maharashtra.yaml
│   ├── sms_templates/
│   │   ├── en/
│   │   ├── hi/
│   │   ├── ml/
│   │   └── ar/
│   └── seasonal_calendars/
│       ├── kerala_coconut.yaml
│       ├── kerala_rubber.yaml
│       └── maharashtra_sugarcane.yaml
│
├── infrastructure/
│   ├── terraform/                # Infrastructure as Code
│   ├── docker/
│   │   └── docker-compose.yml    # Local dev (PG, Redis, Meilisearch)
│   └── k8s/                      # Kubernetes manifests
│
├── scripts/
│   ├── generate-api-clients.sh   # OpenAPI → TS + Dart + Go
│   ├── seed-db.sh
│   └── migrate.sh
│
└── docs/
```

### API contract — single source of truth

```
api/openapi.yaml is the SINGLE SOURCE OF TRUTH.

From this one file, generate:

  ┌──────────────────┐
  │  openapi.yaml    │
  └────────┬─────────┘
           │
     ┌─────┼──────────────────┐
     │     │                  │
     ▼     ▼                  ▼
  ┌──────┐ ┌──────────┐ ┌─────────┐
  │  Go  │ │SvelteKit │ │ Flutter │
  │server│ │TS client │ │Dart     │
  │stubs │ │(openapi- │ │client   │
  │(oapi-│ │ fetch)   │ │(openapi │
  │ code-│ │          │ │generator│
  │ gen) │ │          │ │ dart)   │
  └──────┘ └──────────┘ └─────────┘

Change the API spec → regenerate all clients.
Backend and frontends never go out of sync.
No manual API client code. No type mismatches.
```

### Go backend — libraries and tools

```
WEB FRAMEWORK:     Fiber (Express-like, fastest Go framework)
                   or Echo (slightly more mature, same performance)

DATABASE:
  ├── sqlc           — generate type-safe Go code from SQL
  │                    (write SQL, get Go functions — no ORM magic)
  ├── pgx            — PostgreSQL driver (fastest, PostGIS support)
  ├── golang-migrate — database migrations
  └── go-redis       — Redis client

SEARCH:            Meilisearch Go client

QUEUE:             Asynq (Redis-based job queue)
  ├── SMS dispatch
  ├── Payment settlement
  ├── Score recalculation
  ├── Seasonal reminders
  ├── AI content generation (SEO pages)
  ├── ML model retraining trigger
  └── Retry with exponential backoff

AUTH:              Custom JWT + OTP
  ├── golang-jwt    — token creation/validation
  └── Redis         — OTP storage, session management

VALIDATION:        go-playground/validator — struct tag validation

OPENAPI:           oapi-codegen — generate handlers from OpenAPI spec

GEO:               PostGIS queries via sqlc (PostgreSQL does the work)

LOGGING:           zerolog (zero-allocation, structured JSON)

CONFIG:            Viper (config files + env vars + hot reload)

ML SERVING:        ONNX Runtime Go bindings
  ├── Matching model (predict match quality)
  ├── Fraud detection model
  └── Pricing model

TESTING:           stdlib testing + testify + testcontainers-go
```

### Database architecture

```
SCHEMA ORGANIZATION (PostgreSQL schemas):

  public          — shared reference data
  ├── jurisdictions
  ├── categories
  ├── crops
  └── languages

  users           — user domain
  ├── users
  ├── user_profiles
  ├── provider_profiles
  ├── company_profiles
  ├── verification_documents
  └── trust_scores

  jobs            — job domain
  ├── jobs
  ├── job_quotes
  ├── job_schedules
  ├── recurring_jobs
  └── job_history

  routes          — route domain
  ├── routes
  ├── route_stops
  ├── route_requests
  └── route_schedules

  payments        — payment domain
  ├── transactions
  ├── escrow_holds
  ├── payouts
  ├── wallets
  ├── invoices
  └── refunds

  reviews         — review domain
  ├── reviews
  ├── review_responses
  └── moderation_queue

  disputes        — dispute domain
  ├── disputes
  ├── dispute_evidence
  ├── dispute_messages
  └── resolutions

  gamification    — gamification domain
  ├── point_ledger
  ├── levels
  ├── achievements
  └── leaderboards

  communication   — messaging domain
  ├── conversations
  ├── messages
  ├── sms_log
  └── notification_log

  analytics       — analytics domain
  ├── events
  ├── daily_metrics
  └── provider_stats

MULTI-JURISDICTION DATA STRATEGY:

  Phase 1: Shared database with jurisdiction_id column
  + row-level security policies for data isolation.

  Phase 2: Database per jurisdiction when regulations
  require it (EU data residency, etc.). Jurisdiction
  routing at API gateway level.

MIGRATIONS:
  ├── golang-migrate for schema changes
  ├── All migrations version-controlled
  └── Zero-downtime (expand-contract pattern):
      1. Add new column (nullable)
      2. Deploy code writing to both old and new
      3. Backfill existing data
      4. Deploy code reading from new only
      5. Drop old column
```

### Security architecture

```
APPLICATION SECURITY:
  ├── Input validation          — Go struct validation tags +
  │                               Zod schemas on SvelteKit frontend
  ├── SQL injection             — sqlc parameterized queries
  │                               (never raw SQL with user input)
  ├── XSS                       — Svelte auto-escapes, CSP headers
  ├── CSRF                      — SameSite cookies, CSRF tokens
  ├── Rate limiting             — per-IP, per-user, per-endpoint
  │                               (Redis-backed, configurable)
  ├── CORS                      — strict origin whitelist
  └── DDoS protection           — Cloudflare WAF

DATA SECURITY:
  ├── Encryption at rest        — AES-256 (RDS, R2)
  ├── Encryption in transit     — TLS 1.3 everywhere
  ├── KYC documents             — encrypted with per-user keys,
  │                               access-logged, auto-deleted
  │                               after verification
  ├── PII handling              — masked in logs (zerolog redaction),
  │                               anonymized in analytics
  ├── Payment data              — never stored, tokenized
  │                               via payment gateway
  └── Secrets management        — AWS Secrets Manager / SSM

ACCESS CONTROL (RBAC):
  ├── Roles: customer, provider, company_admin, ops_agent,
  │          mediator, finance, jurisdiction_admin, super_admin
  ├── Resource-level: provider sees only their jobs,
  │   customer sees only their bookings, mediator sees
  │   disputes in their jurisdiction only
  └── Implemented via Go middleware + database policies

API SECURITY:
  ├── API keys for third-party integrations
  ├── OAuth 2.0 for partner APIs
  ├── Webhook signature verification (HMAC)
  ├── Request signing for payment callbacks
  └── API versioning (URL-based: /api/v1/, /api/v2/)
```

### Feature flags and gradual rollout

```
Tool: Unleash (open-source) or custom Redis-based flags

Flag types:
  ├── Per jurisdiction    — enable crop services only in
  │                         rural jurisdictions
  ├── Per user segment    — beta features for Pro subscribers
  ├── Percentage rollout  — 10% of users get new matching
  │                         algorithm, measure impact
  ├── Kill switch         — disable a payment gateway
  │                         instantly if issues arise
  └── Time-based          — seasonal campaign features

Examples:
  ├── "route_management"       → enabled: [kerala, karnataka]
  ├── "new_matching_v2"        → enabled: 20% of users
  ├── "crew_booking"           → enabled: false (in development)
  └── "ai_voice_ivr"          → enabled: [india]
```

### Scaling strategy

```
Stage 1: Single region, < 10K users
  ├── Single PostgreSQL instance (managed)
  ├── Single Redis instance
  ├── 2-3 Go containers (Fly.io or ECS Fargate)
  ├── Vercel/Cloudflare Pages for SvelteKit
  ├── Meilisearch (single instance)
  └── Total infra cost: ~$50-150/month
      (Go's efficiency makes this dramatically cheaper)

Stage 2: Growing, 10K-100K users
  ├── PostgreSQL read replicas
  ├── Redis cluster
  ├── Auto-scaling Go containers (4-8)
  ├── Meilisearch cluster (3 nodes)
  ├── CDN for all static assets
  └── Total infra cost: ~$300-1,000/month

Stage 3: Multi-jurisdiction, 100K-1M users
  ├── Database per jurisdiction (where required)
  ├── Multi-region deployment
  ├── Kafka for event streaming
  ├── Dedicated analytics pipeline
  └── Total infra cost: ~$2,000-8,000/month

Stage 4: Scale, 1M+ users
  ├── Full Kubernetes (EKS/GKE)
  ├── Global CDN with edge compute
  ├── Data warehouse (BigQuery/ClickHouse)
  ├── ML pipeline for matching optimization
  └── Total infra cost: architecture-dependent
```

---

## 25. AI Integration

AI is embedded at every layer of the platform — not as a separate feature, but as the intelligence that makes every interaction smarter.

### 25.1 Conversational booking (WhatsApp / SMS / Voice)

The biggest unlock. Most users are more comfortable talking or texting naturally than navigating an app.

```
WHATSAPP BOOKING FLOW:

  Customer: "I need someone to pluck coconuts,
            I have about 20 trees in Thrissur"

  AI: "I found 3 coconut climbers near Thrissur.

       1. Rajan — ₹40/tree, 4.8⭐, available this week
       2. Suresh — ₹35/tree, 4.5⭐, available next week
       3. Mohan — ₹45/tree, 4.9⭐, available tomorrow

       Which one? Or should I book the earliest?"

  Customer: "Book Mohan for tomorrow"

  AI: "Done. Mohan will visit tomorrow morning.
       He'll call 30 min before arriving.
       Est. cost: ₹900 (20 trees × ₹45).
       Reply CANCEL to cancel anytime."

  No app download. No navigation. No forms.

SMS BOOKING (basic phone):

  Customer: "Need plumber tap leaking"

  AI parses: category=plumbing, sub=tap/leak,
             urgency=medium-high,
             location=from registered phone

  AI reply: "Tap repair. 2 plumbers near you:
   1. Kumar ₹300 (30min away) ⭐4.7
   2. Prasad ₹250 (1hr away) ⭐4.3
   Reply 1 or 2 to book"
```

**Architecture:**
```
┌──────────────────┐    ┌──────────────────┐
│ WhatsApp Business│    │    SMS Gateway   │
│ API (Meta)       │    │  (Twilio/MSG91)  │
└────────┬─────────┘    └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────▼───────┐
              │  AI Agent    │
              │  Service     │
              │              │
              │  ├── Intent recognition
              │  ├── Entity extraction
              │  ├── Conversation memory (Redis)
              │  ├── Language detection + auto-response
              │  └── Action execution (calls Go API)
              │      ├── searchProviders()
              │      ├── createBooking()
              │      ├── cancelBooking()
              │      ├── getJobStatus()
              │      └── submitComplaint()
              └──────┬───────┘
                     │
              ┌──────▼───────┐
              │  Go API      │
              │  (same APIs  │
              │   web/mobile │
              │   apps use)  │
              └──────────────┘

LLM: Claude API (tool use / function calling)
     AI doesn't just chat — it calls real APIs
     to book, cancel, search.

Fallback: if AI can't resolve → human handoff
          with full conversation context.
```

### 25.2 Photo-to-job

Customers don't know terminology. They don't know if it's a "ball valve" or "gate valve." But they can take a photo.

```
Customer uploads photo of leaking pipe:

  Vision AI detects:
  ├── Leaking pipe joint
  ├── Under-sink plumbing, kitchen
  ├── Severity: moderate (drip, not burst)

  AI generates:
  "Kitchen sink pipe joint leak. Moderate —
   active drip, not emergency. Likely needs
   joint replacement or re-sealing."

  Category: Plumbing > Pipe leak
  Urgency: Medium
  Est. cost: ₹300-600
  Est. time: 30-45 min

  [Confirm and find plumber]

CROP VERSION:

  Photo of coconut tree with yellowing leaves:

  AI: "This appears to be bud rot (Phytophthora).
       Early stage — treatable.

       Recommended: Coconut tree treatment specialist
       2 specialists near you:
       1. Vijayan — 4.8⭐, disease specialist
       2. AgriCare — 4.6⭐, certified

       Book treatment?"
```

**Tech:** Claude Vision API for analysis. Google Cloud Vision for OCR (KYC documents).

### 25.3 Voice AI for IVR

Replaces rigid "Press 1 for..." menu trees with natural conversation.

```
Provider calls platform number:

  AI: "Hi Rajan, you have 3 jobs today. What do you need?"
  Rajan: "I can't make it to the 2pm job"
  AI: "Your 2pm is tap repair for Priya in Ollur.
       Reschedule or find replacement?"
  Rajan: "Find someone else, I'm sick"
  AI: "Done. I'll reassign your 2pm and 4pm jobs
       and mark you offline today. Feel better."

Customer calls (in Malayalam):

  AI: "Hi, how can I help?"
  Customer: "എന്റെ വീട്ടിലെ ടാപ്പ് ലീക്ക് ചെയ്യുന്നു"
  AI (Malayalam): "ഞാൻ ഒരു പ്ലംബറെ കണ്ടെത്താം.
                   ഇന്ന് വേണോ നാളെ വേണോ?"

TECH:
  ├── STT: Deepgram (best for Indian languages,
  │   lowest latency, streaming support)
  ├── LLM: Claude API for understanding + response
  ├── TTS: Google Cloud TTS (multi-language, natural,
  │   supports Malayalam, Hindi, Arabic, etc.)
  ├── Telephony: Twilio or Exotel
  └── Latency target: <2 seconds
      (streaming STT + LLM + TTS)
```

### 25.4 Smart matching (ML-enhanced)

Starts rule-based, improves with AI as data accumulates.

```
RULE-BASED (day 1):
  Score = distance × 0.25 + trust × 0.25 +
          response_time × 0.15 + price × 0.15 +
          completion_rate × 0.10 + language × 0.05 +
          loyalty × 0.05

ML-ENHANCED (after data):

  Training data:
  Job posted → providers shown → provider selected
  → job completed → rating given

  Model learns:
  ├── "Kochi customers prioritize speed over price"
  ├── "For AC repair, certification > distance"
  ├── "For coconut plucking, route compatibility > rating"
  ├── "Repeat customers prefer previous provider
  │    even if slightly more expensive"
  └── "Evening jobs need providers with own transport"

  Weights become learned, not manually set.
  Different weights per category, jurisdiction,
  customer segment — automatically.

TECH:
  ├── Initial: rule-based scoring in Go
  ├── Phase 2: XGBoost trained on match outcomes
  ├── Serve: ONNX model loaded in Go (ONNX Runtime)
  └── Retrain: weekly batch job (Python)
```

### 25.5 Real-time translation

Provider speaks Malayalam. Customer speaks Hindi. Platform bridges.

```
IN-APP CHAT:

  Customer (Hindi): "कब तक आएंगे?"
  Provider sees: "എപ്പോൾ വരും?" (Malayalam)

  Provider (Malayalam): "അരമണിക്കൂറിൽ എത്തും"
  Customer sees: "आधे घंटे में पहुँचूँगा" (Hindi)

SMS RELAY WITH TRANSLATION:

  Customer chat (Hindi)
  → translate to Malayalam
  → send as SMS to basic phone provider
  → provider replies SMS in Malayalam
  → translate to Hindi
  → customer sees Hindi in app

TECH:
  ├── Google Cloud Translation API (all Indian languages)
  ├── LLM fallback for nuanced translations
  ├── Language detection: automatic
  └── Cache common phrases ("I'm on my way",
      "Job is done") — no API call needed
```

### 25.6 Fraud detection

```
FAKE REVIEW DETECTION:
  ├── Review posted within seconds of completion
  ├── Reviewer has no other activity
  ├── Multiple 5-star reviews from related accounts
  ├── Graph analysis: cluster of providers rating each other
  └── Model: anomaly detection + graph neural network

FAKE PROVIDER DETECTION:
  ├── Same device fingerprint, multiple accounts
  ├── Accepts jobs but never completes
  ├── KYC document appears edited (AI image analysis)
  └── Location spoofing (GPS inconsistencies)

PRICE MANIPULATION:
  ├── Quotes significantly above market rate
  ├── Charges differ for similar jobs
  └── Model: outlier detection vs category/location baseline

TECH:
  ├── Feature engineering: Go backend
  ├── Training: Python (scikit-learn / PyTorch)
  ├── Serving: ONNX model in Go
  └── Retraining: weekly
```

### 25.7 Price intelligence

```
CUSTOMER-FACING:

  "Fair price range for tap repair in Koramangala:

   ₹200 ──────[████████]────── ₹800
              ₹300-500 (most common)

   Based on 847 similar jobs in last 6 months"

PROVIDER-FACING:

  "Your rate (₹600) is above typical range (₹300-500).
   Providers priced ₹300-400 get 3x more bookings."

DYNAMIC INSIGHTS:

  ├── "AC demand 3x higher — consider raising rate 10-15%"
  ├── "Only 2 coconut climbers in your postcode.
  │    You could serve 15 more by extending route 5km."
  └── "Painters charge ₹15/sqft, you charge ₹12.
       You could increase without losing bookings."

TECH:
  ├── Aggregate job prices by category × postcode × period
  ├── Stats initially, ML regression for predictions
  └── Nightly batch → cached in Redis
```

### 25.8 Proactive maintenance intelligence

```
AI LEARNS FROM JOB HISTORY:

  Customer's AC last serviced 5 months ago + summer approaching:

  "Your AC was last serviced 5 months ago. Summer demand
   increases 4x in March-April. Book now at current rates?
   Your technician Rajesh has slots next week."

CROP INTELLIGENCE:

  "Pepper harvest starts in 3 weeks. Last year you booked
   3 workers. 60% of slots already booked — book now?"

  "Heavy rain forecast next week. Your rubber tapping
   may be interrupted. Tapper Suresh has been notified."

TECH:
  ├── Seasonal calendar config (already in system)
  ├── Weather API (OpenWeatherMap)
  ├── Job history analysis per customer
  └── Scheduled notifications via Asynq
```

### 25.9 AI-generated SEO content

```
For every /[service]/[city]/[area] landing page:

  AI generates data-driven, unique content:

  "Looking for a plumber in Ollur, Thrissur?
   12 verified plumbers. Average response: 25 min.
   Common: tap repair (₹300-500), pipe leak (₹400-800),
   bathroom fitting (₹2,000-5,000)..."

  ├── Not generic templates — real data per page
  ├── Generated in local language + English
  ├── Regenerated monthly with fresh data
  └── Batch job: nightly via Asynq + Claude API
```

### 25.10 AI admin and ops tools

```
DISPUTE RESOLUTION ASSIST:
  ├── AI reads evidence (messages, photos)
  ├── Suggests resolution from similar past disputes
  ├── "85% of similar cases → partial refund 30-50%"
  ├── Mediator reviews, approves/modifies
  └── Hours → minutes

KYC VERIFICATION:
  ├── Document type detection (Aadhaar, passport, etc.)
  ├── OCR: extract name, number, address
  ├── Face match: selfie vs document photo
  ├── Tampering detection
  └── Auto-approve clear cases, flag edge cases

ANOMALY ALERTS:
  ├── "Bookings in Kochi dropped 40% — SMS gateway down?"
  ├── "Provider #4521: 3 disputes in 24h — auto-suspended"
  ├── "Payment failures spiked 5x — Razorpay outage?"
  └── Pattern-based, not threshold-based
```

### AI tech stack

```
LLM (conversation, content, analysis):
  ├── Claude API — primary (best reasoning, tool use)
  ├── Fine-tuned small model — high-volume simple
  │   conversations (cost reduction)
  └── Fallback: rule-based intent matching for
      ultra-simple SMS flows

VISION AI:
  ├── Claude Vision — photo-to-job, crop diagnosis
  └── Google Cloud Vision — OCR, document processing

SPEECH:
  ├── STT: Deepgram (Indian languages, low latency)
  ├── TTS: Google Cloud TTS (multi-language)
  └── Alt STT: Whisper (self-hosted, no per-min cost)

TRANSLATION:
  ├── Google Cloud Translation (fast, cheap, accurate)
  └── LLM fallback for nuanced cases

ML MODELS (matching, fraud, pricing):
  ├── Training: Python (scikit-learn, XGBoost, PyTorch)
  ├── Export: ONNX format
  ├── Serving: ONNX Runtime in Go (no Python service)
  └── Retraining: weekly batch job

COST MANAGEMENT:
  ├── Cache common LLM responses (Redis)
  ├── Small models for simple tasks (intent classification)
  ├── Batch non-urgent tasks (SEO, analytics → nightly)
  └── Pre-compute where possible (prices, fraud scores
      → computed hourly, served from cache)
```

### What AI does NOT replace

```
AI DOES:
  ├── First-line conversation (booking, status, FAQs)
  ├── Photo understanding
  ├── Language bridging
  ├── Pattern detection (fraud, pricing, demand)
  ├── Content generation
  └── Decision support (dispute suggestions, matching)

AI DOES NOT:
  ├── Replace human mediators for serious disputes
  │   AI suggests, human decides
  ├── Make final KYC decisions on edge cases
  ├── Handle safety emergencies (SOS → human)
  ├── Set pricing (AI advises, provider decides)
  └── Override user decisions (AI recommends, never forces)

PRINCIPLE: AI augments every interaction but a human
is always reachable. "Talk to a person" → they get one.
```
