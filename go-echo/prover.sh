#!/bin/ash

ulimit -St "$3"

./prover -Xmx"$2" out "$1"
