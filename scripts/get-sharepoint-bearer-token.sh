#!/usr/bin/env bash

set -euo pipefail

# Check that a required environment variable is set and not empty
check_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    if [ -z "$var_value" ]; then
        echo "Error: Environment variable '$var_name' is not set or is empty." >&2
        exit 1
    fi
}

# Validate required environment variables
check_var "BATON_AZURE_TENANT_ID"
check_var "BATON_AZURE_CLIENT_ID"
check_var "BATON_SHAREPOINT_DOMAIN"
check_var "BATON_PFX_CERTIFICATE"
check_var "BATON_PFX_CERTIFICATE_PASSWORD"

PFX_PASSWORD="$BATON_PFX_CERTIFICATE_PASSWORD"
AUDIENCE="https://login.microsoftonline.com/$BATON_AZURE_TENANT_ID/v2.0"
SCOPE="https://$BATON_SHAREPOINT_DOMAIN.sharepoint.com/.default"

# TEMP FILES
TMP_PFX="temp_cert.pfx"
KEY_FILE="private.key"
CERT_FILE="public.crt"

# Decode PFX certificate from base64
echo "$BATON_PFX_CERTIFICATE" | base64 -d > "$TMP_PFX"

# Extract key and certificate from PFX
openssl pkcs12 -in "$TMP_PFX" -nocerts -nodes -out "$KEY_FILE" -passin pass:"$PFX_PASSWORD"
openssl pkcs12 -in "$TMP_PFX" -clcerts -nokeys -out "$CERT_FILE" -passin pass:"$PFX_PASSWORD"

# Generate x5t
X5T=$(openssl x509 -in "$CERT_FILE" -noout -fingerprint -sha1 | cut -d'=' -f2 | tr -d ':' | xxd -r -p | base64 | tr '+/' '-_' | tr -d '=')

# Get timestamps
NOW=$(date -u +%s)
EXP=$((NOW + 3600))
NBF=$((NOW - 300))
JTI=$(uuidgen)

# Create header and payload
HEADER=$(jq -n --arg x5t "$X5T" '{alg:"RS256", typ:"JWT", x5t:$x5t}')
PAYLOAD=$(jq -n \
    --arg aud "$AUDIENCE" \
    --arg iss "$BATON_AZURE_CLIENT_ID" \
    --arg sub "$BATON_AZURE_CLIENT_ID" \
    --arg jti "$JTI" \
    --argjson nbf "$NBF" \
    --argjson exp "$EXP" \
    '{aud:$aud, iss:$iss, sub:$sub, jti:$jti, nbf:$nbf, exp:$exp}')

# Base64Url encode helper
b64url() {
    openssl base64 -e -A | tr '+/' '-_' | tr -d '='
}

HEADER_ENC=$(echo -n "$HEADER" | b64url)
PAYLOAD_ENC=$(echo -n "$PAYLOAD" | b64url)
UNSIGNED="$HEADER_ENC.$PAYLOAD_ENC"

# Sign JWT
SIGNATURE=$(echo -n "$UNSIGNED" | openssl dgst -sha256 -sign "$KEY_FILE" | openssl base64 -A | tr '+/' '-_' | tr -d '=')
JWT="$UNSIGNED.$SIGNATURE"

# Build form body
BODY="client_id=${BATON_AZURE_CLIENT_ID}"\
"&scope=${SCOPE}"\
"&grant_type=client_credentials"\
"&client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer"\
"&client_assertion=${JWT}"

# Request token
curl -X POST \
  "https://login.microsoftonline.com/${BATON_AZURE_TENANT_ID}/oauth2/v2.0/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data "${BODY}"

# Clean up
rm -f "$TMP_PFX" "$KEY_FILE" "$CERT_FILE"
