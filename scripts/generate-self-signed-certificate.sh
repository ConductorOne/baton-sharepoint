#!/usr/bin/env bash

set -euo pipefail

# Display usage help
usage() {
    echo "Usage: $0 -o <cert-path> -p <password> [-d <days>]"
    echo ""
    echo "  -o Path and filename (without extension) for the certificate (e.g., ./certs/MyCert)"
    echo "     The directory must exist and the file must not already exist."
    echo "  -p Password for the PFX certificate"
    echo "  -d Number of validity days (default: 730)"
    exit 1
}

# Check that a required environment variable is set and not empty
check_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    if [ -z "$var_value" ]; then
        echo "Error: Environment variable '$var_name' is not set or is empty." >&2
        exit 1
    fi
}

# Default value
DAYS=730

# Parse arguments
while getopts "o:p:d:" opt; do
    case "$opt" in
        o) CERT_PATH="$OPTARG" ;;
        p) CERT_PASSWORD="$OPTARG" ;;
        d) DAYS="$OPTARG" ;;
        *) usage ;;
    esac
done

# Validate required arguments
if [ -z "${CERT_PATH:-}" ] || [ -z "${CERT_PASSWORD:-}" ]; then
    usage
fi

# Ensure environment variable is defined
check_var "BATON_AZURE_TENANT_ID"

# Check that the parent directory exists
CERT_DIR="$(dirname "$CERT_PATH")"
if [ ! -d "$CERT_DIR" ]; then
    echo "Error: Directory '$CERT_DIR' does not exist." >&2
    exit 1
fi

# Check that the file path doesn't already exist
if [ -e "${CERT_PATH}.pfx" ]; then
    echo "Error: File '${CERT_PATH}.pfx' already exists." >&2
    exit 1
fi

# Generate certificate
openssl genrsa -out "${CERT_PATH}.key" 2048
openssl req -x509 -new -key "${CERT_PATH}.key" -out "${CERT_PATH}.crt" -days "$DAYS" -sha256 -subj "/CN=tenant-id-$BATON_AZURE_TENANT_ID"
openssl pkcs12 -export -out "${CERT_PATH}.pfx" -inkey "${CERT_PATH}.key" -in "${CERT_PATH}.crt" -passout "pass:$CERT_PASSWORD"

# Clean up intermediate files
rm "${CERT_PATH}.key"

echo "Certificate successfully created: ${CERT_PATH}.pfx"
