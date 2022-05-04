-include .env
export $(shell sed 's/=.*//' .env)

lint:
	golangci-lint run

test:
	go test ./...

mempool:
	cd cmd/mempool && go run . -c ../../build/dipdup.mainnet.yml

local:
	docker-compose -f docker-compose.yml up -d --build
