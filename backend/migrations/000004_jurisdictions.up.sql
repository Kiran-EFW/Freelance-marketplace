-- Seed jurisdiction configurations for supported countries.

-- Update India jurisdiction with full config
UPDATE jurisdictions SET config = '{
  "commission_rates": {"low": 0.03, "mid": 0.05, "high": 0.08},
  "commission_thresholds": {"low_max": 10000, "mid_max": 2000},
  "subscription_prices": {"professional": 299, "enterprise": 999},
  "lead_fees": {"small": 15, "medium": 50, "large": 100, "premium": 300},
  "urgent_booking_fee": 49,
  "payment_methods": ["upi", "card", "cash", "wallet"],
  "default_payment_gateway": "razorpay",
  "id_verification": "aadhaar",
  "phone_format": "^\\+91[6-9]\\d{9}$",
  "postcode_format": "^\\d{6}$",
  "supported_languages": ["en", "hi", "ta", "te", "kn", "ml", "bn", "gu", "mr", "pa"],
  "service_categories_enabled": ["home_repair", "cleaning", "beauty_wellness", "professional_services", "vehicle_services", "education_tutoring", "care_services", "events_occasions", "moving_logistics", "tech_digital", "crop_land", "construction"],
  "seasonal_calendar_enabled": true,
  "sms_templates": {
    "otp": "Your Seva verification code is {code}. Valid for 5 minutes.",
    "job_alert": "New {category} job near you! Open Seva to respond.",
    "quote_received": "You received a quote of {currency}{amount} from {provider}."
  }
}'::jsonb WHERE id = 'in';

-- UK jurisdiction
INSERT INTO jurisdictions (id, name, default_language, currency, currency_symbol, phone_prefix, timezone, is_active, config)
VALUES ('uk', 'United Kingdom', 'en', 'GBP', '£', '+44', 'Europe/London', true, '{
  "commission_rates": {"low": 0.03, "mid": 0.05, "high": 0.07},
  "commission_thresholds": {"low_max": 500, "mid_max": 100},
  "subscription_prices": {"professional": 14.99, "enterprise": 49.99},
  "lead_fees": {"small": 1, "medium": 3, "large": 5, "premium": 15},
  "urgent_booking_fee": 4.99,
  "payment_methods": ["card", "bank_transfer"],
  "default_payment_gateway": "stripe",
  "id_verification": "govuk",
  "phone_format": "^\\+44\\d{10}$",
  "postcode_format": "^[A-Z]{1,2}\\d[A-Z\\d]?\\s?\\d[A-Z]{2}$",
  "supported_languages": ["en"],
  "service_categories_enabled": ["home_repair", "cleaning", "beauty_wellness", "professional_services", "vehicle_services", "education_tutoring", "care_services", "events_occasions", "moving_logistics", "tech_digital", "construction"],
  "seasonal_calendar_enabled": false
}'::jsonb)
ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, is_active = true;

-- US jurisdiction
INSERT INTO jurisdictions (id, name, default_language, currency, currency_symbol, phone_prefix, timezone, is_active, config)
VALUES ('us', 'United States', 'en', 'USD', '$', '+1', 'America/New_York', false, '{
  "commission_rates": {"low": 0.03, "mid": 0.05, "high": 0.07},
  "subscription_prices": {"professional": 19.99, "enterprise": 59.99},
  "payment_methods": ["card", "bank_transfer"],
  "default_payment_gateway": "stripe",
  "phone_format": "^\\+1\\d{10}$"
}'::jsonb)
ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;

-- Germany
INSERT INTO jurisdictions (id, name, default_language, currency, currency_symbol, phone_prefix, timezone, is_active, config)
VALUES ('de', 'Germany', 'de', 'EUR', '€', '+49', 'Europe/Berlin', false, '{
  "commission_rates": {"low": 0.03, "mid": 0.05, "high": 0.07},
  "subscription_prices": {"professional": 14.99, "enterprise": 49.99},
  "payment_methods": ["card", "bank_transfer", "sepa"],
  "default_payment_gateway": "stripe",
  "phone_format": "^\\+49\\d{10,11}$",
  "supported_languages": ["de", "en"]
}'::jsonb)
ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;

-- France
INSERT INTO jurisdictions (id, name, default_language, currency, currency_symbol, phone_prefix, timezone, is_active, config)
VALUES ('fr', 'France', 'fr', 'EUR', '€', '+33', 'Europe/Paris', false, '{
  "commission_rates": {"low": 0.03, "mid": 0.05, "high": 0.07},
  "subscription_prices": {"professional": 14.99, "enterprise": 49.99},
  "payment_methods": ["card", "bank_transfer"],
  "default_payment_gateway": "stripe",
  "phone_format": "^\\+33\\d{9}$",
  "supported_languages": ["fr", "en"]
}'::jsonb)
ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;
