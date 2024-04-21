#!/bin/sh

password_file=geth.password

for i in `seq 10`
do
    prvkey_file=geth-$i.private
    pubkey_file=geth-$i.public

    # generate a private key
    printf "%064x" $i > $prvkey_file

    # import the private key and save the public key
    geth account import --password $password_file $prvkey_file | sed 's/^.*{\(.*\)}.*$/\1/' > $pubkey_file
done
