#!/usr/bin/env bash
set -euo pipefail

echo "=== Generating API Clients from OpenAPI Spec ==="
echo ""

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SPEC_FILE="$PROJECT_ROOT/api/openapi.yaml"

if [ ! -f "$SPEC_FILE" ]; then
    echo "Error: OpenAPI spec not found at $SPEC_FILE"
    exit 1
fi

# Check for openapi-generator or use npx
if command -v openapi-generator-cli &> /dev/null; then
    GENERATOR="openapi-generator-cli"
elif command -v npx &> /dev/null; then
    GENERATOR="npx @openapitools/openapi-generator-cli"
else
    echo "Error: openapi-generator-cli not found. Install via npm:"
    echo "  npm install -g @openapitools/openapi-generator-cli"
    exit 1
fi

# Generate TypeScript client for SvelteKit web app
echo "Generating TypeScript client for web..."
$GENERATOR generate \
    -i "$SPEC_FILE" \
    -g typescript-fetch \
    -o "$PROJECT_ROOT/web/src/lib/api/generated" \
    --additional-properties=supportsES6=true,npmVersion=10.0.0,typescriptThreePlus=true \
    --skip-validate-spec \
    2>/dev/null

echo "✓ TypeScript client generated at web/src/lib/api/generated/"

# Generate Dart client for Flutter mobile apps
echo ""
echo "Generating Dart client for mobile..."
$GENERATOR generate \
    -i "$SPEC_FILE" \
    -g dart \
    -o "$PROJECT_ROOT/mobile/packages/api_client" \
    --additional-properties=pubName=api_client,pubAuthor=Seva \
    --skip-validate-spec \
    2>/dev/null

echo "✓ Dart client generated at mobile/packages/api_client/"

echo ""
echo "=== API clients generated successfully ==="
echo ""
echo "Next steps:"
echo "  Web:    Import from '\$lib/api/generated' in SvelteKit"
echo "  Mobile: Add api_client to pubspec.yaml dependencies"
