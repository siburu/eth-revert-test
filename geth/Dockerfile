FROM ethereum/client-go:alltools-v1.13.15

WORKDIR /root

COPY geth.sh .
COPY geth.password .
COPY genesis.json.tmpl .
COPY init-accounts.sh .
COPY init-chain.sh .

RUN apk add jq
RUN ./init-accounts.sh
RUN ./init-chain.sh

EXPOSE 8545/tcp

ENTRYPOINT ["./geth.sh"]