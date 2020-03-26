FROM golang:1.14.1-alpine3.11 AS builder

RUN apk --no-cache add git gcc g++
COPY . /srv

RUN cd /srv/ && go build -o /go/bin/hsearch cmd/hsearch/*.go


FROM alpine:3.11.5
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/ /usr/local/bin/
COPY --from=builder /srv/migrations /srv/migrations

CMD ["hsearch"]
