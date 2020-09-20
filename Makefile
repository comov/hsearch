.PHONY: mod run

mod:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

migrate:
	go run cmd/hsearch/*.go migrate

run:
	docker-compose -f local.yml up -d postgres
	go run cmd/hsearch/*.go

dockerbuild:
	docker build -t comov/hsearch:latest .

dockerrun: dockerbuild
	docker run -d hsearch
