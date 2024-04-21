#!/bin/sh

jq -n --rawfile accounts <(for i in `seq 10`; do cat geth-$i.public; done) '$accounts | split("\n") | map(select(. != ""))' | jq -f genesis.json.tmpl > genesis.json

geth init genesis.json
