#!/bin/ash

go build -o server -ldflags="-s -w" -trimpath
cd ..
go-echo/server
