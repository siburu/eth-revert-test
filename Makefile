.PHONY: build up down deploy abigen

build:
	docker compose build

up:
	docker compose up -d geth

down:
	docker compose down

deploy:
	docker compose run --rm -v $$PWD/share:/mnt hardhat sh -c 'npx hardhat --network geth deploy && mv address.txt artifacts /mnt'

abigen:
	docker compose run --rm -v $$PWD/share:/mnt geth sh -c 'jq .abi < /mnt/artifacts/contracts/test.sol/C.json > C.abi && abigen --abi C.abi --pkg hoge --out hoge.go && mv hoge.go /mnt'
	mkdir -p ./go/hoge
	mv ./share/hoge.go ./go/hoge
