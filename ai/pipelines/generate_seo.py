"""
Batch SEO content generation pipeline.

Queries all service + city + area combinations from the database,
generates SEO landing page content using Claude, and stores the
generated content back in the database.

Usage:
    python generate_seo.py --db-url postgresql://... --batch-size 20
"""

import argparse
import json
import logging
import os
import sys
import time
from datetime import datetime

import anthropic
import psycopg2
import psycopg2.extras

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# SQL Queries
# ---------------------------------------------------------------------------

QUERY_COMBINATIONS = """
SELECT DISTINCT
    c.id AS category_id,
    c.name AS category_name,
    c.slug AS category_slug,
    l.city,
    l.area,
    l.country,
    COUNT(DISTINCT p.id) AS provider_count,
    ROUND(AVG(p.rating)::numeric, 1) AS avg_rating,
    ROUND(AVG(p.hourly_rate)::numeric, 0) AS avg_hourly_rate,
    ROUND(MIN(p.hourly_rate)::numeric, 0) AS min_rate,
    ROUND(MAX(p.hourly_rate)::numeric, 0) AS max_rate
FROM categories c
CROSS JOIN (
    SELECT DISTINCT city, area, country FROM service_areas WHERE is_active = true
) l
JOIN providers p ON p.id IN (
    SELECT provider_id FROM provider_categories WHERE category_id = c.id
)
JOIN service_areas sa ON sa.city = l.city AND sa.area = l.area
WHERE c.is_active = true
  AND NOT EXISTS (
    SELECT 1 FROM seo_content sc
    WHERE sc.category_id = c.id
      AND sc.city = l.city
      AND sc.area = l.area
      AND sc.generated_at >= NOW() - INTERVAL '30 days'
  )
GROUP BY c.id, c.name, c.slug, l.city, l.area, l.country
HAVING COUNT(DISTINCT p.id) >= 3
ORDER BY COUNT(DISTINCT p.id) DESC;
"""

QUERY_TOP_SERVICES = """
SELECT DISTINCT j.title, COUNT(*) as cnt
FROM jobs j
WHERE j.category_id = %s
  AND j.status = 'completed'
GROUP BY j.title
ORDER BY cnt DESC
LIMIT 5;
"""

UPSERT_SEO_CONTENT = """
INSERT INTO seo_content (category_id, city, area, country, content, generated_at)
VALUES (%s, %s, %s, %s, %s, %s)
ON CONFLICT (category_id, city, area)
DO UPDATE SET content = EXCLUDED.content, generated_at = EXCLUDED.generated_at;
"""

SYSTEM_PROMPT_PATH = os.path.join(
    os.path.dirname(__file__), "..", "prompts", "seo_content.md"
)


def load_system_prompt() -> str:
    """Load the SEO content system prompt."""
    with open(SYSTEM_PROMPT_PATH, "r") as f:
        return f.read()


def load_combinations(db_url: str) -> list[dict]:
    """Load all service + city + area combinations needing content."""
    conn = psycopg2.connect(db_url)
    cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)

    logger.info("Loading service/area combinations needing content...")
    cur.execute(QUERY_COMBINATIONS)
    combinations = [dict(row) for row in cur.fetchall()]

    conn.close()
    logger.info(f"Found {len(combinations)} combinations to generate.")
    return combinations


def get_top_services(db_url: str, category_id: str) -> list[str]:
    """Get the most common job titles for a category."""
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()
    cur.execute(QUERY_TOP_SERVICES, (category_id,))
    services = [row[0] for row in cur.fetchall()]
    conn.close()
    return services


def generate_content(
    client: anthropic.Anthropic,
    system_prompt: str,
    combo: dict,
    top_services: list[str],
) -> dict | None:
    """Generate SEO content for a single combination using Claude."""
    currency_map = {
        "IN": "INR",
        "AE": "AED",
        "GB": "GBP",
        "US": "USD",
    }
    currency = currency_map.get(combo.get("country", "IN"), "INR")

    user_message = f"""Generate SEO landing page content for:

Service: {combo['category_name']}
City: {combo['city']}
Area: {combo['area']}
Country: {combo.get('country', 'IN')}
Currency: {currency}
Provider count: {combo['provider_count']}
Average rating: {combo.get('avg_rating', 'N/A')}
Price range: {currency} {combo.get('min_rate', 'N/A')} - {combo.get('max_rate', 'N/A')}/hr
Top services requested: {', '.join(top_services) if top_services else 'General'}

Please generate the complete SEO content in JSON format as specified in your instructions."""

    try:
        response = client.messages.create(
            model="claude-sonnet-4-20250514",
            max_tokens=4096,
            system=system_prompt,
            messages=[{"role": "user", "content": user_message}],
        )

        content_text = response.content[0].text

        # Try to parse as JSON
        # The response may have markdown code fences
        if "```json" in content_text:
            content_text = content_text.split("```json")[1].split("```")[0]
        elif "```" in content_text:
            content_text = content_text.split("```")[1].split("```")[0]

        return json.loads(content_text.strip())

    except json.JSONDecodeError as e:
        logger.warning(f"Failed to parse JSON for {combo['category_name']} in {combo['area']}, {combo['city']}: {e}")
        # Store raw text as fallback
        return {"raw_content": content_text, "parse_error": str(e)}
    except anthropic.APIError as e:
        logger.error(f"API error for {combo['category_name']} in {combo['area']}: {e}")
        return None


def store_content(db_url: str, combo: dict, content: dict):
    """Store generated content in the database."""
    conn = psycopg2.connect(db_url)
    cur = conn.cursor()

    cur.execute(
        UPSERT_SEO_CONTENT,
        (
            combo["category_id"],
            combo["city"],
            combo["area"],
            combo.get("country", "IN"),
            json.dumps(content, ensure_ascii=False),
            datetime.utcnow(),
        ),
    )

    conn.commit()
    conn.close()


def main():
    parser = argparse.ArgumentParser(description="Generate SEO content for Seva")
    parser.add_argument(
        "--db-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL",
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=20,
        help="Number of combinations to process per run",
    )
    parser.add_argument(
        "--delay",
        type=float,
        default=1.0,
        help="Delay between API calls in seconds",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Generate but do not store content",
    )
    args = parser.parse_args()

    if not args.db_url:
        logger.error("DATABASE_URL not set.")
        sys.exit(1)

    api_key = os.environ.get("ANTHROPIC_API_KEY")
    if not api_key:
        logger.error("ANTHROPIC_API_KEY not set.")
        sys.exit(1)

    client = anthropic.Anthropic(api_key=api_key)
    system_prompt = load_system_prompt()

    combinations = load_combinations(args.db_url)
    batch = combinations[: args.batch_size]

    logger.info(f"Processing {len(batch)} combinations...")

    success_count = 0
    error_count = 0

    for i, combo in enumerate(batch):
        label = f"{combo['category_name']} in {combo['area']}, {combo['city']}"
        logger.info(f"[{i+1}/{len(batch)}] Generating content for: {label}")

        top_services = get_top_services(args.db_url, combo["category_id"])
        content = generate_content(client, system_prompt, combo, top_services)

        if content is None:
            error_count += 1
            continue

        if not args.dry_run:
            store_content(args.db_url, combo, content)
            logger.info(f"  Stored content for: {label}")
        else:
            logger.info(f"  [DRY RUN] Would store content for: {label}")

        success_count += 1

        # Rate limit
        if i < len(batch) - 1:
            time.sleep(args.delay)

    logger.info(
        f"SEO content generation complete. "
        f"Success: {success_count}, Errors: {error_count}"
    )


if __name__ == "__main__":
    main()
