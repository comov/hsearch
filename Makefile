.PHONY: mod run

mod:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

migrate:
	go run cmd/hsearch/*.go migrate

run:
	go run cmd/hsearch/*.go

dockerbuild:
	docker build -t comov/hsearch:latest .

dockerrun: dockerbuild
	docker run -d hsearch
