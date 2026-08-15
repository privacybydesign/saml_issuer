#!/bin/bash

# Setup script for local development
# This script generates self-signed certificates for SSL and SAML signing

set -e

CONFIG_DIR="$(dirname "$0")/config"

echo "Setting up local development environment..."

# Create SSL certificate for Apache (HTTPS)
if [ ! -f "$CONFIG_DIR/ssl/tls.crt" ]; then
    echo "Generating SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$CONFIG_DIR/ssl/tls.key" \
        -out "$CONFIG_DIR/ssl/tls.crt" \
        -subj "/CN=localhost/O=Development/C=NL" \
        2>/dev/null
    echo "SSL certificate generated."
else
    echo "SSL certificate already exists, skipping."
fi

# Create SAML signing certificate
if [ ! -f "$CONFIG_DIR/cert/saml.crt" ]; then
    echo "Generating SAML signing certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$CONFIG_DIR/cert/saml.pem" \
        -out "$CONFIG_DIR/cert/saml.crt" \
        -subj "/CN=saml-issuer/O=Development/C=NL" \
        2>/dev/null
    echo "SAML signing certificate generated."
else
    echo "SAML signing certificate already exists, skipping."
fi

# Create .env file if it doesn't exist
if [ ! -f "$(dirname "$0")/.env" ]; then
    echo "Creating .env file..."
    cat > "$(dirname "$0")/.env" << 'EOF'
# SimpleSAMLphp configuration
SSP_SECRET_SALT=localdevelopmentsalt123456789
SSP_ADMIN_PASSWORD=localdev-change-me

# Yivi/IRMA configuration
IRMA_SERVER_URL=https://is.staging.yivi.app
API_TOKEN=
IRMA_ISSUER_ID=irma-demo.sidn-pbdf

# SURFconext configuration
SURFCONEXT_IDP_METADATA_URL=https://engine.test.surfconext.nl/authentication/idp/metadata

# LinkedIn OAuth (optional)
LINKEDIN_CLIENT_ID=
LINKEDIN_CLIENT_SECRET=
EOF
    echo ".env file created. Edit it with your actual credentials."
else
    echo ".env file already exists, skipping."
fi

echo ""
echo "Setup complete!"
echo ""
echo "To start the container, run:"
echo "  docker compose up --build"
echo ""
echo "The application will be available at:"
echo "  https://localhost:8443"
echo ""
echo "Note: Your browser will show a security warning for the self-signed certificate."
