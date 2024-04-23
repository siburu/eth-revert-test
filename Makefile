.PHONY: build up down deploy abigen clean

build:
	docker compose build

up:
	docker compose up -d geth

down:
	docker compose down

deploy:
	docker compose run --rm -v $$PWD/share:/mnt hardhat sh -c 'npx hardhat --network geth deploy && mv address.txt artifacts /mnt'

abigen:
	docker compose run --rm -v $$PWD/share:/mnt geth sh -c 'jq .abi < /mnt/artifacts/contracts/test.sol/C.json > C.abi && abigen --abi C.abi --pkg contracts --type C --out c.go && mv c.go /mnt'
	mkdir -p ./go/contracts
	mv ./share/c.go ./go/contracts

clean:
	-rm -rf share
	-rm -rf go/contracts
