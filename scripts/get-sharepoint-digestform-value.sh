#!/usr/bin/env bash

set -euo pipefail

# Check for required arguments
if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <sharepoint_site_url> <bearer_token>" >&2
  exit 1
fi

site_url="$1"
bearer_token="$2"

# Validate that site_url is not empty
if [[ -z "$site_url" ]]; then
  echo "Error: SharePoint site URL is empty." >&2
  exit 1
fi

# Validate that bearer_token is not empty
if [[ -z "$bearer_token" ]]; then
  echo "Error: Bearer token is empty." >&2
  exit 1
fi

# Make the request to get the FormDigestValue
curl -s -X POST \
  "${site_url}/_api/contextinfo" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer ${bearer_token}" \
  --data ""
