-- name: GetProviderEarningsHistory :many
-- Monthly earnings for the last N months for a given provider.
SELECT
    DATE_TRUNC('month', t.paid_at) AS month,
    COALESCE(SUM(t.provider_payout_amount::numeric), 0)::float8 AS earnings,
    COUNT(*)::int AS job_count
FROM transactions t
JOIN jobs j ON j.id = t.job_id
WHERE j.provider_id = $1
  AND t.payment_status = 'captured'
  AND t.paid_at >= NOW() - ($2::int || ' months')::interval
GROUP BY DATE_TRUNC('month', t.paid_at)
ORDER BY month ASC;

-- name: GetDemandByCategory :many
-- Job request counts per category within a radius of a point.
SELECT
    c.id AS category_id,
    c.slug AS category_slug,
    COALESCE((c.name->>'en')::text, c.slug) AS category_name,
    COUNT(j.id)::int AS demand_count
FROM jobs j
JOIN categories c ON c.id = j.category_id
WHERE j.status IN ('posted', 'quoted', 'accepted', 'in_progress', 'completed')
  AND j.location IS NOT NULL
  AND ST_DWithin(
      j.location::geography,
      ST_MakePoint($1, $2)::geography,
      $3
  )
  AND j.created_at >= NOW() - INTERVAL '90 days'
GROUP BY c.id, c.slug, c.name
ORDER BY demand_count DESC;

-- name: GetDemandByPostcode :many
-- Job request counts per postcode for heatmap display.
SELECT
    j.postcode::text AS postcode,
    COUNT(j.id)::int AS demand_count,
    AVG(ST_Y(j.location::geometry))::float8 AS lat,
    AVG(ST_X(j.location::geometry))::float8 AS lng
FROM jobs j
WHERE j.status IN ('posted', 'quoted', 'accepted', 'in_progress', 'completed')
  AND j.postcode IS NOT NULL
  AND j.location IS NOT NULL
  AND ST_DWithin(
      j.location::geography,
      ST_MakePoint($1, $2)::geography,
      $3
  )
  AND j.created_at >= NOW() - INTERVAL '90 days'
GROUP BY j.postcode
ORDER BY demand_count DESC;

-- name: GetProviderPerformanceMetrics :one
-- Provider performance: response rate, completion rate, avg rating trend.
SELECT
    -- Response rate: percentage of jobs where provider submitted a quote
    COALESCE(
        (SELECT COUNT(DISTINCT jq.job_id)::float8 / NULLIF(COUNT(DISTINCT j2.id), 0)
         FROM jobs j2
         LEFT JOIN job_quotes jq ON jq.job_id = j2.id AND jq.provider_id = $1
         WHERE j2.category_id IN (SELECT category_id FROM provider_categories WHERE provider_id = $1)
           AND j2.created_at >= NOW() - INTERVAL '90 days'
           AND j2.status != 'draft'),
        0
    )::float8 AS response_rate,
    -- Completion rate: completed / (accepted + in_progress + completed)
    COALESCE(
        (SELECT COUNT(*)::float8 FROM jobs WHERE provider_id = $1 AND status = 'completed' AND created_at >= NOW() - INTERVAL '90 days')
        / NULLIF(
            (SELECT COUNT(*)::float8 FROM jobs WHERE provider_id = $1 AND status IN ('accepted', 'in_progress', 'completed') AND created_at >= NOW() - INTERVAL '90 days'),
            0
        ),
        0
    )::float8 AS completion_rate,
    -- Average rating
    COALESCE(
        (SELECT AVG(r.rating)::float8 FROM reviews r WHERE r.reviewee_id = $1),
        0
    )::float8 AS avg_rating,
    -- Total reviews
    (SELECT COUNT(*)::int FROM reviews r WHERE r.reviewee_id = $1) AS total_reviews,
    -- Total earnings
    COALESCE(
        (SELECT SUM(t.provider_payout_amount::numeric)::float8
         FROM transactions t
         JOIN jobs j ON j.id = t.job_id
         WHERE j.provider_id = $1 AND t.payment_status = 'captured'),
        0
    )::float8 AS total_earnings;

-- name: GetPeakDemandHours :many
-- Job creation counts by hour of day for the provider's service area.
SELECT
    EXTRACT(HOUR FROM j.created_at)::int AS hour_of_day,
    COUNT(j.id)::int AS demand_count
FROM jobs j
WHERE j.status IN ('posted', 'quoted', 'accepted', 'in_progress', 'completed')
  AND j.category_id IN (SELECT category_id FROM provider_categories WHERE provider_id = $1)
  AND j.created_at >= NOW() - INTERVAL '90 days'
GROUP BY EXTRACT(HOUR FROM j.created_at)
ORDER BY hour_of_day ASC;

-- name: GetCompetitorDensity :many
-- Count of providers per category in each postcode near the given provider.
SELECT
    pp.postcode::text AS postcode,
    c.slug AS category_slug,
    COALESCE((c.name->>'en')::text, c.slug) AS category_name,
    COUNT(DISTINCT pp.id)::int AS provider_count
FROM provider_profiles pp
JOIN provider_categories pc ON pc.provider_id = pp.id
JOIN categories c ON c.id = pc.category_id
WHERE pp.postcode IS NOT NULL
  AND pc.category_id IN (SELECT category_id FROM provider_categories WHERE provider_id = $1)
GROUP BY pp.postcode, c.slug, c.name
ORDER BY provider_count DESC
LIMIT 100;
