"""
Weekly matching model retraining pipeline.

Loads job outcome data from PostgreSQL, engineers features from provider-customer
interactions, trains an XGBoost ranking model to predict match quality, exports
the model to ONNX format, and uploads the artifact to object storage.

Usage:
    python retrain_matching.py --db-url postgresql://... --output-dir ./models
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
import xgboost as xgb
from onnxmltools import convert_xgboost
from onnxmltools.convert.common.data_types import FloatTensorType
from sklearn.model_selection import train_test_split
from sklearn.metrics import ndcg_score

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# SQL Queries
# ---------------------------------------------------------------------------

QUERY_JOB_OUTCOMES = """
SELECT
    j.id AS job_id,
    j.customer_id,
    j.provider_id,
    j.category_id,
    j.status,
    j.urgency,
    j.budget_min,
    j.budget_max,
    j.agreed_price,
    j.created_at,
    j.completed_at,
    j.latitude AS job_lat,
    j.longitude AS job_lng,
    p.latitude AS provider_lat,
    p.longitude AS provider_lng,
    p.rating AS provider_rating,
    p.review_count AS provider_review_count,
    p.trust_score AS provider_trust_score,
    p.completed_jobs AS provider_completed_jobs,
    p.response_time_minutes AS provider_response_time,
    r.rating AS review_rating,
    EXTRACT(EPOCH FROM (j.completed_at - j.created_at)) / 3600 AS turnaround_hours,
    EXTRACT(EPOCH FROM (j.started_at - j.accepted_at)) / 60 AS arrival_minutes
FROM jobs j
JOIN providers p ON p.id = j.provider_id
LEFT JOIN reviews r ON r.job_id = j.id AND r.reviewer_id = j.customer_id
WHERE j.status IN ('completed', 'cancelled')
  AND j.provider_id IS NOT NULL
  AND j.created_at >= %s
ORDER BY j.created_at;
"""


def haversine_km(lat1, lon1, lat2, lon2):
    """Compute the great-circle distance between two points (in km)."""
    R = 6371.0
    lat1, lon1, lat2, lon2 = map(np.radians, [lat1, lon1, lat2, lon2])
    dlat = lat2 - lat1
    dlon = lon2 - lon1
    a = np.sin(dlat / 2) ** 2 + np.cos(lat1) * np.cos(lat2) * np.sin(dlon / 2) ** 2
    return R * 2 * np.arcsin(np.sqrt(a))


def load_data(db_url: str, lookback_days: int = 180) -> pd.DataFrame:
    """Load job outcome data from PostgreSQL."""
    logger.info("Connecting to database...")
    conn = psycopg2.connect(db_url)
    cutoff = datetime.utcnow() - timedelta(days=lookback_days)

    logger.info(f"Loading job outcomes since {cutoff.date()}...")
    df = pd.read_sql_query(QUERY_JOB_OUTCOMES, conn, params=(cutoff,))
    conn.close()

    logger.info(f"Loaded {len(df)} job outcome records.")
    return df


def engineer_features(df: pd.DataFrame) -> tuple[pd.DataFrame, list[str]]:
    """Create features for the matching model."""
    logger.info("Engineering features...")

    # Distance between provider and job location
    df["distance_km"] = haversine_km(
        df["job_lat"].fillna(0),
        df["job_lng"].fillna(0),
        df["provider_lat"].fillna(0),
        df["provider_lng"].fillna(0),
    )

    # Budget midpoint
    df["budget_mid"] = (
        df[["budget_min", "budget_max"]].mean(axis=1).fillna(0)
    )

    # Price deviation from budget midpoint
    df["price_deviation"] = np.where(
        df["budget_mid"] > 0,
        (df["agreed_price"].fillna(0) - df["budget_mid"]) / df["budget_mid"],
        0,
    )

    # Time features
    df["hour_of_day"] = pd.to_datetime(df["created_at"]).dt.hour
    df["day_of_week"] = pd.to_datetime(df["created_at"]).dt.dayofweek
    df["is_weekend"] = (df["day_of_week"] >= 5).astype(int)

    # Urgency encoding
    urgency_map = {"low": 0, "normal": 1, "high": 2, "emergency": 3}
    df["urgency_score"] = df["urgency"].map(urgency_map).fillna(1)

    # Outcome label: 1.0 for successfully completed + high review,
    # 0.0 for cancelled, partial for completed + lower review
    df["outcome"] = 0.0
    completed = df["status"] == "completed"
    df.loc[completed, "outcome"] = 0.5
    df.loc[completed & (df["review_rating"] >= 4), "outcome"] = 1.0
    df.loc[completed & (df["review_rating"] == 5), "outcome"] = 1.0
    df.loc[completed & (df["review_rating"] <= 2), "outcome"] = 0.2

    feature_cols = [
        "distance_km",
        "provider_rating",
        "provider_review_count",
        "provider_trust_score",
        "provider_completed_jobs",
        "provider_response_time",
        "budget_mid",
        "price_deviation",
        "urgency_score",
        "hour_of_day",
        "day_of_week",
        "is_weekend",
        "turnaround_hours",
        "arrival_minutes",
    ]

    # Fill missing values
    for col in feature_cols:
        df[col] = df[col].fillna(0)

    logger.info(f"Engineered {len(feature_cols)} features.")
    return df, feature_cols


def train_model(
    df: pd.DataFrame,
    feature_cols: list[str],
) -> xgb.XGBRegressor:
    """Train an XGBoost model to predict match quality."""
    logger.info("Training XGBoost model...")

    X = df[feature_cols].values
    y = df["outcome"].values

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42
    )

    model = xgb.XGBRegressor(
        objective="reg:squarederror",
        n_estimators=200,
        max_depth=6,
        learning_rate=0.05,
        subsample=0.8,
        colsample_bytree=0.8,
        min_child_weight=5,
        reg_alpha=0.1,
        reg_lambda=1.0,
        random_state=42,
        n_jobs=-1,
    )

    model.fit(
        X_train,
        y_train,
        eval_set=[(X_test, y_test)],
        verbose=False,
    )

    # Evaluate
    y_pred = model.predict(X_test)
    mse = np.mean((y_test - y_pred) ** 2)
    logger.info(f"Test MSE: {mse:.4f}")

    # Feature importance
    importances = dict(zip(feature_cols, model.feature_importances_))
    logger.info("Feature importances:")
    for feat, imp in sorted(importances.items(), key=lambda x: -x[1]):
        logger.info(f"  {feat}: {imp:.4f}")

    # NDCG evaluation (treating this as a ranking problem)
    try:
        # Group by a proxy for query groups and compute NDCG
        ndcg = ndcg_score(
            y_test.reshape(1, -1),
            y_pred.reshape(1, -1),
        )
        logger.info(f"NDCG@all: {ndcg:.4f}")
    except ValueError:
        logger.warning("Could not compute NDCG (may need more data).")

    return model


def export_onnx(
    model: xgb.XGBRegressor,
    feature_cols: list[str],
    output_path: str,
) -> str:
    """Export the trained model to ONNX format."""
    logger.info(f"Exporting model to ONNX: {output_path}")

    initial_type = [("features", FloatTensorType([None, len(feature_cols)]))]
    onnx_model = convert_xgboost(
        model,
        initial_types=initial_type,
        target_opset=12,
    )

    onnx.save_model(onnx_model, output_path)
    file_size = os.path.getsize(output_path) / 1024
    logger.info(f"ONNX model saved: {output_path} ({file_size:.1f} KB)")
    return output_path


def upload_artifact(model_path: str, bucket: str = "seva-models"):
    """Upload the model artifact to S3-compatible storage."""
    logger.info(f"Uploading model to {bucket}...")

    try:
        import boto3

        s3 = boto3.client(
            "s3",
            endpoint_url=os.environ.get("S3_ENDPOINT_URL"),
            aws_access_key_id=os.environ.get("S3_ACCESS_KEY_ID"),
            aws_secret_access_key=os.environ.get("S3_SECRET_ACCESS_KEY"),
        )

        timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
        key = f"matching/matching_model_{timestamp}.onnx"

        s3.upload_file(model_path, bucket, key)
        logger.info(f"Uploaded to s3://{bucket}/{key}")

        # Also upload as "latest"
        latest_key = "matching/matching_model_latest.onnx"
        s3.upload_file(model_path, bucket, latest_key)
        logger.info(f"Updated latest: s3://{bucket}/{latest_key}")

    except ImportError:
        logger.warning("boto3 not installed; skipping S3 upload.")
    except Exception as e:
        logger.error(f"Failed to upload model: {e}")


def main():
    parser = argparse.ArgumentParser(description="Retrain Seva matching model")
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
        default=180,
        help="Days of historical data to use",
    )
    parser.add_argument(
        "--upload",
        action="store_true",
        help="Upload model to S3",
    )
    args = parser.parse_args()

    if not args.db_url:
        logger.error("DATABASE_URL not set. Use --db-url or set the environment variable.")
        sys.exit(1)

    os.makedirs(args.output_dir, exist_ok=True)

    # Pipeline stages
    df = load_data(args.db_url, args.lookback_days)

    if len(df) < 100:
        logger.error(f"Insufficient data ({len(df)} records). Need at least 100.")
        sys.exit(1)

    df, feature_cols = engineer_features(df)
    model = train_model(df, feature_cols)

    timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    model_path = os.path.join(args.output_dir, f"matching_model_{timestamp}.onnx")
    export_onnx(model, feature_cols, model_path)

    if args.upload:
        upload_artifact(model_path)

    logger.info("Matching model retraining complete.")


if __name__ == "__main__":
    main()
