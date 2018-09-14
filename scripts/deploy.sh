#!/bin/bash
# Deploy script used for deploying on CI
set -e
main() {
  if [ -z "$DEPLOY_VERSION" ]; then
    echo "---------- No tag to deploy ---------- "
    return
  fi
  echo "---------- Deploying $DEPLOY_VERSION ----------"
  if [ -z "$FORCE_DEPLOY" ]; then
    tkrelease -k="$AWS_KEY" -s="$AWS_SECRET" $DEPLOY_VERSION
  else
    tkrelease -f -k="$AWS_KEY" -s="$AWS_SECRET" $DEPLOY_VERSION
  fi
}
main "$@"
