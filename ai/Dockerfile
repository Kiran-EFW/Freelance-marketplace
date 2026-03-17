FROM python:3.11-slim

LABEL maintainer="Seva Platform <eng@seva.app>"
LABEL description="Seva AI/ML pipeline runner"

# Install system dependencies for psycopg2 and scientific computing
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        libpq-dev \
        curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy pipeline code and prompts
COPY pipelines/ ./pipelines/
COPY prompts/ ./prompts/

# Create models output directory
RUN mkdir -p /app/models

# Default entrypoint runs the matching model retraining
# Override with: docker run seva-ai python pipelines/retrain_fraud.py
ENTRYPOINT ["python"]
CMD ["pipelines/retrain_matching.py", "--output-dir", "/app/models"]
