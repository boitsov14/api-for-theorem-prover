#!/bin/ash

set -e

go build -o server -ldflags="-s -w" -trimpath
cd ..
go-echo/server
