#!/bin/sh

geth --password geth.password \
     --unlock 0 --allow-insecure-unlock \
     --mine --miner.etherbase 0x`cat geth-1.public` \
     --http --http.addr 0.0.0.0 --http.vhosts '*' --http.api eth,debug \
     --nodiscover
