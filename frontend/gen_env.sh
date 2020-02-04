#!/bin/bash

GOOGLE_ANALYTICS_ID=$(aws ssm get-parameter --output json --name sabadoscodes.googleanalytics.id | jq .Parameter.Value -r)
GOOGLE_OAUTH_CLIENT_ID=$(aws ssm get-parameter --output json --name sabadoscodes.google.oauth_client_id | jq .Parameter.Value -r)

echo "VUE_APP_GOOGLE_ANALYTICS_ID=${GOOGLE_ANALYTICS_ID}" > .env.local
echo "VUE_APP_GOOGLE_OAUTH_CLIENT_ID=${GOOGLE_OAUTH_CLIENT_ID}" >> .env.local
