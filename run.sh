#!/bin/zsh
echo "Building AgeForge..."
go build -o ageforge . && ./ageforge
