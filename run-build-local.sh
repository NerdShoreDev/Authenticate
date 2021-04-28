#!/bin/sh
docker build \
  -t yep.backend \
  --build-arg RDS_CA_FILE_PATH=https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=1735689600000 \
  -f Backend.Dockerfile .