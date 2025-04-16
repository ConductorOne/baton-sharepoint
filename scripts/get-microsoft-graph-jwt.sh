#!/usr/bin/env bash

set -euxo pipefail

check_var() {
    local var_name="$1"
    local var_value="${!var_name}"
    if [ -z "$var_value" ]; then
        echo "Error: Environment variable '$var_name' is not set or is empty." >&2
        exit 1
    fi
}

check_var "BATON_AZURE_CLIENT_ID"
check_var "BATON_AZURE_CLIENT_SECRET"
check_var "BATON_AZURE_TENANT_ID"

curl -X POST \
     -d "client_id=$BATON_AZURE_CLIENT_ID" \
     -d "scope=https://${BATON_AZURE_GRAPH_DOMAIN:-graph.microsoft.com}/.default" \
     -d "client_secret=$BATON_AZURE_CLIENT_SECRET" \
     -d "grant_type=client_credentials" \
     https://login.microsoftonline.com/$BATON_AZURE_TENANT_ID/oauth2/v2.0/token
