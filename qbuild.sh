#!/bin/bash
clear

# Clean up old targets, if any
echo "- cleaning up old targets"
rm -rf macos 2> /dev/null
rm nettest-macos.tar.gz 2> /dev/null

# Compile for target platform
echo "- compiling and building"
GOOS=darwin GOARCH=amd64 go build -o macos/nettest *.go

# Wrap up
cp config.yaml macos/
echo
echo "All done."

