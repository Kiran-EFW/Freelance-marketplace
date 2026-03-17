-- Remove seeded crop catalog entries for India jurisdiction
DELETE FROM crop_catalog WHERE jurisdiction_id = 'in' AND crop_slug IN (
    'coconut', 'rice', 'wheat', 'sugarcane', 'cotton',
    'arecanut', 'rubber', 'tea', 'coffee', 'banana', 'mango'
);
