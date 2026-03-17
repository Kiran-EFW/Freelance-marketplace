-- Seed crop catalog for India jurisdiction
-- Each entry includes multilingual names, work types with pricing, and
-- a monthly seasonal calendar indicating which work types are relevant.

-- Coconut (Kerala/Karnataka/Tamil Nadu)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'coconut', '{"en": "Coconut", "ml": "തെങ്ങ്", "kn": "ತೆಂಗಿನಕಾಯಿ", "ta": "தேங்காய்", "hi": "नारियल"}'::jsonb,
'[
  {"slug": "tree_climbing", "name": {"en": "Tree Climbing & Harvesting"}, "pricing_model": "per_tree", "typical_price": {"min": 30, "max": 80, "currency": "INR"}},
  {"slug": "tree_pruning", "name": {"en": "Frond Cutting & Pruning"}, "pricing_model": "per_tree", "typical_price": {"min": 20, "max": 50, "currency": "INR"}},
  {"slug": "pest_treatment", "name": {"en": "Pest Treatment"}, "pricing_model": "per_tree", "typical_price": {"min": 50, "max": 150, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer Application"}, "pricing_model": "per_tree", "typical_price": {"min": 30, "max": 80, "currency": "INR"}}
]'::jsonb,
'{"1": ["tree_climbing", "fertilizing"], "2": ["tree_climbing", "pest_treatment"], "3": ["tree_climbing", "tree_pruning"], "4": ["tree_climbing", "fertilizing"], "5": ["tree_climbing"], "6": ["tree_climbing", "pest_treatment"], "7": ["tree_climbing", "tree_pruning"], "8": ["tree_climbing", "fertilizing"], "9": ["tree_climbing"], "10": ["tree_climbing", "pest_treatment"], "11": ["tree_climbing", "fertilizing"], "12": ["tree_climbing", "tree_pruning"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Rice/Paddy (Pan-India)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'rice', '{"en": "Rice/Paddy", "hi": "धान", "ta": "நெல்", "te": "వరి", "kn": "ಭತ್ತ", "bn": "ধান"}'::jsonb,
'[
  {"slug": "ploughing", "name": {"en": "Ploughing & Land Preparation"}, "pricing_model": "per_day", "typical_price": {"min": 1500, "max": 3000, "currency": "INR"}},
  {"slug": "transplanting", "name": {"en": "Transplanting"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 800, "currency": "INR"}},
  {"slug": "harvesting", "name": {"en": "Harvesting & Threshing"}, "pricing_model": "per_day", "typical_price": {"min": 2000, "max": 5000, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pesticide Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}}
]'::jsonb,
'{"1": [], "2": [], "3": [], "4": ["ploughing"], "5": ["ploughing", "transplanting"], "6": ["transplanting", "spraying"], "7": ["spraying"], "8": ["spraying"], "9": ["harvesting"], "10": ["harvesting"], "11": [], "12": []}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Wheat (North India)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'wheat', '{"en": "Wheat", "hi": "गेहूं", "pa": "ਕਣਕ", "gu": "ઘઉં", "mr": "गहू"}'::jsonb,
'[
  {"slug": "ploughing", "name": {"en": "Ploughing & Land Preparation"}, "pricing_model": "per_day", "typical_price": {"min": 1500, "max": 3000, "currency": "INR"}},
  {"slug": "sowing", "name": {"en": "Sowing & Seed Drilling"}, "pricing_model": "per_day", "typical_price": {"min": 1000, "max": 2000, "currency": "INR"}},
  {"slug": "irrigation", "name": {"en": "Irrigation Management"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 1000, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pesticide & Weed Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "harvesting", "name": {"en": "Harvesting & Threshing"}, "pricing_model": "per_day", "typical_price": {"min": 2000, "max": 4000, "currency": "INR"}}
]'::jsonb,
'{"1": ["irrigation", "spraying"], "2": ["irrigation", "spraying"], "3": ["harvesting"], "4": ["harvesting"], "5": [], "6": [], "7": [], "8": [], "9": [], "10": ["ploughing"], "11": ["ploughing", "sowing"], "12": ["sowing", "irrigation"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Sugarcane (Maharashtra/UP/Karnataka)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'sugarcane', '{"en": "Sugarcane", "hi": "गन्ना", "mr": "ऊस", "kn": "ಕಬ್ಬು", "ta": "கரும்பு"}'::jsonb,
'[
  {"slug": "ploughing", "name": {"en": "Land Preparation & Ridging"}, "pricing_model": "per_day", "typical_price": {"min": 1500, "max": 3500, "currency": "INR"}},
  {"slug": "planting", "name": {"en": "Sett Planting"}, "pricing_model": "per_day", "typical_price": {"min": 600, "max": 1200, "currency": "INR"}},
  {"slug": "weeding", "name": {"en": "Weeding & Earthing Up"}, "pricing_model": "per_day", "typical_price": {"min": 400, "max": 800, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pesticide Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "harvesting", "name": {"en": "Harvesting & Loading"}, "pricing_model": "per_day", "typical_price": {"min": 2500, "max": 5000, "currency": "INR"}}
]'::jsonb,
'{"1": ["harvesting"], "2": ["ploughing", "planting"], "3": ["planting", "weeding"], "4": ["weeding", "spraying"], "5": ["weeding", "spraying"], "6": ["spraying"], "7": ["spraying"], "8": ["weeding"], "9": ["weeding"], "10": ["spraying"], "11": ["harvesting"], "12": ["harvesting"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Cotton (Gujarat/Maharashtra/Telangana)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'cotton', '{"en": "Cotton", "hi": "कपास", "gu": "કપાસ", "mr": "कापूस", "te": "పత్తి"}'::jsonb,
'[
  {"slug": "ploughing", "name": {"en": "Land Preparation"}, "pricing_model": "per_day", "typical_price": {"min": 1500, "max": 3000, "currency": "INR"}},
  {"slug": "sowing", "name": {"en": "Sowing"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pesticide & Bollworm Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 1000, "max": 2000, "currency": "INR"}},
  {"slug": "picking", "name": {"en": "Cotton Picking"}, "pricing_model": "per_day", "typical_price": {"min": 400, "max": 700, "currency": "INR"}},
  {"slug": "weeding", "name": {"en": "Weeding & Thinning"}, "pricing_model": "per_day", "typical_price": {"min": 400, "max": 800, "currency": "INR"}}
]'::jsonb,
'{"1": [], "2": [], "3": [], "4": ["ploughing"], "5": ["ploughing", "sowing"], "6": ["sowing", "weeding"], "7": ["weeding", "spraying"], "8": ["spraying"], "9": ["spraying", "picking"], "10": ["picking"], "11": ["picking"], "12": []}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Arecanut/Betel Nut (Karnataka/Kerala)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'arecanut', '{"en": "Arecanut / Betel Nut", "kn": "ಅಡಿಕೆ", "ml": "അടക്ക", "ta": "பாக்கு"}'::jsonb,
'[
  {"slug": "tree_climbing", "name": {"en": "Tree Climbing & Harvesting"}, "pricing_model": "per_tree", "typical_price": {"min": 25, "max": 60, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pest & Disease Spraying"}, "pricing_model": "per_tree", "typical_price": {"min": 40, "max": 100, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer Application"}, "pricing_model": "per_tree", "typical_price": {"min": 20, "max": 50, "currency": "INR"}},
  {"slug": "processing", "name": {"en": "Post-Harvest Processing"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 1000, "currency": "INR"}}
]'::jsonb,
'{"1": ["tree_climbing", "fertilizing"], "2": ["tree_climbing", "spraying"], "3": ["tree_climbing"], "4": ["tree_climbing", "fertilizing"], "5": ["spraying"], "6": ["spraying"], "7": ["fertilizing"], "8": ["tree_climbing"], "9": ["tree_climbing", "processing"], "10": ["tree_climbing", "processing"], "11": ["tree_climbing", "processing"], "12": ["tree_climbing", "fertilizing"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Rubber (Kerala/Tripura)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'rubber', '{"en": "Rubber", "ml": "റബ്ബർ", "hi": "रबर"}'::jsonb,
'[
  {"slug": "tapping", "name": {"en": "Rubber Tapping"}, "pricing_model": "per_day", "typical_price": {"min": 600, "max": 1000, "currency": "INR"}},
  {"slug": "rain_guarding", "name": {"en": "Rain Guard Installation"}, "pricing_model": "per_tree", "typical_price": {"min": 15, "max": 30, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer Application"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 1000, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Disease Control Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}}
]'::jsonb,
'{"1": ["tapping"], "2": ["tapping", "fertilizing"], "3": ["tapping"], "4": ["tapping"], "5": ["tapping", "rain_guarding"], "6": [], "7": [], "8": ["rain_guarding"], "9": ["tapping", "spraying"], "10": ["tapping"], "11": ["tapping", "fertilizing"], "12": ["tapping"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Tea (Assam/West Bengal/Kerala)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'tea', '{"en": "Tea", "as": "চাহ", "bn": "চা", "hi": "चाय"}'::jsonb,
'[
  {"slug": "plucking", "name": {"en": "Tea Leaf Plucking"}, "pricing_model": "per_day", "typical_price": {"min": 300, "max": 500, "currency": "INR"}},
  {"slug": "pruning", "name": {"en": "Bush Pruning & Shaping"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 900, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pest & Weed Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer Application"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 1000, "currency": "INR"}}
]'::jsonb,
'{"1": ["pruning"], "2": ["pruning", "fertilizing"], "3": ["plucking", "spraying"], "4": ["plucking", "spraying"], "5": ["plucking"], "6": ["plucking"], "7": ["plucking", "spraying"], "8": ["plucking"], "9": ["plucking", "spraying"], "10": ["plucking", "fertilizing"], "11": ["plucking"], "12": ["pruning"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Coffee (Karnataka/Kerala/Tamil Nadu)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'coffee', '{"en": "Coffee", "kn": "ಕಾಫಿ", "ml": "കാപ്പി", "ta": "காபி"}'::jsonb,
'[
  {"slug": "picking", "name": {"en": "Cherry Picking"}, "pricing_model": "per_day", "typical_price": {"min": 400, "max": 700, "currency": "INR"}},
  {"slug": "pruning", "name": {"en": "Shade Tree Pruning"}, "pricing_model": "per_day", "typical_price": {"min": 600, "max": 1000, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pest & Borer Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "processing", "name": {"en": "Pulping & Drying"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 900, "currency": "INR"}}
]'::jsonb,
'{"1": ["picking", "processing"], "2": ["processing", "pruning"], "3": ["pruning"], "4": ["spraying"], "5": ["spraying"], "6": [], "7": [], "8": ["spraying"], "9": ["spraying"], "10": ["picking"], "11": ["picking"], "12": ["picking", "processing"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Banana (Tamil Nadu/Maharashtra/Gujarat)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'banana', '{"en": "Banana", "ta": "வாழை", "hi": "केला", "mr": "केळी", "kn": "ಬಾಳೆ"}'::jsonb,
'[
  {"slug": "planting", "name": {"en": "Sucker Planting"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 1000, "currency": "INR"}},
  {"slug": "deleafing", "name": {"en": "De-leafing & Propping"}, "pricing_model": "per_day", "typical_price": {"min": 400, "max": 700, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pest & Disease Spraying"}, "pricing_model": "per_day", "typical_price": {"min": 800, "max": 1500, "currency": "INR"}},
  {"slug": "harvesting", "name": {"en": "Bunch Harvesting"}, "pricing_model": "per_day", "typical_price": {"min": 600, "max": 1200, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer & Manure Application"}, "pricing_model": "per_day", "typical_price": {"min": 500, "max": 900, "currency": "INR"}}
]'::jsonb,
'{"1": ["deleafing", "fertilizing"], "2": ["planting", "fertilizing"], "3": ["planting", "spraying"], "4": ["spraying", "deleafing"], "5": ["spraying", "deleafing"], "6": ["planting", "spraying"], "7": ["spraying", "fertilizing"], "8": ["deleafing", "harvesting"], "9": ["harvesting"], "10": ["harvesting", "fertilizing"], "11": ["deleafing", "spraying"], "12": ["harvesting", "deleafing"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;

-- Mango (UP/AP/Maharashtra/Karnataka)
INSERT INTO crop_catalog (jurisdiction_id, crop_slug, name, work_types, seasonal_calendar, is_active)
VALUES ('in', 'mango', '{"en": "Mango", "hi": "आम", "ta": "மாங்காய்", "te": "మామిడి", "kn": "ಮಾವಿನಕಾಯಿ"}'::jsonb,
'[
  {"slug": "pruning", "name": {"en": "Pruning & Dead Wood Removal"}, "pricing_model": "per_tree", "typical_price": {"min": 50, "max": 200, "currency": "INR"}},
  {"slug": "spraying", "name": {"en": "Pest & Fungal Spraying"}, "pricing_model": "per_tree", "typical_price": {"min": 40, "max": 120, "currency": "INR"}},
  {"slug": "harvesting", "name": {"en": "Fruit Harvesting"}, "pricing_model": "per_tree", "typical_price": {"min": 30, "max": 100, "currency": "INR"}},
  {"slug": "fertilizing", "name": {"en": "Fertilizer Application"}, "pricing_model": "per_tree", "typical_price": {"min": 30, "max": 80, "currency": "INR"}}
]'::jsonb,
'{"1": ["spraying"], "2": ["spraying"], "3": ["spraying", "harvesting"], "4": ["harvesting"], "5": ["harvesting"], "6": ["harvesting", "pruning"], "7": ["pruning"], "8": ["fertilizing"], "9": ["fertilizing"], "10": ["spraying"], "11": ["spraying"], "12": ["spraying"]}'::jsonb,
true)
ON CONFLICT (jurisdiction_id, crop_slug) DO UPDATE SET
  name = EXCLUDED.name,
  work_types = EXCLUDED.work_types,
  seasonal_calendar = EXCLUDED.seasonal_calendar,
  is_active = EXCLUDED.is_active;
