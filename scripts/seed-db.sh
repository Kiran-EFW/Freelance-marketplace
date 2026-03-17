#!/usr/bin/env bash
set -euo pipefail

echo "=== Seeding Database ==="

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load env
if [ -f "$PROJECT_ROOT/backend/.env" ]; then
    set -a
    source "$PROJECT_ROOT/backend/.env"
    set +a
fi

DATABASE_URL="${DATABASE_URL:-postgres://marketplace:marketplace@localhost:5432/marketplace?sslmode=disable}"

psql "$DATABASE_URL" <<'SQL'

-- Seed top-level categories
INSERT INTO categories (slug, name, icon, sort_order, requires_license, pricing_model) VALUES
('home-services',        '{"en": "Home Services", "hi": "घरेलू सेवाएं", "ml": "ഗൃഹ സേവനങ്ങൾ"}',       'home',         1, false, 'fixed'),
('professional-services','{"en": "Professional Services", "hi": "पेशेवर सेवाएं"}',                      'briefcase',    2, true,  'hourly'),
('crop-and-land',        '{"en": "Crop & Land Services", "hi": "फसल और भूमि सेवाएं", "ml": "കൃഷി സേവനങ്ങൾ"}', 'leaf', 3, false, 'per_day'),
('automotive',           '{"en": "Automotive", "hi": "ऑटोमोटिव"}',                                      'car',          4, false, 'fixed'),
('health-wellness',      '{"en": "Health & Wellness", "hi": "स्वास्थ्य और कल्याण"}',                     'heart',        5, true,  'hourly'),
('education-tutoring',   '{"en": "Education & Tutoring", "hi": "शिक्षा और ट्यूशन"}',                     'book',         6, false, 'hourly'),
('events',               '{"en": "Events & Occasions", "hi": "इवेंट्स"}',                                'calendar',     7, false, 'fixed'),
('pet-services',         '{"en": "Pet Services", "hi": "पालतू जानवर सेवाएं"}',                            'paw',          8, false, 'fixed'),
('tech-services',        '{"en": "Tech Services", "hi": "तकनीकी सेवाएं"}',                               'monitor',      9, false, 'fixed'),
('logistics-moving',     '{"en": "Logistics & Moving", "hi": "लॉजिस्टिक्स"}',                            'truck',       10, false, 'fixed'),
('beauty-personal-care', '{"en": "Beauty & Personal Care", "hi": "सौंदर्य और देखभाल"}',                   'scissors',    11, false, 'fixed'),
('cleaning-maintenance', '{"en": "Cleaning & Maintenance", "hi": "सफाई और रखरखाव"}',                     'sparkles',    12, false, 'per_sqft')
ON CONFLICT (slug) DO NOTHING;

-- Seed subcategories for Home Services
INSERT INTO categories (slug, name, parent_id, icon, sort_order, requires_license, pricing_model) VALUES
('plumbing',        '{"en": "Plumbing"}',        (SELECT id FROM categories WHERE slug = 'home-services'), 'wrench',      1, false, 'fixed'),
('electrical',      '{"en": "Electrical"}',      (SELECT id FROM categories WHERE slug = 'home-services'), 'zap',         2, true,  'fixed'),
('painting',        '{"en": "Painting"}',        (SELECT id FROM categories WHERE slug = 'home-services'), 'paintbrush',  3, false, 'per_sqft'),
('carpentry',       '{"en": "Carpentry"}',       (SELECT id FROM categories WHERE slug = 'home-services'), 'hammer',      4, false, 'fixed'),
('cleaning',        '{"en": "House Cleaning"}',  (SELECT id FROM categories WHERE slug = 'home-services'), 'sparkles',    5, false, 'per_sqft'),
('pest-control',    '{"en": "Pest Control"}',    (SELECT id FROM categories WHERE slug = 'home-services'), 'bug',         6, true,  'fixed'),
('appliance-repair','{"en": "Appliance Repair"}', (SELECT id FROM categories WHERE slug = 'home-services'), 'settings',   7, false, 'fixed'),
('roofing',         '{"en": "Roofing"}',         (SELECT id FROM categories WHERE slug = 'home-services'), 'umbrella',    8, false, 'per_sqft'),
('flooring',        '{"en": "Flooring"}',        (SELECT id FROM categories WHERE slug = 'home-services'), 'layout',      9, false, 'per_sqft'),
('hvac',            '{"en": "HVAC"}',            (SELECT id FROM categories WHERE slug = 'home-services'), 'thermometer',10, true,  'fixed'),
('locksmith',       '{"en": "Locksmith"}',       (SELECT id FROM categories WHERE slug = 'home-services'), 'lock',       11, false, 'fixed'),
('interior-design', '{"en": "Interior Design"}', (SELECT id FROM categories WHERE slug = 'home-services'), 'palette',    12, false, 'fixed')
ON CONFLICT (slug) DO NOTHING;

-- Seed subcategories for Crop & Land Services
INSERT INTO categories (slug, name, parent_id, icon, sort_order, requires_license, pricing_model) VALUES
('tree-work',         '{"en": "Tree Work", "ml": "മരപ്പണി"}',         (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'tree-pine',   1, false, 'per_tree'),
('field-preparation', '{"en": "Field Preparation"}',                  (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'shovel',      2, false, 'per_day'),
('harvesting',        '{"en": "Harvesting"}',                         (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'wheat',       3, false, 'per_day'),
('irrigation',        '{"en": "Irrigation"}',                         (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'droplets',    4, false, 'fixed'),
('pest-management',   '{"en": "Pest Management"}',                   (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'shield',      5, true,  'per_unit'),
('landscaping',       '{"en": "Landscaping"}',                        (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'trees',       6, false, 'per_sqft'),
('soil-testing',      '{"en": "Soil Testing"}',                       (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'flask',       7, false, 'fixed'),
('equipment-rental',  '{"en": "Equipment Rental"}',                   (SELECT id FROM categories WHERE slug = 'crop-and-land'), 'tractor',     8, false, 'per_day')
ON CONFLICT (slug) DO NOTHING;

-- Seed Kerala crop catalog
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar) VALUES
('in', 'coconut', '{"en": "Coconut", "ml": "തെങ്ങ്"}',
 '[{"slug": "climbing-harvest", "name": "Climbing & Harvest", "pricing_model": "per_tree", "typical_range": "50-150"},
   {"slug": "de-husking", "name": "De-husking", "pricing_model": "per_unit", "typical_range": "2-5"},
   {"slug": "crown-cleaning", "name": "Crown Cleaning", "pricing_model": "per_tree", "typical_range": "30-100"},
   {"slug": "pest-treatment", "name": "Pest Treatment", "pricing_model": "per_tree", "typical_range": "100-300"}]',
 '{"1": ["climbing-harvest", "crown-cleaning"], "2": ["climbing-harvest"], "3": ["climbing-harvest", "pest-treatment"], "4": ["climbing-harvest"], "5": ["climbing-harvest", "crown-cleaning"], "6": ["climbing-harvest", "pest-treatment"], "7": ["climbing-harvest"], "8": ["climbing-harvest"], "9": ["climbing-harvest", "crown-cleaning"], "10": ["climbing-harvest", "pest-treatment"], "11": ["climbing-harvest"], "12": ["climbing-harvest"]}'
),
('in', 'rubber', '{"en": "Rubber", "ml": "റബ്ബർ"}',
 '[{"slug": "tapping", "name": "Rubber Tapping", "pricing_model": "per_day", "typical_range": "500-800"},
   {"slug": "rain-guarding", "name": "Rain Guarding", "pricing_model": "per_day", "typical_range": "400-600"},
   {"slug": "sheet-making", "name": "Sheet Making", "pricing_model": "per_unit", "typical_range": "5-10"}]',
 '{"1": ["tapping", "sheet-making"], "2": ["tapping", "sheet-making"], "3": ["tapping", "sheet-making"], "4": ["tapping", "sheet-making"], "5": ["tapping", "sheet-making"], "6": ["rain-guarding"], "7": ["rain-guarding"], "8": ["rain-guarding"], "9": ["tapping", "sheet-making"], "10": ["tapping", "sheet-making"], "11": ["tapping", "sheet-making"], "12": ["tapping", "sheet-making"]}'
),
('in', 'pepper', '{"en": "Pepper", "ml": "കുരുമുളക്"}',
 '[{"slug": "vine-training", "name": "Vine Training", "pricing_model": "per_day", "typical_range": "600-900"},
   {"slug": "harvesting", "name": "Harvesting", "pricing_model": "per_day", "typical_range": "700-1000"},
   {"slug": "drying", "name": "Drying & Processing", "pricing_model": "per_day", "typical_range": "500-700"}]',
 '{"1": ["harvesting", "drying"], "2": ["harvesting", "drying"], "3": ["vine-training"], "4": ["vine-training"], "5": ["vine-training"], "6": [], "7": [], "8": [], "9": [], "10": [], "11": [], "12": ["harvesting"]}'
),
('in', 'rice', '{"en": "Rice (Paddy)", "ml": "നെല്ല്"}',
 '[{"slug": "ploughing", "name": "Ploughing", "pricing_model": "per_day", "typical_range": "1000-2000"},
   {"slug": "transplanting", "name": "Transplanting", "pricing_model": "per_day", "typical_range": "600-900"},
   {"slug": "weeding", "name": "Weeding", "pricing_model": "per_day", "typical_range": "500-700"},
   {"slug": "harvesting", "name": "Harvesting", "pricing_model": "per_day", "typical_range": "800-1200"}]',
 '{"1": [], "2": [], "3": [], "4": ["ploughing"], "5": ["ploughing", "transplanting"], "6": ["transplanting", "weeding"], "7": ["weeding"], "8": ["harvesting"], "9": ["harvesting"], "10": ["ploughing"], "11": ["transplanting", "weeding"], "12": ["weeding"]}'
)
ON CONFLICT (jurisdiction_id, crop_slug) DO NOTHING;

SELECT 'Seeding complete!' as status;
SQL

echo ""
echo "=== Database seeded successfully ==="
