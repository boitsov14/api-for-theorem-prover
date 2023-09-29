#!/bin/ash

set -e

rm /work/*
cp go.mod go.sum ./*.go ../.env ../prover ../prover.sh /work
cd /work
go test -v
