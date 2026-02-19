#!/bin/bash

# Build script for AgeForge

echo "Building AgeForge..."

# Build for current platform
go build -o ageforge

if [ $? -eq 0 ]; then
    echo "Build successful! Run ./ageforge to start the game."
else
    echo "Build failed."
    exit 1
fi
