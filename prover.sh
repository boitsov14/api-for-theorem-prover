#!/bin/ash

ulimit -St "$3"

java -jar -Xmx"$2" ./prover.jar out "$1"
