-include .env
export $(shell sed 's/=.*//' .env)

lint:
	golangci-lint run

test:
	go test ./...

integration-test:
	#docker volume prune -f
	docker-compose -f docker-compose.test.yml up -d --build
	sleep 15
	cd cmd/mempool && INTEGRATION=true HASURA_HOST=127.0.0.1 HASURA_PORT=22000 bash -c 'go test -v -timeout=15s -run TestIntegration_HasuraMetadata' || true
	docker-compose -f docker-compose.test.yml down -v

mempool:
	cd cmd/mempool && go run . -c ../../build/dipdup.testnet.yml

local:
	docker-compose -f docker-compose.yml up -d --build

local-test:
	docker-compose -f docker-compose.test.yml up -d --build
