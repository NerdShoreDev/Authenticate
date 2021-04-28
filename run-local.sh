#!/bin/sh
environment=${ENVIRONMENT:-"local"}

echo "Adjust .env file"
echo "==================="
cp .env.${environment} .env

echo "Start server"
echo "==================="
go run main.go
