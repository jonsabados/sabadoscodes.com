#!/bin/bash

UI_BUCKET=$(aws ssm get-parameter --output json --name sabadoscodes.uibucket | jq .Parameter.Value -r)

WORKSPACE=$(cd ../infrastructure && terraform workspace show)

if [ "$WORKSPACE" != 'default' ]; then
  UI_BUCKET="$WORKSPACE-$UI_BUCKET"
fi

echo "Syncing dist/ to ${UI_BUCKET}"

echo "Upload new"
# put everything in the bucket with a max age of 1 year leaving old stuff lying around, excluding index
aws s3 sync ./dist "s3://$UI_BUCKET" --exclude index.html --cache-control max-age=31536000 --acl public-read
echo "Upload new index"
# put the new index in max age on index.html to 60 seconds.
aws s3 cp ./dist/index.html "s3://$UI_BUCKET/index.html" --metadata-directive REPLACE  --cache-control max-age=60 --content-type text/html --acl public-read
echo "Cleaning old"
# now resync everything but index.html with --delete to purge old stuff
aws s3 sync ./dist "s3://$UI_BUCKET" --exclude index.html --cache-control max-age=31536000 --delete --acl public-read
