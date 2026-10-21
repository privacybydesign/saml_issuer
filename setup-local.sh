#!/bin/bash

# Setup script for local development
# This script generates a self-signed SP certificate and a sample .env file

set -e

CONFIG_DIR="$(dirname "$0")/config"
mkdir -p "$CONFIG_DIR/cert" "$CONFIG_DIR/metadata"

echo "Setting up local development environment..."

# Create the SP certificate. Only saml.crt is read by the service; it is
# published in the SP metadata.
if [ ! -f "$CONFIG_DIR/cert/saml.crt" ]; then
    echo "Generating SAML certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$CONFIG_DIR/cert/saml.pem" \
        -out "$CONFIG_DIR/cert/saml.crt" \
        -subj "/CN=saml-issuer/O=Development/C=NL" \
        2>/dev/null
    echo "SAML certificate generated."
else
    echo "SAML certificate already exists, skipping."
fi

# Create .env file if it doesn't exist
if [ ! -f "$(dirname "$0")/.env" ]; then
    echo "Creating .env file..."
    cat > "$(dirname "$0")/.env" << 'EOF2'
# Yivi/IRMA configuration
IRMA_SERVER_URL=https://is.staging.yivi.app
API_TOKEN=
IRMA_ISSUER_ID=irma-demo.sidn-pbdf

# LinkedIn OAuth (optional)
LINKEDIN_CLIENT_ID=
LINKEDIN_CLIENT_SECRET=
EOF2
    echo ".env file created. Edit it with your actual credentials."
else
    echo ".env file already exists, skipping."
fi

echo ""
echo "Setup complete!"
echo ""
echo "Put the SURFconext IdP metadata in $CONFIG_DIR/metadata/saml20-idp-remote.xml,"
echo "then start the container with:"
echo "  docker compose up --build"
echo ""
echo "The application will be available at:"
echo "  http://localhost:8080"
