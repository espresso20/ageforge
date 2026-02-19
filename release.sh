#!/bin/bash

# Create a release package for AgeForge

VERSION="2.0.0"
PACKAGE_NAME="ageforge-${VERSION}"

echo "Creating release package for AgeForge v${VERSION}..."

# Create release directory
mkdir -p "releases/${PACKAGE_NAME}"

# Build for different platforms
echo "Building for multiple platforms..."
GOOS=darwin GOARCH=amd64 go build -o "releases/${PACKAGE_NAME}/ageforge-darwin-amd64" main.go
GOOS=darwin GOARCH=arm64 go build -o "releases/${PACKAGE_NAME}/ageforge-darwin-arm64" main.go
GOOS=linux GOARCH=amd64 go build -o "releases/${PACKAGE_NAME}/ageforge-linux-amd64" main.go
GOOS=windows GOARCH=amd64 go build -o "releases/${PACKAGE_NAME}/ageforge-windows-amd64.exe" main.go

# Copy README and other files
cp README.md "releases/${PACKAGE_NAME}/"
mkdir -p "releases/${PACKAGE_NAME}/data/saves"

# Create zip archives
echo "Creating zip archives..."
cd releases
zip -r "${PACKAGE_NAME}-macos.zip" "${PACKAGE_NAME}/ageforge-darwin-amd64" "${PACKAGE_NAME}/ageforge-darwin-arm64" "${PACKAGE_NAME}/README.md" "${PACKAGE_NAME}/data"
zip -r "${PACKAGE_NAME}-linux.zip" "${PACKAGE_NAME}/ageforge-linux-amd64" "${PACKAGE_NAME}/README.md" "${PACKAGE_NAME}/data"
zip -r "${PACKAGE_NAME}-windows.zip" "${PACKAGE_NAME}/ageforge-windows-amd64.exe" "${PACKAGE_NAME}/README.md" "${PACKAGE_NAME}/data"

echo "Release packages created in releases/ directory!"
