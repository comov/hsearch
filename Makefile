#!make
include .env
export $(shell sed 's/=.*//' .env)

.PHONY: mod run

mod:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor

update:
	go get -u ./...

run:
	docker-compose -f local.yml up -d postgres
	go run cmd/hsearch/*.go

dump:
	pg_dump -U hsearch -W -x -F t hsearch -p ${DJANGO_DB_PORT} -h ${DJANGO_DB_HOST} > backup.tar

restore:
	pg_restore --no-owner --if-exists -c -d hsearch -F t -W -h localhost -p 65432 -U hsearch backup.tar

dockerbuild:
	RELEASE=development docker build -t comov/hsearch:latest .

dockerrun: dockerbuild
	docker run -d comov/hsearch:latest
