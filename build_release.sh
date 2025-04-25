#!/usr/bin/env bash

echo "Building pulsar-tools"
env GOOS=linux GOARCH=amd64 go build -o pulsar-tools

echo "Building pulsar-tools.exe"
env GOOS=windows GOARCH=amd64 go build -o pulsar-tools.exe

echo "Building pulsar-tools-osx"
env GOOS=darwin GOARCH=amd64 go build -o pulsar-tools-osx
