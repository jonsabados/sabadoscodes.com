#!/bin/bash

GOOGLE_ANALYTICS_ID=$(aws ssm get-parameter --output json --name sabadoscodes.googleanalytics.id | jq .Parameter.Value -r)

echo "VUE_APP_GOOGLE_ANALYTICS_ID=${GOOGLE_ANALYTICS_ID}" > .env.local
