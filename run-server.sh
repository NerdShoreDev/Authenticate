#!/bin/sh
environment=${ENVIRONMENT:-"local"}

echo "Adjust .env file for ${environment}"
echo "==================="
cp ./.env.${environment} ./.env &&
rm ./.env.*

echo "Start server"
echo "==================="
./main
