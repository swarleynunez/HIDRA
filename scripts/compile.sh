#!/usr/bin/env bash

# Compile HIDRA
# env GOOS=windows GOARCH=amd64 go build -o bin/hidra.exe main.go
env GOOS=linux GOARCH=amd64 go build -o bin/hidra main.go
# env GOOS=darwin GOARCH=amd64 go build -o bin/hidra main.go
