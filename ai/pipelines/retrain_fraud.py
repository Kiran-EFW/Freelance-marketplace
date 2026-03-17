"""
Fraud detection model retraining pipeline.

Loads user behavior data (review patterns, account metadata, interaction signals)
from PostgreSQL, engineers anomaly features, trains an Isolation Forest model
to detect fraudulent reviews and suspicious accounts, and exports to ONNX.

Usage:
    python retrain_fraud.py --db-url postgresql://... --output-dir ./models
"""

import argparse
import logging
import os
import sys
from datetime import datetime, timedelta

import numpy as np
import onnx
import pandas as pd
import psycopg2
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
from skl2onnx import convert_sklearn
from skl2onnx.common.data_types import FloatTensorType
from sklearn.metrics import classification_report

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# SQL Queries
# ---------------------------------------------------------------------------

QUERY_USER_BEHAVIOR = """
SELECT
    u.id AS user_id,
    u.role,
    u.created_at AS account_created_at,
    u.kyc_status,
    EXTRACT(EPOCH FROM (NOW() - u.created_at)) / 86400 AS account_age_days,
    COALESCE(rev.review_count, 0) AS review_count,
    COALESCE(rev.avg_rating_given, 0) AS avg_rating_given,
    COALESCE(rev.rating_stddev, 0) AS rating_stddev,
    COALESCE(rev.reviews_last_24h, 0) AS reviews_last_24h,
    COALESCE(rev.reviews_last_7d, 0) AS reviews_last_7d,
    COALESCE(rev.reviews_last_30d, 0) AS reviews_last_30d,
    COALESCE(rev.avg_review_length, 0) AS avg_review_length,
    COALESCE(rev.min_time_between_reviews_min, 999999) AS min_time_between_reviews_min,
    COALESCE(rev.duplicate_review_text_count, 0) AS duplicate_review_text_count,
    COALESCE(rev.five_star_ratio, 0) AS five_star_ratio,
    COALESCE(rev.one_star_ratio, 0) AS one_star_ratio,
    COALESCE(job.total_jobs, 0) AS total_jobs,
    COALESCE(job.cancelled_jobs, 0) AS cancelled_jobs,
    COALESCE(job.disputed_jobs, 0) AS disputed_jobs,
    COALESCE(job.avg_job_value, 0) AS avg_job_value,
    COALESCE(login.distinct_ips_30d, 0) AS distinct_ips_30d,
    COALESCE(login.distinct_devices_30d, 0) AS distinct_devices_30d,
    COALESCE(flag.times_flagged, 0) AS times_flagged,
    COALESCE(flag.is_suspended, false) AS is_suspended
FROM users u
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS review_count,
        AVG(r.rating) AS avg_rating_given,
        STDDEV(r.rating) AS rating_stddev,
        COUNT(*) FILTER (WHERE r.created_at >= NOW() - INTERVAL '24 hours') AS reviews_last_24h,
        COUNT(*) FILTER (WHERE r.created_at >= NOW() - INTERVAL '7 days') AS reviews_last_7d,
        COUNT(*) FILTER (WHERE r.created_at >= NOW() - INTERVAL '30 days') AS reviews_last_30d,
        AVG(LENGTH(r.comment)) AS avg_review_length,
        MIN(EXTRACT(EPOCH FROM (r.created_at - LAG(r.created_at) OVER (ORDER BY r.created_at)))) / 60
            AS min_time_between_reviews_min,
        COUNT(*) - COUNT(DISTINCT LEFT(r.comment, 50)) AS duplicate_review_text_count,
        COUNT(*) FILTER (WHERE r.rating = 5)::float / GREATEST(COUNT(*), 1) AS five_star_ratio,
        COUNT(*) FILTER (WHERE r.rating = 1)::float / GREATEST(COUNT(*), 1) AS one_star_ratio
    FROM reviews r
    WHERE r.reviewer_id = u.id
) rev ON true
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS total_jobs,
        COUNT(*) FILTER (WHERE j.status = 'cancelled') AS cancelled_jobs,
        COUNT(*) FILTER (WHERE j.status = 'disputed') AS disputed_jobs,
        AVG(j.agreed_price) AS avg_job_value
    FROM jobs j
    WHERE j.customer_id = u.id OR j.provider_id = u.id
) job ON true
LEFT JOIN LATERAL (
    SELECT
        COUNT(DISTINCT ip_address) AS distinct_ips_30d,
        COUNT(DISTINCT device_id) AS distinct_devices_30d
    FROM login_events le
    WHERE le.user_id = u.id
      AND le.created_at >= NOW() - INTERVAL '30 days'
) login ON true
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS times_flagged,
        bool_or(status = 'suspended') AS is_suspended
    FROM user_flags uf
    WHERE uf.user_id = u.id
) flag ON true
WHERE u.created_at >= %s;
"""


def load_data(db_url: str, lookback_days: int = 365) -> pd.DataFrame:
    """Load user behavior data from PostgreSQL."""
    logger.info("Connecting to database...")
    conn = psycopg2.connect(db_url)
    cutoff = datetime.utcnow() - timedelta(days=lookback_days)

    logger.info(f"Loading user behavior data since {cutoff.date()}...")
    df = pd.read_sql_query(QUERY_USER_BEHAVIOR, conn, params=(cutoff,))
    conn.close()

    logger.info(f"Loaded {len(df)} user records.")
    return df


def engineer_features(df: pd.DataFrame) -> tuple[pd.DataFrame, list[str]]:
    """Create anomaly detection features."""
    logger.info("Engineering fraud detection features...")

    # Review velocity: reviews per day of account age
    df["review_velocity"] = np.where(
        df["account_age_days"] > 0,
        df["review_count"] / df["account_age_days"],
        0,
    )

    # Review burst score: ratio of recent reviews to total
    df["review_burst_7d"] = np.where(
        df["review_count"] > 0,
        df["reviews_last_7d"] / df["review_count"],
        0,
    )

    # Review burst score 24h
    df["review_burst_24h"] = np.where(
        df["review_count"] > 0,
        df["reviews_last_24h"] / df["review_count"],
        0,
    )

    # Rating polarization: tendency toward extreme ratings
    df["rating_polarization"] = df["five_star_ratio"] + df["one_star_ratio"]

    # Cancellation rate
    df["cancellation_rate"] = np.where(
        df["total_jobs"] > 0,
        df["cancelled_jobs"] / df["total_jobs"],
        0,
    )

    # Dispute rate
    df["dispute_rate"] = np.where(
        df["total_jobs"] > 0,
        df["disputed_jobs"] / df["total_jobs"],
        0,
    )

    # IP diversity (high number of distinct IPs might indicate VPN/proxy use)
    df["ip_diversity"] = df["distinct_ips_30d"]

    # Device diversity
    df["device_diversity"] = df["distinct_devices_30d"]

    # Review authenticity signals
    df["short_review_ratio"] = np.where(
        df["review_count"] > 0,
        (df["avg_review_length"] < 10).astype(float),
        0,
    )

    # Duplicate text ratio
    df["duplicate_text_ratio"] = np.where(
        df["review_count"] > 0,
        df["duplicate_review_text_count"] / df["review_count"],
        0,
    )

    # KYC encoding
    kyc_map = {"not_started": 0, "pending": 1, "verified": 2, "rejected": -1}
    df["kyc_score"] = df["kyc_status"].map(kyc_map).fillna(0)

    feature_cols = [
        "account_age_days",
        "review_count",
        "avg_rating_given",
        "rating_stddev",
        "review_velocity",
        "review_burst_7d",
        "review_burst_24h",
        "min_time_between_reviews_min",
        "rating_polarization",
        "duplicate_text_ratio",
        "avg_review_length",
        "total_jobs",
        "cancellation_rate",
        "dispute_rate",
        "avg_job_value",
        "ip_diversity",
        "device_diversity",
        "times_flagged",
        "kyc_score",
    ]

    for col in feature_cols:
        df[col] = df[col].fillna(0)
        df[col] = df[col].replace([np.inf, -np.inf], 0)

    logger.info(f"Engineered {len(feature_cols)} features.")
    return df, feature_cols


def train_model(
    df: pd.DataFrame,
    feature_cols: list[str],
) -> tuple[IsolationForest, StandardScaler]:
    """Train an Isolation Forest anomaly detection model."""
    logger.info("Training Isolation Forest model...")

    X = df[feature_cols].values

    # Standardize features
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(X)

    model = IsolationForest(
        n_estimators=200,
        max_samples="auto",
        contamination=0.05,  # Expect ~5% of accounts to be suspicious
        max_features=1.0,
        bootstrap=False,
        random_state=42,
        n_jobs=-1,
    )

    model.fit(X_scaled)

    # Score all users
    scores = model.decision_function(X_scaled)
    predictions = model.predict(X_scaled)

    n_anomalies = (predictions == -1).sum()
    n_normal = (predictions == 1).sum()
    logger.info(f"Detected {n_anomalies} anomalous accounts ({n_anomalies/len(df)*100:.1f}%)")
    logger.info(f"Normal accounts: {n_normal}")
    logger.info(f"Score range: [{scores.min():.3f}, {scores.max():.3f}]")
    logger.info(f"Score mean: {scores.mean():.3f}, std: {scores.std():.3f}")

    # If we have known fraud labels, evaluate
    if "is_suspended" in df.columns:
        known_fraud = df["is_suspended"].astype(bool)
        if known_fraud.sum() > 0:
            detected = predictions == -1
            precision = (detected & known_fraud).sum() / max(detected.sum(), 1)
            recall = (detected & known_fraud).sum() / max(known_fraud.sum(), 1)
            logger.info(f"Against known suspensions: precision={precision:.2f}, recall={recall:.2f}")

    return model, scaler


def export_onnx(
    model: IsolationForest,
    scaler: StandardScaler,
    feature_cols: list[str],
    output_dir: str,
) -> str:
    """Export the model and scaler to ONNX format."""
    timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")

    # Export scaler
    scaler_path = os.path.join(output_dir, f"fraud_scaler_{timestamp}.onnx")
    initial_type = [("features", FloatTensorType([None, len(feature_cols)]))]
    onnx_scaler = convert_sklearn(scaler, initial_types=initial_type, target_opset=12)
    onnx.save_model(onnx_scaler, scaler_path)
    logger.info(f"Scaler saved: {scaler_path}")

    # Export model
    model_path = os.path.join(output_dir, f"fraud_model_{timestamp}.onnx")
    onnx_model = convert_sklearn(model, initial_types=initial_type, target_opset=12)
    onnx.save_model(onnx_model, model_path)
    logger.info(f"Model saved: {model_path}")

    return model_path


def upload_artifacts(output_dir: str, bucket: str = "seva-models"):
    """Upload model artifacts to S3-compatible storage."""
    logger.info(f"Uploading artifacts to {bucket}...")

    try:
        import boto3

        s3 = boto3.client(
            "s3",
            endpoint_url=os.environ.get("S3_ENDPOINT_URL"),
            aws_access_key_id=os.environ.get("S3_ACCESS_KEY_ID"),
            aws_secret_access_key=os.environ.get("S3_SECRET_ACCESS_KEY"),
        )

        for filename in os.listdir(output_dir):
            if filename.startswith("fraud_") and filename.endswith(".onnx"):
                filepath = os.path.join(output_dir, filename)
                key = f"fraud/{filename}"
                s3.upload_file(filepath, bucket, key)
                logger.info(f"Uploaded: s3://{bucket}/{key}")

                # Also upload as latest
                latest_name = filename.rsplit("_", 1)[0] + "_latest.onnx"
                latest_key = f"fraud/{latest_name}"
                s3.upload_file(filepath, bucket, latest_key)

    except ImportError:
        logger.warning("boto3 not installed; skipping S3 upload.")
    except Exception as e:
        logger.error(f"Failed to upload artifacts: {e}")


def main():
    parser = argparse.ArgumentParser(description="Retrain Seva fraud detection model")
    parser.add_argument(
        "--db-url",
        default=os.environ.get("DATABASE_URL"),
        help="PostgreSQL connection URL",
    )
    parser.add_argument(
        "--output-dir",
        default="./models",
        help="Directory to save model artifacts",
    )
    parser.add_argument(
        "--lookback-days",
        type=int,
        default=365,
        help="Days of historical data to use",
    )
    parser.add_argument("--upload", action="store_true", help="Upload models to S3")
    args = parser.parse_args()

    if not args.db_url:
        logger.error("DATABASE_URL not set.")
        sys.exit(1)

    os.makedirs(args.output_dir, exist_ok=True)

    df = load_data(args.db_url, args.lookback_days)

    if len(df) < 50:
        logger.error(f"Insufficient data ({len(df)} records). Need at least 50.")
        sys.exit(1)

    df, feature_cols = engineer_features(df)
    model, scaler = train_model(df, feature_cols)
    export_onnx(model, scaler, feature_cols, args.output_dir)

    if args.upload:
        upload_artifacts(args.output_dir)

    logger.info("Fraud detection model retraining complete.")


if __name__ == "__main__":
    main()
