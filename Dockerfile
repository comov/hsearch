FROM golang:1.13.1-alpine3.10 AS builder

RUN apk --no-cache add git gcc g++
COPY . /srv

RUN go build -mod vendor -o /go/bin/hsearch /srv/cmd/hsearch/*.go


FROM alpine:3.10.2
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/ /usr/local/bin/
COPY --from=builder /srv/migrations /srv/migrations

CMD ["hsearch"]
