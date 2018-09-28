FROM golang:alpine AS builder

ENV GOBIN /go/bin

RUN apk add --no-cache git
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/github.com/sweetspotgg/

COPY ./cmd/ ./cmd/
COPY ./vendor/ ./vendor/

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /go/bin/chat ./cmd/chat/

FROM scratch
WORKDIR /app
COPY --from=builder /go/bin /app/
COPY ./web /app/web/

ENTRYPOINT ["/app/chat"]

EXPOSE 8080