FROM golang:1.13.1-alpine3.10 AS builder

RUN apk --no-cache add git gcc g++
COPY . /srv

RUN set -x \
    && cd /srv/ \
    && go build -ldflags="-X 'main.Release=${RELEASE}'" -mod vendor -o /go/bin/hsearch cmd/hsearch/*.go \
    && go build -ldflags="-X 'main.Release=${RELEASE}'" -mod vendor -o /go/bin/newsletter cmd/newsletter/*.go


FROM alpine:3.11.5
RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/ /usr/local/bin/

CMD ["hsearch"]
