.PHONY: mod run

mod:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

update:
	go get -u ./...

migrate:
	go run cmd/hsearch/*.go migrate

run:
	docker-compose -f local.yml up -d postgres
	go run cmd/hsearch/*.go

dockerbuild:
	RELEASE=development docker build -t comov/hsearch:latest .

dockerrun: dockerbuild
	docker run -d comov/hsearch:latest
